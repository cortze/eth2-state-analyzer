package model

// Postgres intregration variables
var (
	CreateEpochMetricsTable = `
	CREATE TABLE IF NOT EXISTS t_epoch_metrics_summary(
		f_epoch INT,
		f_slot INT,
		f_num_att INT,
		f_num_vals INT,
		f_total_balance BIGINT,
		f_total_effective_balance BIGINT,
		CONSTRAINT PK_Epoch PRIMARY KEY (f_slot));`

	InsertNewEpochLineTable = `
	INSERT INTO t_epoch_metrics_summary (f_epoch, f_slot, f_num_att, f_num_vals, f_total_balance, f_total_effective_balance)
	VALUES ($1, $2, $3, $4, $5, $6);
	`

	UpdateAttestation = `
	UPDATE t_epoch_metrics_summary 
	SET f_num_att = $1 , f_num_vals = $2
	WHERE f_slot=$3;
	`

	SelectByEpoch = `
	SELECT f_epoch, f_slot, f_num_att, f_num_vals, f_total_balance, f_total_effective_balance FROM t_epoch_metrics_summary
	WHERE f_epoch=$1;
	`
)

type EpochMetrics struct {
	Epoch                 uint64
	Slot                  uint64
	PrevNumAttestations   uint64
	PrevNumValidators     uint64
	TotalBalance          uint64
	TotalEffectiveBalance uint64
}

func NewEpochMetrics(iEpoch uint64, iSlot uint64, iNumAtt uint64, iNumVals uint64, iTotBal uint64, iTotEfBal uint64) EpochMetrics {
	return EpochMetrics{
		Epoch:                 iEpoch,
		Slot:                  iSlot,
		PrevNumAttestations:   iNumAtt,
		PrevNumValidators:     iNumVals,
		TotalBalance:          iTotBal,
		TotalEffectiveBalance: iTotEfBal,
	}
}

func NewEmptyEpochMetrics() EpochMetrics {
	return EpochMetrics{}
}
