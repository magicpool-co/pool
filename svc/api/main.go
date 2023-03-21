package main

import (
	"flag"
	"fmt"
	"net/http"

	"github.com/magicpool-co/pool/app/api"
	"github.com/magicpool-co/pool/internal/log"
	"github.com/magicpool-co/pool/internal/node"
	"github.com/magicpool-co/pool/internal/pooldb"
	"github.com/magicpool-co/pool/internal/redis"
	"github.com/magicpool-co/pool/internal/telegram"
	"github.com/magicpool-co/pool/internal/tsdb"
	"github.com/magicpool-co/pool/svc"
	"github.com/magicpool-co/pool/types"
)

func newAPI(secrets map[string]string, port int) (*http.Server, *log.Logger, error) {
	telegramClient, err := telegram.New(secrets)
	if err != nil {
		return nil, nil, err
	}

	logger, err := log.New(secrets, "api", telegramClient)
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

	chains := []string{"ERG", "ETC", "KAS", "NEXA"}
	nodes := make([]types.MiningNode, len(chains))
	for i, chain := range chains {
		nodes[i], err = node.GetMiningNode(true, chain, secrets[chain+"_PRIVATE_KEY"], nil, logger, nil)
		if err != nil {
			return nil, nil, err
		}
	}

	ctx := api.NewContext(logger, nil, pooldbClient, tsdbClient, redisClient, nodes)
	server := api.New(ctx, port)

	return server, logger, nil
}

func main() {
	argPort := flag.Int("port", 8080, "The port to use")
	argSecretVar := flag.String("secret", "", "ENV variable defined by ECS")

	flag.Parse()

	secrets, err := svc.ParseSecrets(*argSecretVar)
	if err != nil {
		panic(err)
	}

	apiServer, logger, err := newAPI(secrets, *argPort)
	if err != nil {
		panic(err)
	}

	logger.Info(fmt.Sprintf("api running on port %d", *argPort))

	runner := svc.NewRunner(logger)
	runner.AddHTTPServer(apiServer)
	runner.Run()
}
