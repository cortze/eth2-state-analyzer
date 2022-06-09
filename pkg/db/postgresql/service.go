package postgresql

import (
	"context"
	"strings"

	"github.com/cortze/eth2-state-analyzer/pkg/model"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

// Static postgres queries, for each modification in the tables, the table needs to be reseted
var (
	// logrus associated with the postgres db
	PsqlType = "postgres-db"
	log      = logrus.WithField(
		"module", PsqlType,
	)
)

type PostgresDBService struct {
	// Control Variables
	ctx           context.Context
	cancel        context.CancelFunc
	connectionUrl string // the url might not be necessary (better to remove it?¿)
	psqlPool      *pgxpool.Pool
	// Network DB Model
}

// Connect to the PostgreSQL Database and get the multithread-proof connection
// from the given url-composed credentials
func ConnectToDB(ctx context.Context, url string) (*PostgresDBService, error) {
	mainCtx, cancel := context.WithCancel(ctx)
	// spliting the url to don't share any confidential information on logs
	log.Infof("Conneting to postgres DB %s", url)
	if strings.Contains(url, "@") {
		log.Debugf("Connecting to PostgresDB at %s", strings.Split(url, "@")[1])
	}
	psqlPool, err := pgxpool.Connect(mainCtx, url)
	if err != nil {
		return nil, err
	}
	if strings.Contains(url, "@") {
		log.Infof("PostgresDB %s succesfully connected", strings.Split(url, "@")[1])
	}
	// filter the type of network that we are filtering

	psqlDB := &PostgresDBService{
		ctx:           mainCtx,
		cancel:        cancel,
		connectionUrl: url,
		psqlPool:      psqlPool,
	}
	// init the psql db
	err = psqlDB.init(ctx, psqlDB.psqlPool)
	if err != nil {
		return psqlDB, errors.Wrap(err, "error initializing the tables of the psqldb")
	}
	return psqlDB, err
}

// Close the connection with the PostgreSQL
func (p *PostgresDBService) Close() {
}

func (p *PostgresDBService) init(ctx context.Context, pool *pgxpool.Pool) error {
	// create the tables
	err := p.createRewardsTable(ctx, pool)
	if err != nil {
		return err
	}
	return nil
}

func (p *PostgresDBService) createRewardsTable(ctx context.Context, pool *pgxpool.Pool) error {
	// create the tables
	_, err := pool.Exec(ctx, model.CreateValidatorRewardsTable)
	if err != nil {
		return errors.Wrap(err, "error creating rewards table")
	}
	return nil
}

func (p *PostgresDBService) InsertNewValidatorRow(epochMetrics model.SingleEpochMetrics) error {

	valRewardsObj := model.NewValidatorRewardsFromSingleEpochMetrics(epochMetrics)

	_, err := p.psqlPool.Exec(p.ctx, model.InsertNewLineTable, valRewardsObj.ValidatorIndex, valRewardsObj.Slot, valRewardsObj.ValidatorBalance)
	if err != nil {
		return errors.Wrap(err, "error inserting row in validator rewards table")
	}
	return nil
}

func (p *PostgresDBService) GetValidatorRow(iValIdx uint64, iSlot uint64) (model.ValidatorRewards, error) {

	row := p.psqlPool.QueryRow(p.ctx, model.SelectByVal, iValIdx, iSlot)
	validatorRow := model.NewEmptyValidatorRewards()

	err := row.Scan(&validatorRow.ValidatorIndex, &validatorRow.Slot, &validatorRow.ValidatorBalance)

	if err != nil {
		return model.NewEmptyValidatorRewards(), errors.Wrap(err, "error retrieving row from validator rewards table")
	}
	return validatorRow, nil
}