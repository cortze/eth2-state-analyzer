package custom_spec

import (
	"fmt"

	"github.com/attestantio/go-eth2-client/http"
	"github.com/attestantio/go-eth2-client/spec"
	"github.com/attestantio/go-eth2-client/spec/phase0"
	"github.com/sirupsen/logrus"
)

var (
	log = logrus.WithField(
		"module", "custom_spec",
	)
)

type Phase0Spec struct {
	BState                        spec.VersionedBeaconState
	PrevBState                    spec.VersionedBeaconState
	PreviousEpochAttestingVals    []uint64
	PreviousEpochAttestingNum     uint64
	PreviousEpochAttestingBalance uint64
	ValAttestationInclusion       map[uint64]ValVote
	TotalActiveVals               uint64
	Api                           *http.Service
	PrevEpochStructs              EpochData
	EpochStructs                  EpochData
	AttestedValsPerSlot           map[uint64][]uint64
}

func NewPhase0Spec(bstate *spec.VersionedBeaconState, prevBstate spec.VersionedBeaconState, iApi *http.Service) Phase0Spec {
	// func NewPhase0Spec(bstate *spec.VersionedBeaconState, iCli *clientapi.APIClient) Phase0Spec {

	if prevBstate.Phase0 == nil {
		prevBstate = *bstate
	}

	phase0Obj := Phase0Spec{
		BState:                        *bstate,
		PrevBState:                    prevBstate,
		PreviousEpochAttestingVals:    make([]uint64, len(prevBstate.Phase0.Validators)),
		PreviousEpochAttestingNum:     0,
		PreviousEpochAttestingBalance: 0,
		ValAttestationInclusion:       make(map[uint64]ValVote),
		TotalActiveVals:               0,
		Api:                           iApi,
		PrevEpochStructs:              NewEpochData(iApi, prevBstate.Phase0.Slot),
		EpochStructs:                  NewEpochData(iApi, bstate.Phase0.Slot),
		AttestedValsPerSlot:           make(map[uint64][]uint64),
		// the maximum inclusionDelay is 32, and we are counting aggregations from the current Epoch
	}

	var attestations []*phase0.PendingAttestation
	if len(bstate.Phase0.PreviousEpochAttestations) > 0 {
		// we are not in genesis
		attestations = bstate.Phase0.PreviousEpochAttestations
	} else {
		// we are in genesis
		attestations = bstate.Phase0.CurrentEpochAttestations
	}

	phase0Obj.PreviousEpochAttestingVals = phase0Obj.CalculateAttestingVals(attestations, uint64(len(prevBstate.Phase0.Validators)))
	phase0Obj.PreviousEpochAttestingBalance = phase0Obj.ValsBalance(phase0Obj.PreviousEpochAttestingVals)

	phase0Obj.CalculateCurrentEpochAggegations()
	return phase0Obj

}

func (p Phase0Spec) CurrentSlot() uint64 {
	return p.BState.Phase0.Slot
}

func (p Phase0Spec) CurrentEpoch() uint64 {
	return uint64(p.CurrentSlot() / 32)
}

func (p Phase0Spec) PrevStateSlot() uint64 {
	return p.PrevBState.Phase0.Slot
}

func (p Phase0Spec) PrevStateEpoch() uint64 {
	return uint64(p.PrevStateSlot() / 32)
}

