package analyzer

import (
	"math"

	"github.com/cortze/eth2-state-analyzer/pkg/custom_spec"
	"github.com/pkg/errors"

	api "github.com/attestantio/go-eth2-client/api/v1"
	"github.com/attestantio/go-eth2-client/spec/phase0"
)

const (
	baseRewardFactor    = 64.0
	baseRewardsPerEpoch = 4.0
	// participationRate   = 0.945 // about to calculate participation rate
)

func GetValidatorBalance(customBState custom_spec.CustomBeaconState, valIdx uint64) (uint64, error) {

	balance, err := customBState.Balance(valIdx)

	if err != nil {
		return 0, err
	}

	return balance, nil
}

// func GetParticipationRate(customBState CustomBeaconState, s *StateAnalyzer, m map[string]bitfield.Bitlist) (uint64, error) {

// 	// participationRate := 0.85

// 	currentSlot := customBState.CurrentSlot()
// 	currentEpoch := customBState.CurrentEpoch()
// 	totalAttPreviousEpoch := customBState.PreviousEpochAttestations()
// 	totalAttestingVals := customBState.PreviousEpochValNum()

// 	// TODO: for now we print it but the goal is to store in a DB
// 	fmt.Println("Current Epoch: ", currentEpoch)
// 	fmt.Println("Using Block at: ", currentSlot)
// 	fmt.Println("Attestations in the last Epoch: ", totalAttPreviousEpoch)
// 	fmt.Println("Total number of Validators: ", totalAttestingVals)

// 	return 0, nil
// }

// https://kb.beaconcha.in/rewards-and-penalties
// https://consensys.net/blog/codefi/rewards-and-penalties-on-ethereum-20-phase-0/
// TODO: -would be nice to incorporate top the max value wheather there were 2-3 consecutive missed blocks afterwards
func GetMaxReward(valIdx uint64, totValStatus *map[phase0.ValidatorIndex]*api.Validator, totalEffectiveBalance uint64, participationRate float64) (uint64, error) {
	// First iteration just taking 31/8*BaseReward as Max value
	// BaseReward = ( effectiveBalance * (BaseRewardFactor)/(BaseRewardsPerEpoch * sqrt(activeBalance)) )

	idx := phase0.ValidatorIndex(valIdx)

	valStatus, ok := (*totValStatus)[idx]
	if !ok {
		return 0, errors.New("")
	}
	// apply formula
	baseReward := GetBaseReward(valStatus.Validator.EffectiveBalance, totalEffectiveBalance)
	voteReward := 3.0 * baseReward * participationRate
	inclusionDelay := baseReward * 7.0 / 8.0

	maxReward := voteReward + inclusionDelay

	// maxReward := ((31.0 / 8.0) * participationRate * (float64(uint64(valStatus.Validator.EffectiveBalance)) * baseRewardFactor))
	// maxReward = maxReward / (baseRewardsPerEpoch * math.Sqrt(float64(totalActiveBalance)))

	return uint64(maxReward), nil
}

// directly calculated on the MaxReward fucntion
func GetBaseReward(valEffectiveBalance phase0.Gwei, totalEffectiveBalance uint64) float64 {
	// BaseReward = ( effectiveBalance * (BaseRewardFactor)/(BaseRewardsPerEpoch * sqrt(activeBalance)) )
	var baseReward float64

	sqrt := uint64(math.Sqrt(float64(totalEffectiveBalance)))

	denom := float64(baseRewardsPerEpoch * sqrt)

	num := (float64(valEffectiveBalance) * baseRewardFactor)
	baseReward = num / denom

	return baseReward
}
