package main

import (
	"flag"
	"fmt"

	"github.com/magicpool-co/pool/app/worker"
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
	chains = []string{"AE", "CFX", "CTXC", "ERGO", "ETC", "FIRO", "FLUX", "RVN"}
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

func newWorker(secrets map[string]string, mainnet bool, metricsClient *metrics.Client) (*worker.Worker, *log.Logger, error) {
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

	tunnel, err := initTunnel(secrets)
	if err != nil {
		return nil, nil, err
	}

	nodes := make([]types.MiningNode, 0)
	for _, chain := range chains {
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
		nodes = append(nodes, node)
	}

	workerClient := worker.NewWorker(secrets["ENVIRONMENT"], mainnet, logger, nodes, pooldbClient, tsdbClient, redisClient, awsClient, metricsClient)

	return workerClient, logger, nil
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
