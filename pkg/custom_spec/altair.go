package custom_spec

import (
	"fmt"
	"math"

	"github.com/attestantio/go-eth2-client/http"
	"github.com/attestantio/go-eth2-client/spec"
	"github.com/attestantio/go-eth2-client/spec/altair"
	"github.com/attestantio/go-eth2-client/spec/phase0"
	"github.com/cortze/eth2-state-analyzer/pkg/utils"
	"github.com/prysmaticlabs/go-bitfield"
)

type AltairSpec struct {
	BState        spec.VersionedBeaconState
	Committees    map[string]bitfield.Bitlist
	DoubleVotes   uint64
	Api           *http.Service
	EpochStructs  EpochData
	AttestingVals []uint64
}

func NewAltairSpec(bstate *spec.VersionedBeaconState, iApi *http.Service) AltairSpec {
	altairObj := AltairSpec{
		BState:        *bstate,
		Committees:    make(map[string]bitfield.Bitlist),
		DoubleVotes:   0,
		Api:           iApi,
		EpochStructs:  NewEpochData(iApi, bstate.Altair.Slot),
		AttestingVals: make([]uint64, 0),
	}
	altairObj.CalculatePreviousAttestingVals()

	return altairObj
}

func (p AltairSpec) CurrentSlot() uint64 {
	return p.BState.Altair.Slot
}

func (p AltairSpec) CurrentEpoch() uint64 {
	return uint64(p.CurrentSlot() / 32)
}

func (p AltairSpec) PrevStateSlot() uint64 {
	// return p.PrevBState.Phase0.Slot
	return 0
}

func (p AltairSpec) PrevStateEpoch() uint64 {
	return uint64(p.PrevStateSlot() / 32)
}

func (p *AltairSpec) CalculatePreviousAttestingVals() {
	flag := altair.ParticipationFlags(math.Pow(2, float64(altair.TimelySourceFlagIndex)))

	for valIndex, item := range p.BState.Altair.PreviousEpochParticipation {
		// Here we have one item per validator
		// This is an 8-bit string, where the bit 0 is the source
		// If it is set, we consider there was a vote from this validator
		// if utils.IsBitSet(uint8(item), 0) {
		// }

		if (item & flag) == flag {
			// The attestation has a timely source value, therefore we consider it attest
			p.AttestingVals = append(p.AttestingVals, uint64(valIndex))
		}
	}
}

func (p AltairSpec) PreviousEpochAttestations() uint64 {

	return uint64(len(p.AttestingVals))

}

func (p AltairSpec) GetValidatorBalance(valIdx uint64) uint64 {
	return p.BState.Altair.Balances[valIdx]

}

func (p AltairSpec) PreviousEpochMissedAttestations() uint64 {
	numOfMissedAttestations := 0
	attestationsPerVal := p.BState.Altair.PreviousEpochParticipation

	for _, item := range attestationsPerVal {
		// Here we have one item per validator
		if !utils.IsBitSet(uint8(item), 0) {
			numOfMissedAttestations++
		}
	}

	return uint64(numOfMissedAttestations)

}

func (p AltairSpec) PreviousEpochValNum() uint64 {
	vals := p.BState.Altair.Validators
	totalAttestingVals := 0
	for _, item := range vals {
		// validator must be either active, exiting or slashed
		if item.ActivationEligibilityEpoch < phase0.Epoch(p.CurrentEpoch()) &&
			item.ExitEpoch > phase0.Epoch(p.CurrentEpoch()) {
			totalAttestingVals += 1
		}
	}
	return uint64(totalAttestingVals)
}

func (p AltairSpec) GetDoubleVotes() uint64 {
	return p.DoubleVotes
}

func (p AltairSpec) Balance(valIdx uint64) (uint64, error) {
	if uint64(len(p.BState.Altair.Balances)) < valIdx {
		err := fmt.Errorf("phase0 - validator index %d wasn't activated in slot %d", valIdx, p.BState.Altair.Slot)
		return 0, err
	}
	balance := p.BState.Altair.Balances[valIdx]

	return balance, nil
}

func (p AltairSpec) GetMaxProposerAttReward(valIdx uint64, valPubKey phase0.BLSPubKey, valEffectiveBalance uint64, totalEffectiveBalance uint64) float64 {

	return 0

}

func (p AltairSpec) GetMaxSyncComReward(valIdx uint64, valPubKey phase0.BLSPubKey, valEffectiveBalance uint64, totalEffectiveBalance uint64) float64 {

	return 0

}

func (p AltairSpec) GetMaxAttestationReward(valIdx uint64, valEffectiveBalance uint64, totalEffectiveBalance uint64) float64 {

	return 0
}

func (p AltairSpec) GetMaxReward(valIdx uint64) (uint64, error) {

	return 0, nil
}

func (p AltairSpec) GetAttestingSlot(valIdx uint64) uint64 {
	return 0
}

func (p AltairSpec) PrevEpochReward(valIdx uint64) uint64 {
	return 0
}
