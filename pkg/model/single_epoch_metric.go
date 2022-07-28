package model

type SingleEpochMetrics struct {
	ValidatorIdx     uint64
	Slot             uint64
	Epoch            uint64
	ValidatorBalance uint64  // Gwei ?¿
	MaxReward        uint64  // Gwei ?¿
	Reward           int64   // Gweis ?¿
	RewardPercentage float64 // %
	AttSlot          uint64

	MissingSource uint64
	MissingHead   uint64
	MissingTarget uint64
}
