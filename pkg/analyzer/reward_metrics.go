package analyzer

import (
	"errors"
	"fmt"
	"sync"

	api "github.com/attestantio/go-eth2-client/api/v1"
	"github.com/attestantio/go-eth2-client/spec/phase0"
	"github.com/cortze/eth2-state-analyzer/pkg/custom_spec"
	"github.com/cortze/eth2-state-analyzer/pkg/model"

	"github.com/cortze/eth2-state-analyzer/pkg/utils"
)

type RewardMetrics struct {
	m                 sync.Mutex
	baseslot          uint64
	innerCnt          uint64
	validatorIdx      uint64
	ValidatorBalances []uint64  // Gwei ?¿
	MaxRewards        []uint64  // Gwei ?¿
	Rewards           []int64   // Gweis ?¿
	RewardPercentage  []float64 // %
	AttSlot           []uint64  // attesting slot inside epoch

	MissingSources []uint64
	MissingHeads   []uint64
	MissingTargets []uint64
}

func NewRewardMetrics(initslot uint64, epochRange uint64, validatorIdx uint64) (*RewardMetrics, error) {
	log.Debugf("len of the epoch range %d", epochRange)
	return &RewardMetrics{
		baseslot:          initslot,
		validatorIdx:      validatorIdx,
		ValidatorBalances: make([]uint64, epochRange),
		MaxRewards:        make([]uint64, epochRange),
		Rewards:           make([]int64, epochRange),
		RewardPercentage:  make([]float64, epochRange),
		AttSlot:           make([]uint64, epochRange),
		MissingSources:    make([]uint64, epochRange),
		MissingHeads:      make([]uint64, epochRange),
		MissingTargets:    make([]uint64, epochRange),
	}, nil
}

// Supposed to be
func (m *RewardMetrics) CalculateEpochPerformance(customBState custom_spec.CustomBeaconState, validators *map[phase0.ValidatorIndex]*api.Validator, totalEffectiveBalance uint64) error {

	validatorBalance, err := GetValidatorBalance(customBState, m.validatorIdx)
	if err != nil {
		return err
	}
	// add always the balance to the array
	log.Debugf("starting performance calculation of validator %d", m.validatorIdx)
	m.m.Lock()
	defer func() {
		log.Debugf("finishing the performance calculation of validator %d", m.validatorIdx)
		m.m.Unlock()
	}()
	m.ValidatorBalances[m.innerCnt] = uint64(validatorBalance)

	m.AttSlot[m.innerCnt] = customBState.GetAttestingSlot(m.validatorIdx)

	// Proccess Max-Rewards
	maxReward, err := customBState.GetMaxReward(m.validatorIdx)
	if err != nil {
		return err
	}
	log.Debugf("max reward for validator %d = %d", m.validatorIdx, maxReward)
	m.MaxRewards[m.innerCnt] = maxReward

	if m.innerCnt != 0 {
		// calculate Reward from the previous Balance
		previousBalance := m.ValidatorBalances[m.innerCnt-1]
		reward := int64(validatorBalance) - int64(previousBalance)
		if previousBalance > validatorBalance {
			fmt.Println(reward)
		}
		log.Debugf("reward for validator %d = %d", m.validatorIdx, reward)
		// Add Reward
		m.Rewards[m.innerCnt-1] = reward

		// Proccess Reward-Performance-Ratio
		rewardPerf := (float64(reward) * 100) / float64(maxReward)
		m.RewardPercentage[m.innerCnt-1] = rewardPerf
		log.Debugf("reward performance for %d = %f", m.validatorIdx, rewardPerf)

		// TODO:
		// Add number of missing sources

		// Add number of missing head

		// Add number of missing targets

	}

	// do not forget to increment the internal counter
	m.innerCnt++
	log.Debugf("done with validator %d", m.validatorIdx)

	return nil
}

func (m *RewardMetrics) GetEpochMetrics(slot uint64) (model.SingleEpochMetrics, error) {
	var epochMetrics model.SingleEpochMetrics

	// calculate the index
	if slot < 31 {
		slot = 31
	}
	idx := m.GetIndexFromslot(slot)
	if idx < 0 {
		log.Errorf("requested metrics for slot: %d couldn't be found. Max slot is %d", slot, m.baseslot+(32*uint64(len(m.Rewards))))
		return epochMetrics, errors.New("requested slot can not be found on the analyzed data")
	}

	m.m.Lock()
	// if the index is okey, compose the Singe epoch metrics
	epochMetrics.ValidatorIdx = m.validatorIdx
	epochMetrics.Slot = slot
	epochMetrics.Epoch = utils.GetEpochFromSlot(slot)
	epochMetrics.ValidatorBalance = m.ValidatorBalances[idx]
	epochMetrics.MaxReward = m.MaxRewards[idx]
	epochMetrics.Reward = m.Rewards[idx]
	epochMetrics.RewardPercentage = m.RewardPercentage[idx]
	epochMetrics.AttSlot = m.AttSlot[idx]
	epochMetrics.MissingSource = m.MissingSources[idx]
	epochMetrics.MissingHead = m.MissingHeads[idx]
	epochMetrics.MissingTarget = m.MissingTargets[idx]
	epochMetrics.ValidatorIdx = m.validatorIdx
	m.m.Unlock()
	return epochMetrics, nil
}

func (m *RewardMetrics) GetIndexFromslot(slot uint64) int {
	idx := -1
	m.m.Lock()
	defer m.m.Unlock()
	// idx = int(slot/m.baseslot) - 1
	idx = int(slot-m.baseslot) / 32
	if idx >= len(m.Rewards) {
		return -1
	}

	return idx
}

func (m *RewardMetrics) AddslotPerformance() error {

	return nil
}
