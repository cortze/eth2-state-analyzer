package model

// Postgres intregration variables
var (
	CreateValidatorRewardsTable = `
	CREATE TABLE IF NOT EXISTS t_validator_rewards_summary(
		f_val_idx INT,
		f_slot INT,
		f_epoch INT,
		f_balance BIGINT,
		f_reward BIGINT,
		f_max_reward BIGINT,
		f_attesting_slot BIGINT,
		CONSTRAINT PK_ValidatorSlot PRIMARY KEY (f_val_idx,f_slot));`

	InsertNewValidatorLineTable = `
	INSERT INTO t_validator_rewards_summary (f_val_idx, f_slot, f_epoch, f_balance, f_reward, f_max_reward, f_attesting_slot)
	VALUES ($1, $2, $3, $4, $5, $6, $7);
	`

	UpdateValidatorLineTable = `
	UPDATE t_validator_rewards_summary
	SET f_reward=$3
	WHERE f_val_idx=$1 AND f_slot=$2
	`

	SelectByValSlot = `
	SELECT f_val_idx, f_slot, f_epoch, f_balance, f_reward, f_max_reward, f_attesting_slot FROM t_validator_rewards_summary
	WHERE f_val_idx=$1 AND f_slot=$2;
	`
)

type ValidatorRewards struct {
	ValidatorIndex   uint64
	Slot             uint64
	Epoch            uint64
	ValidatorBalance uint64
	Reward           int64
	MaxReward        uint64
	AttSlot          uint64
}

func NewValidatorRewards(iValIdx uint64, iSlot uint64, iEpoch uint64, iValBal uint64, iReward int64, iMaxReward, iAttSlot uint64) ValidatorRewards {
	return ValidatorRewards{
		ValidatorIndex:   iValIdx,
		Slot:             iSlot,
		Epoch:            iEpoch,
		ValidatorBalance: iValBal,
		Reward:           iReward,
		MaxReward:        iMaxReward,
		AttSlot:          iAttSlot,
	}
}

func NewValidatorRewardsFromSingleEpochMetrics(iMetrics SingleEpochMetrics) ValidatorRewards {
	return NewValidatorRewards(iMetrics.ValidatorIdx, iMetrics.Slot, iMetrics.Epoch, iMetrics.ValidatorBalance, iMetrics.Reward, iMetrics.MaxReward, iMetrics.AttSlot)
}

func NewEmptyValidatorRewards() ValidatorRewards {
	return ValidatorRewards{}
}
