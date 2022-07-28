package postgresql

import (
	"context"

	"github.com/cortze/eth2-state-analyzer/pkg/model"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/pkg/errors"
)

func (p *PostgresDBService) createRewardsTable(ctx context.Context, pool *pgxpool.Pool) error {
	// create the tables
	_, err := pool.Exec(ctx, model.CreateValidatorRewardsTable)
	if err != nil {
		return errors.Wrap(err, "error creating rewards table")
	}
	return nil
}

func (p *PostgresDBService) InsertNewValidatorRow(valRewardsObj model.ValidatorRewards) error {

	_, err := p.psqlPool.Exec(p.ctx, model.InsertNewValidatorLineTable, valRewardsObj.ValidatorIndex, valRewardsObj.Slot, valRewardsObj.Epoch, valRewardsObj.ValidatorBalance, valRewardsObj.Reward, valRewardsObj.MaxReward, valRewardsObj.AttSlot)
	if err != nil {
		return errors.Wrap(err, "error inserting row in validator rewards table")
	}
	return nil
}

func (p *PostgresDBService) UpdateValidatorRowReward(valRewardsObj model.ValidatorRewards) error {

	_, err := p.psqlPool.Exec(p.ctx, model.UpdateValidatorLineTable, valRewardsObj.ValidatorIndex, valRewardsObj.Slot, valRewardsObj.Reward)
	if err != nil {
		return errors.Wrap(err, "error inserting row in validator rewards table")
	}
	return nil
}

func (p *PostgresDBService) GetValidatorRow(iValIdx uint64, iSlot uint64) (model.ValidatorRewards, error) {

	row := p.psqlPool.QueryRow(p.ctx, model.SelectByValSlot, iValIdx, iSlot)
	validatorRow := model.NewEmptyValidatorRewards()

	err := row.Scan(&validatorRow.ValidatorIndex, &validatorRow.Slot, &validatorRow.ValidatorBalance, &validatorRow.Reward, &validatorRow.MaxReward)

	if err != nil {
		return model.NewEmptyValidatorRewards(), errors.Wrap(err, "error retrieving row from validator rewards table")
	}
	return validatorRow, nil
}
