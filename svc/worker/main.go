package main

import (
	"flag"
	"fmt"

	"github.com/magicpool-co/pool/app/worker"
	"github.com/magicpool-co/pool/core/trade"
	"github.com/magicpool-co/pool/internal/log"
	"github.com/magicpool-co/pool/internal/metrics"
	"github.com/magicpool-co/pool/internal/node"
	"github.com/magicpool-co/pool/internal/pooldb"
	"github.com/magicpool-co/pool/internal/redis"
	"github.com/magicpool-co/pool/internal/telegram"
	"github.com/magicpool-co/pool/internal/tsdb"
	"github.com/magicpool-co/pool/pkg/aws"
	"github.com/magicpool-co/pool/pkg/sshtunnel"
	"github.com/magicpool-co/pool/svc"
	"github.com/magicpool-co/pool/types"
)

var (
	miningChains = []string{"ERG", "ETC", "FIRO", "FLUX", "KAS", "NEXA", "RVN"}
	payoutChains = []string{"BTC", "ETH"}
)

func initTunnel(secrets map[string]string) (*sshtunnel.SSHTunnel, error) {
	keys := []string{"TUNNEL_USER", "TUNNEL_HOST", "TUNNEL_KEYPAIR"}
	for _, k := range keys {
		if _, ok := secrets[k]; !ok {
			return nil, nil
		}
	}

	if secrets["ENVIRONMENT"] != "local" {
		return nil, fmt.Errorf("tunnelling should not be used in any environment other than local")
	}

	keyFile, err := sshtunnel.PrivateKeyFile(secrets["TUNNEL_KEYPAIR"], secrets["TUNNEL_KEYPAIR_PASSWORD"])
	if err != nil {
		return nil, err
	}

	tunnel, err := sshtunnel.New(secrets["TUNNEL_USER"], secrets["TUNNEL_HOST"], keyFile)
	if err != nil {
		return nil, err
	}

	return tunnel, nil
}

func newWorker(
	secrets map[string]string,
	mainnet bool,
	metricsClient *metrics.Client,
) (*worker.Worker, *log.Logger, error) {
	telegramClient, err := telegram.New(secrets)
	if err != nil {
		return nil, nil, err
	}

	logger, err := log.New(secrets, "worker", telegramClient)
	if err != nil {
		return nil, nil, err
	}

	pooldbClient, err := pooldb.New(secrets)
	if err != nil {
		return nil, nil, err
	} else if err := pooldbClient.UpgradeMigrations(); err != nil {
		return nil, nil, err
	}

	tsdbClient, err := tsdb.New(secrets)
	if err != nil {
		return nil, nil, err
	} else if err := tsdbClient.UpgradeMigrations(); err != nil {
		return nil, nil, err
	}

	redisClient, err := redis.New(secrets)
	if err != nil {
		return nil, nil, err
	}

	awsClient, err := aws.NewSession(secrets["AWS_REGION"], secrets["AWS_PROFILE"])
	if err != nil {
		return nil, nil, err
	}

	kucoin, err := trade.NewExchange(types.KucoinID, secrets["KUCOIN_API_KEY"],
		secrets["KUCOIN_API_SECRET"], secrets["KUCOIN_API_PASSPHRASE"])
	if err != nil {
		return nil, nil, err
	}

	mexc, err := trade.NewExchange(types.MEXCGlobalID, secrets["MEXC_API_KEY"], secrets["MEXC_API_SECRET"], "")
	if err != nil {
		return nil, nil, err
	}
	exchanges := []types.Exchange{kucoin, mexc}

	tunnel, err := initTunnel(secrets)
	if err != nil {
		return nil, nil, err
	}

	miningNodes := make([]types.MiningNode, 0)
	payoutNodes := make([]types.PayoutNode, 0)
	for _, chain := range miningChains {
		urls, err := pooldb.GetNodeURLsByChain(pooldbClient.Reader(), chain, mainnet)
		if err != nil {
			return nil, nil, err
		} else if len(urls) == 0 {
			logger.Info(fmt.Sprintf("ignoring %s, no node hosts found", chain))
			continue
		}

		for i, url := range urls {
			if tunnel != nil {
				urls[i] = "tunnel://" + url
			} else {
				urls[i] = "http://" + url
			}
		}

		node, err := node.GetMiningNode(mainnet, chain, secrets[chain+"_PRIVATE_KEY"], urls, logger, tunnel)
		if err != nil {
			return nil, nil, err
		}
		miningNodes = append(miningNodes, node)
		payoutNodes = append(payoutNodes, node)
	}

	for _, chain := range payoutChains {
		priv := secrets[chain+"_PRIVATE_KEY"]
		url := secrets[chain+"_NODE_URL"]
		blockchairKey := secrets["BLOCKCHAIR_API_KEY"]
		node, err := node.GetPayoutNode(mainnet, chain, priv, blockchairKey, url, logger)
		if err != nil {
			return nil, nil, err
		}
		payoutNodes = append(payoutNodes, node)
	}

	workerClient, err := worker.NewWorker(secrets["ENVIRONMENT"], mainnet, logger, miningNodes, payoutNodes,
		pooldbClient, tsdbClient, redisClient, awsClient, metricsClient, exchanges, telegramClient)

	return workerClient, logger, err
}

func main() {
	argMainnet := flag.Bool("mainnet", true, "Whether or not to run on the mainnet")
	argSecretVar := flag.String("secret", "", "ENV variable defined by ECS")
	argMetricsPort := flag.Int("metrics-port", 6060, "The metrics port to use")

	flag.Parse()

	secrets, err := svc.ParseSecrets(*argSecretVar)
	if err != nil {
		panic(err)
	}

	metricsClient, err := initMetrics(secrets["ENVIRONMENT"], *argMetricsPort)
	if err != nil {
		panic(err)
	}

	workerServer, logger, err := newWorker(secrets, *argMainnet, metricsClient)
	if err != nil {
		panic(err)
	}

	logger.Info("worker running")

	runner := svc.NewRunner(logger)
	runner.AddWorker(workerServer)
	runner.AddHTTPServer(metricsClient.Server())
	runner.Run()
}
