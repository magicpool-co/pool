package worker

import (
	"github.com/robfig/cron/v3"

	"github.com/magicpool-co/pool/internal/log"
	"github.com/magicpool-co/pool/internal/metrics"
	"github.com/magicpool-co/pool/internal/redis"
	"github.com/magicpool-co/pool/internal/telegram"
	"github.com/magicpool-co/pool/pkg/aws"
	"github.com/magicpool-co/pool/pkg/dbcl"
	"github.com/magicpool-co/pool/types"
)

type Worker struct {
	env         string
	mainnet     bool
	cron        *cron.Cron
	logger      *log.Logger
	miningNodes []types.MiningNode
	payoutNodes []types.PayoutNode
	pooldb      *dbcl.Client
	tsdb        *dbcl.Client
	redis       *redis.Client
	exchange    types.Exchange
	aws         *aws.Client
	metrics     *metrics.Client
	telegram    *telegram.Client
}

func NewWorker(env string, mainnet bool, logger *log.Logger, miningNodes []types.MiningNode, payoutNodes []types.PayoutNode, pooldbClient, tsdbClient *dbcl.Client, redisClient *redis.Client, awsClient *aws.Client, metricsClient *metrics.Client, exchange types.Exchange, telegramClient *telegram.Client) *Worker {
	cronClient := cron.New(
		cron.WithParser(
			cron.NewParser(
				cron.SecondOptional | cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow)))

	worker := &Worker{
		env:         env,
		mainnet:     mainnet,
		cron:        cronClient,
		logger:      logger,
		miningNodes: miningNodes,
		payoutNodes: payoutNodes,
		redis:       redisClient,
		pooldb:      pooldbClient,
		tsdb:        tsdbClient,
		exchange:    exchange,
		aws:         awsClient,
		metrics:     metricsClient,
		telegram:    telegramClient,
	}

	return worker
}

func (w *Worker) Start() {
	locker := w.redis.NewLocker()

	if w.env != "local" {
		w.cron.AddJob("* * * * *", &NodeStatusJob{
			locker:  locker,
			logger:  w.logger,
			nodes:   w.miningNodes,
			pooldb:  w.pooldb,
			metrics: w.metrics,
		})

		w.cron.AddJob("*/5 * * * *", &NodeCheckJob{
			env:     w.env,
			mainnet: w.mainnet,
			locker:  locker,
			logger:  w.logger,
			aws:     w.aws,
			pooldb:  w.pooldb,
		})

		w.cron.AddJob("*/5 * * * *", &NodeBackupJob{
			env:     w.env,
			mainnet: w.mainnet,
			locker:  locker,
			logger:  w.logger,
			aws:     w.aws,
			pooldb:  w.pooldb,
		})

		w.cron.AddJob("*/5 * * * *", &NodeUpdateJob{
			env:     w.env,
			mainnet: w.mainnet,
			locker:  locker,
			logger:  w.logger,
			aws:     w.aws,
			pooldb:  w.pooldb,
		})

		w.cron.AddJob("*/5 * * * *", &NodeResizeJob{
			env:     w.env,
			mainnet: w.mainnet,
			locker:  locker,
			logger:  w.logger,
			aws:     w.aws,
			pooldb:  w.pooldb,
		})
	}

	w.cron.AddJob("* * * * *", &BlockUnlockJob{
		locker: locker,
		logger: w.logger,
		pooldb: w.pooldb,
		nodes:  w.miningNodes,
	})

	w.cron.AddJob("* * * * *", &BlockCreditJob{
		locker: locker,
		logger: w.logger,
		pooldb: w.pooldb,
		nodes:  w.miningNodes,
	})

	w.cron.AddJob("*/5 * * * *", &AuditJob{
		locker: locker,
		logger: w.logger,
		pooldb: w.pooldb,
		nodes:  w.payoutNodes,
	})

	w.cron.AddJob("*/2 * * * *", &MinerJob{
		locker: locker,
		logger: w.logger,
		redis:  w.redis,
		pooldb: w.pooldb,
		nodes:  w.miningNodes,
	})

	w.cron.AddJob("*/5 * * * *", &TradeJob{
		locker:   locker,
		logger:   w.logger,
		pooldb:   w.pooldb,
		redis:    w.redis,
		nodes:    w.payoutNodes,
		exchange: w.exchange,
		telegram: w.telegram,
	})

	// w.cron.AddJob("*/5 * * * *", &PayoutJob{
	// 	locker:   locker,
	// 	logger:   w.logger,
	// 	pooldb:   w.pooldb,
	//  redis: w.redis,
	// 	nodes:    w.payoutNodes,
	// 	telegram: w.telegram,
	// })

	w.cron.AddJob("* * * * *", &ChartJob{
		locker: locker,
		logger: w.logger,
		redis:  w.redis,
		pooldb: w.pooldb,
		tsdb:   w.tsdb,
		nodes:  w.miningNodes,
	})

	w.cron.Start()
}

func (w *Worker) Stop() {
	ctx := w.cron.Stop()

	select {
	case <-ctx.Done():
		return
	}
}