func (p *Phase0Spec) CalculateAttestingVals(attestations []*phase0.PendingAttestation, valNum uint64) []uint64 {

	resultAttVals := make([]uint64, valNum)

	for _, item := range attestations {

		slot := item.Data.Slot            // Block that is being attested, not included
		committeeIndex := item.Data.Index // committee in the attested slot

		validatorIDs := p.PrevEpochStructs.GetValList(uint64(slot), uint64(committeeIndex))

		attestingIndices := item.AggregationBits.BitIndices()

		for _, index := range attestingIndices {
			attestingValIdx := validatorIDs[index]

			resultAttVals[attestingValIdx] = resultAttVals[attestingValIdx] + 1
		}
	}

	return resultAttVals
}
func (p *Phase0Spec) CalculateCurrentEpochAttestingVals() {

	currentAttVals := make([]uint64, len(p.BState.Phase0.Validators))
	attestingBalance := uint64(0)
	for _, item := range p.BState.Phase0.CurrentEpochAttestations {

		slot := item.Data.Slot            // Block that is being attested, not included
		committeeIndex := item.Data.Index // committee in the attested slot

		validatorIDs := p.EpochStructs.GetValList(uint64(slot), uint64(committeeIndex))

		attestingIndices := item.AggregationBits.BitIndices()

		for _, index := range attestingIndices {
			attestingValIdx := validatorIDs[index]
			if currentAttVals[attestingValIdx] == 0 {
				attestingBalance += uint64(p.BState.Phase0.Validators[attestingValIdx].EffectiveBalance)
			}
			currentAttVals[attestingValIdx] = currentAttVals[attestingValIdx] + 1
		}
	}

	fmt.Println(attestingBalance)
}

func (p Phase0Spec) PreviousEpochAttestations() uint64 {

	return p.PreviousEpochAttestingNum
}

// the length of the valList = number of validators
// each position represents a valIdx
// if the item has a number > 0, count it
func (p Phase0Spec) ValsBalance(valList []uint64) uint64 {

	attestingBalance := uint64(0)

	for valIdx, numAtt := range valList { // loop over validators
		if numAtt > 0 {
			attestingBalance += uint64(p.PrevBState.Phase0.Validators[valIdx].EffectiveBalance)
		}
	}

	return uint64(attestingBalance)
}

func (p Phase0Spec) PreviousEpochAggregations() []*phase0.PendingAttestation {
	return p.BState.Phase0.PreviousEpochAttestations
}

func (p Phase0Spec) PreviousEpochValNum() uint64 {

	return p.TotalActiveVals
}

func (p Phase0Spec) PreviousEpochActiveValBalance() uint64 {

	activeEffectiveBalance := uint64(0)
	vals := p.PrevBState.Phase0.Validators

	for _, item := range vals {
		// validator must be either active, exiting or slashed
		if item.ActivationEligibilityEpoch < phase0.Epoch(p.CurrentEpoch()) &&
			item.ExitEpoch > phase0.Epoch(p.CurrentEpoch()) {
			activeEffectiveBalance += uint64(item.EffectiveBalance)
		}
	}
	return uint64(activeEffectiveBalance)
}

func (p Phase0Spec) CurrentEpochActiveBalance() uint64 {

	activeEffectiveBalance := 0
	vals := p.BState.Phase0.Validators

	if p.CurrentSlot() < 32 {
		// genesis epoch, validators preactivated
		return uint64(len(vals) * EFFECTIVE_BALANCE_INCREMENT * MAX_EFFECTIVE_INCREMENTS)
	}

	for _, item := range vals {
		// validator must be either active, exiting or slashed
		if item.ActivationEligibilityEpoch < phase0.Epoch(p.CurrentEpoch()) &&
			item.ExitEpoch > phase0.Epoch(p.CurrentEpoch()) {
			activeEffectiveBalance += int(item.EffectiveBalance)
		}
	}
	return uint64(activeEffectiveBalance)
}

