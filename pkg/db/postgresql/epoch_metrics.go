package postgresql

import (
	"context"

	"github.com/cortze/eth2-state-analyzer/pkg/model"
	"github.com/cortze/eth2-state-analyzer/pkg/utils"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/pkg/errors"
)

func (p *PostgresDBService) createEpochMetricsTable(ctx context.Context, pool *pgxpool.Pool) error {
	// create the tables
	_, err := pool.Exec(ctx, model.CreateEpochMetricsTable)
	if err != nil {
		return errors.Wrap(err, "error creating epoch metrics table")
	}
	return nil
}

func (p *PostgresDBService) InsertNewEpochRow(iEpochObj model.EpochMetrics) error {

	_, err := p.psqlPool.Exec(p.ctx, model.InsertNewEpochLineTable, iEpochObj.Epoch, iEpochObj.Slot, 0, 0, iEpochObj.TotalBalance, iEpochObj.TotalEffectiveBalance)

	if err != nil {
		return errors.Wrap(err, "error inserting row in epoch metrics table")
	}
	return nil
}

func (p *PostgresDBService) UpdatePrevEpochAtt(iEpochObj model.EpochMetrics) error {

	if iEpochObj.Slot > utils.SlotBase {
		log.Debugf("updating row %d from epoch metrics", iEpochObj.Slot-utils.SlotBase)
		_, err := p.psqlPool.Exec(p.ctx, model.UpdateAttestation, iEpochObj.PrevNumAttestations, iEpochObj.PrevNumValidators, iEpochObj.Slot-utils.SlotBase)
		if err != nil {
			return errors.Wrap(err, "error updating row in epoch metrics table")
		}
		return nil
	} else {
		log.Debugf("not updating row as we are in the first epoch")
		return nil
	}

}

func (p *PostgresDBService) GetEpochRow(iEpoch uint64) (model.EpochMetrics, error) {

	row := p.psqlPool.QueryRow(p.ctx, model.SelectByEpoch, iEpoch)
	epochRow := model.NewEmptyEpochMetrics()

	err := row.Scan(&epochRow.Epoch, &epochRow.Slot, &epochRow.PrevNumAttestations, &epochRow.PrevNumValidators)

	if err != nil {
		return model.NewEmptyEpochMetrics(), errors.Wrap(err, "error retrieving row from epoch metrics table")
	}
	return epochRow, nil
}
