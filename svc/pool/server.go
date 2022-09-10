package main

import (
	"fmt"

	"github.com/magicpool-co/pool/app/pool"
	"github.com/magicpool-co/pool/internal/log"
	"github.com/magicpool-co/pool/internal/metrics"
	"github.com/magicpool-co/pool/internal/node"
	"github.com/magicpool-co/pool/internal/pooldb"
	"github.com/magicpool-co/pool/internal/redis"
	"github.com/magicpool-co/pool/internal/telegram"
	"github.com/magicpool-co/pool/pkg/sshtunnel"
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

func newPool(secrets map[string]string, mainnet bool, opts *pool.Options, metricsClient *metrics.Client) (*pool.Pool, *log.Logger, error) {
	telegramClient, err := telegram.New(secrets)
	if err != nil {
		return nil, nil, err
	}

	logger, err := log.New(secrets, "pool", telegramClient)
	if err != nil {
		return nil, nil, err
	}

	dbClient, err := pooldb.New(secrets)
	if err != nil {
		return nil, nil, err
	} else if err := dbClient.UpgradeMigrations(); err != nil {
		return nil, nil, err
	}

	redisClient, err := redis.New(secrets)
	if err != nil {
		return nil, nil, err
	}

	tunnel, err := initTunnel(secrets)
	if err != nil {
		return nil, nil, err
	}

	urls, err := pooldb.GetNodeURLsByChain(dbClient.Reader(), opts.Chain, mainnet)
	if err != nil {
		return nil, nil, err
	}

	for i, url := range urls {
		if tunnel != nil {
			urls[i] = "tunnel://" + url
		} else {
			urls[i] = "http://" + url
		}
	}

	miningNode, err := node.GetMiningNode(mainnet, opts.Chain, secrets[opts.Chain+"_PRIVATE_KEY"], urls, tunnel)
	if err != nil {
		return nil, nil, err
	}

	poolServer, err := pool.New(miningNode, dbClient, redisClient, logger, telegramClient, metricsClient, opts)
	if err != nil {
		return nil, nil, err
	}

	logger.Debug(fmt.Sprintf("initialized node using address %s", miningNode.Address()))

	return poolServer, logger, nil
}