func (p Phase0Spec) CalculateCurrentEpochAggegations() {

	attestations := append(p.BState.Phase0.CurrentEpochAttestations, p.BState.Phase0.PreviousEpochAttestations...)
	attestingVals := make([]uint64, len(p.BState.Phase0.Validators))
	// we need to take into account also previous epoch attestations that were included in this epoch

	for _, item := range attestations {

		slot := item.Data.Slot            // Block that is being attested, not included
		committeeIndex := item.Data.Index // committee in the attested slot
		inclusionSlot := slot + item.InclusionDelay

		attestingIndices := item.AggregationBits.BitIndices()

		committee := p.PrevEpochStructs.GetValList(uint64(slot), uint64(committeeIndex))

		if committee == nil {
			committee = p.EpochStructs.GetValList(uint64(slot), uint64(committeeIndex))
		}

		// loop over the vals that attested
		for _, index := range attestingIndices {
			valID := committee[index]
			if uint64(inclusionSlot) >= (p.BState.Phase0.Slot - (EPOCH_SLOTS - 1)) {
				attestingVals[valID] = attestingVals[valID] + 1
			}

			if val, ok := p.ValAttestationInclusion[uint64(valID)]; ok {
				// it already existed
				val.AddNewAtt(uint64(slot), uint64(inclusionSlot))
				p.ValAttestationInclusion[uint64(valID)] = val
			} else {

				// it did not exist
				newAtt := ValVote{
					ValId: uint64(valID),
				}
				newAtt.AddNewAtt(uint64(slot), uint64(inclusionSlot))
				p.ValAttestationInclusion[uint64(valID)] = newAtt

			}
		}
	}

	attestingBalance := uint64(0)
	for valIdx, numAtt := range attestingVals {
		if numAtt > 0 {
			attestingBalance += uint64(p.BState.Phase0.Validators[valIdx].EffectiveBalance)
		}

	}

	fmt.Println(attestingBalance)

	// at this point we have a map where the key is a valID
	// each entry contains an array of attestedSlot and minimum inclusionSlot

}

func (p Phase0Spec) Balance(valIdx uint64) (uint64, error) {
	if uint64(len(p.BState.Phase0.Balances)) < valIdx {
		err := fmt.Errorf("phase0 - validator index %d wasn't activated in slot %d", valIdx, p.BState.Phase0.Slot)
		return 0, err
	}
	balance := p.BState.Phase0.Balances[valIdx]

	return balance, nil
}

func (p Phase0Spec) PrevEpochReward(valIdx uint64) uint64 {
	return p.BState.Phase0.Balances[valIdx] - p.PrevBState.Phase0.Balances[valIdx]
}

func (p Phase0Spec) GetMaxProposerReward(valIdx uint64, valEffectiveBalance uint64, totalEffectiveBalance uint64) float64 {

	isProposer := false
	proposerSlot := 0
	for _, duty := range p.PrevEpochStructs.ProposerDuties {
		if duty.ValidatorIndex == phase0.ValidatorIndex(valIdx) {
			isProposer = true
			proposerSlot = int(duty.Slot)
			break
		}
	}

	if isProposer {
		votesIncluded := 0
		for _, valAttestation := range p.ValAttestationInclusion {
			for _, item := range valAttestation.InclusionSlot {
				if item == uint64(proposerSlot) {
					// the block the attestation was included is the same as the slot the val proposed a block
					// therefore, proposer included this attestation
					votesIncluded += 1
				}
			}
		}

		baseReward := GetBaseReward(valEffectiveBalance, totalEffectiveBalance)
		return (baseReward / PROPOSER_REWARD_QUOTIENT) * float64(votesIncluded)
	}

	return 0
}

func (p Phase0Spec) GetMaxReward(valIdx uint64) (uint64, error) {

	valEffectiveBalance := p.BState.Phase0.Validators[valIdx].EffectiveBalance
	previousAttestedBalance := p.ValsBalance(p.PreviousEpochAttestingVals)

	currentActiveBalance := p.CurrentEpochActiveBalance()

	participationRate := float64(previousAttestedBalance) / float64(currentActiveBalance)

	// First iteration just taking 31/8*BaseReward as Max value
	// BaseReward = ( effectiveBalance * (BaseRewardFactor)/(BaseRewardsPerEpoch * sqrt(activeBalance)) )

	// apply formula
	baseReward := GetBaseReward(uint64(valEffectiveBalance), currentActiveBalance)

	voteReward := 3.0 * baseReward * participationRate
	inclusionDelayReward := baseReward * 7.0 / 8.0

	proposerReward := p.GetMaxProposerReward(valIdx, uint64(valEffectiveBalance), currentActiveBalance)

	maxReward := voteReward + inclusionDelayReward + proposerReward

	return uint64(maxReward), nil
}

func (p Phase0Spec) GetAttestingSlot(valIdx uint64) uint64 {
	return 0
}
