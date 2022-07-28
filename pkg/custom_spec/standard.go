package custom_spec

import (
	"context"
	"fmt"
	"math"
	"strconv"

	api "github.com/attestantio/go-eth2-client/api/v1"
	"github.com/attestantio/go-eth2-client/http"
	"github.com/attestantio/go-eth2-client/spec"
	"github.com/attestantio/go-eth2-client/spec/phase0"
	"github.com/cortze/eth2-state-analyzer/pkg/utils"
	"github.com/prysmaticlabs/go-bitfield"
)

const (
	MAX_EFFECTIVE_INCREMENTS    = 32
	BASE_REWARD_FACTOR          = 64.0
	BASE_REWARD_PER_EPOCH       = 4.0
	EFFECTIVE_BALANCE_INCREMENT = 1000000000
	ATTESTATION_FACTOR          = 0.84375 // 14+26+14/64
	PROPOSER_WEIGHT             = 0.125
	SYNC_COMMITTEE_FACTOR       = 0.00000190734
	EPOCH_SLOTS                 = 32
	SHUFFLE_ROUND_COUNT         = uint64(90)
	PROPOSER_REWARD_QUOTIENT    = 8
	// participationRate   = 0.945 // about to calculate participation rate
)

// directly calculated on the MaxReward fucntion
func GetBaseReward(valEffectiveBalance uint64, totalEffectiveBalance uint64) float64 {
	// BaseReward = ( effectiveBalance * (BaseRewardFactor)/(BaseRewardsPerEpoch * sqrt(activeBalance)) )
	var baseReward float64

	sqrt := uint64(math.Sqrt(float64(totalEffectiveBalance)))

	denom := float64(BASE_REWARD_PER_EPOCH * sqrt)

	num := (float64(valEffectiveBalance) * BASE_REWARD_FACTOR)
	baseReward = num / denom

	return baseReward
}

func GetBaseRewardPerInc(totalEffectiveBalance uint64) float64 {
	// BaseReward = ( effectiveBalance * (BaseRewardFactor)/(BaseRewardsPerEpoch * sqrt(activeBalance)) )
	var baseReward float64

	sqrt := uint64(math.Sqrt(float64(totalEffectiveBalance)))

	num := EFFECTIVE_BALANCE_INCREMENT * BASE_REWARD_FACTOR
	baseReward = num / float64(sqrt)

	return baseReward
}

type CustomBeaconState interface {
	PreviousEpochAttestations() uint64
	PreviousEpochValNum() uint64 // those activated before current Epoch
	CurrentEpoch() uint64
	CurrentSlot() uint64
	PrevStateEpoch() uint64
	PrevStateSlot() uint64
	Balance(valIdx uint64) (uint64, error)
	GetAttestingSlot(valIdx uint64) uint64
	GetMaxReward(valIdx uint64) (uint64, error)
	PrevEpochReward(valIdx uint64) uint64
}

func BStateByForkVersion(bstate *spec.VersionedBeaconState, prevBstate spec.VersionedBeaconState, iApi *http.Service) (CustomBeaconState, error) {
	switch bstate.Version {

	case spec.DataVersionPhase0:
		return NewPhase0Spec(bstate, prevBstate, iApi), nil

	case spec.DataVersionAltair:
		return NewAltairSpec(bstate, iApi), nil

	case spec.DataVersionBellatrix:
		return NewBellatrixSpec(bstate, iApi), nil
	default:
		return nil, fmt.Errorf("could not figure out the Beacon State Fork Version")
	}
}

type CustomAggregation struct {
	AggregationBits bitfield.Bitlist
	ValidatorsIDs   []phase0.ValidatorIndex
}

type ValVote struct {
	ValId         uint64
	AttestedSlot  []uint64
	InclusionSlot []uint64
}

func (p *ValVote) AddNewAtt(attestedSlot uint64, inclusionSlot uint64) {

	if p.AttestedSlot == nil {
		p.AttestedSlot = make([]uint64, 0)
	}

	if p.InclusionSlot == nil {
		p.InclusionSlot = make([]uint64, 0)
	}

	// keep in mind that for the proposer, the vote only counts if it is the first to include this attestation
	for i, item := range p.AttestedSlot {
		if item == attestedSlot {
			if inclusionSlot < p.InclusionSlot[i] {
				p.InclusionSlot[i] = inclusionSlot
			}
			return
		}
	}

	p.AttestedSlot = append(p.AttestedSlot, attestedSlot)
	p.InclusionSlot = append(p.InclusionSlot, inclusionSlot)

}

func (p CustomAggregation) GetAttestingVals() []uint64 {

	attestingVals := make([]uint64, 0)

	indices := p.AggregationBits.BitIndices() // get attesting indices of committee
	for _, index := range indices {
		newAttestingValID := uint64(p.ValidatorsIDs[index])
		attestingVals = append(attestingVals, newAttestingValID)
	}

	return attestingVals
}

type EpochData struct {
	ProposerDuties    []*api.ProposerDuty
	BeaconCommittees  []*api.BeaconCommittee // Beacon Committees organized by slot for the whole epoch
	ValidatorAttSlot  map[uint64]uint64      // for each validator we have which slot it had to attest to
	ValidatorsPerSlot map[uint64][]uint64    // each Slot, which validators had to attest
}

func NewEpochData(iApi *http.Service, slot uint64) EpochData {

	epochCommittees, err := iApi.BeaconCommittees(context.Background(), strconv.Itoa(int(slot)))

	if err != nil {
		log.Errorf(err.Error())
	}

	validatorsAttSlot := make(map[uint64]uint64) // each validator, when it had to attest
	validatorsPerSlot := make(map[uint64][]uint64)

	for _, committee := range epochCommittees {
		for _, valID := range committee.Validators {
			validatorsAttSlot[uint64(valID)] = uint64(committee.Slot)

			if val, ok := validatorsPerSlot[uint64(committee.Slot)]; ok {
				// the slot exists in the map
				validatorsPerSlot[uint64(committee.Slot)] = append(val, uint64(valID))
			} else {
				// the slot does not exist, create
				validatorsPerSlot[uint64(committee.Slot)] = []uint64{uint64(valID)}
			}
		}
	}

	proposerDuties, err := iApi.ProposerDuties(context.Background(), phase0.Epoch(utils.GetEpochFromSlot(uint64(slot))), nil)

	if err != nil {
		log.Errorf(err.Error())
	}

	return EpochData{
		ProposerDuties:    proposerDuties,
		BeaconCommittees:  epochCommittees,
		ValidatorAttSlot:  validatorsAttSlot,
		ValidatorsPerSlot: validatorsPerSlot,
	}
}

func (p EpochData) GetValList(slot uint64, committeeIndex uint64) []phase0.ValidatorIndex {
	for _, committee := range p.BeaconCommittees {
		if (uint64(committee.Slot) == slot) && (uint64(committee.Index) == committeeIndex) {
			return committee.Validators
		}
	}

	return nil
}

func GetEffectiveBalance(balance float64) float64 {
	return math.Min(MAX_EFFECTIVE_INCREMENTS*EFFECTIVE_BALANCE_INCREMENT, balance)
}
