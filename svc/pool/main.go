package main

import (
	"flag"
	"fmt"
	"strings"
	"time"

	"github.com/magicpool-co/pool/app/pool"
	"github.com/magicpool-co/pool/svc"
)

var defaultOptions = map[string]*pool.Options{
	"AE": &pool.Options{
		Chain:          "AE",
		WindowSize:     100000,
		ExtraNonceSize: 4,
		JobListSize:    5,
		VarDiffEnabled: false,
		StreamEnabled:  false,
		PollingPeriod:  time.Millisecond * 100,
	},
	"CFX": &pool.Options{
		Chain:           "CFX",
		WindowSize:      100000,
		JobListSize:     100,
		JobListAgeLimit: -1,
		VarDiffEnabled:  false,
		StreamEnabled:   false,
		PollingPeriod:   time.Millisecond * 100,
	},
	"CTXC": &pool.Options{
		Chain:           "CTXC",
		WindowSize:      100000,
		JobListSize:     25,
		JobListAgeLimit: 7,
		VarDiffEnabled:  false,
		StreamEnabled:   false,
		PollingPeriod:   time.Millisecond * 100,
	},
	"ERG": &pool.Options{
		Chain:                "ERG",
		WindowSize:           100000,
		ExtraNonceSize:       2,
		JobListSize:          5,
		VarDiffEnabled:       false,
		StreamEnabled:        false,
		ForceErrorOnResponse: true,
		PollingPeriod:        time.Millisecond * 100,
		PingingPeriod:        time.Minute,
	},
	"ETC": &pool.Options{
		Chain:           "ETC",
		WindowSize:      100000,
		JobListSize:     25,
		JobListAgeLimit: 7,
		VarDiffEnabled:  false,
		StreamEnabled:   false,
		PollingPeriod:   time.Millisecond * 100,
	},
	"ETHW": &pool.Options{
		Chain:           "ETHW",
		WindowSize:      2000000,
		JobListSize:     25,
		JobListAgeLimit: 7,
		VarDiffEnabled:  false,
		StreamEnabled:   false,
		PollingPeriod:   time.Millisecond * 100,
	},
	"FIRO": &pool.Options{
		Chain:          "FIRO",
		WindowSize:     300000,
		ExtraNonceSize: 1,
		JobListSize:    5,
		VarDiffEnabled: false,
		StreamEnabled:  false,
		PollingPeriod:  time.Second,
	},
	"FLUX": &pool.Options{
		Chain:          "FLUX",
		WindowSize:     100000,
		ExtraNonceSize: 4,
		JobListSize:    5,
		VarDiffEnabled: false,
		StreamEnabled:  false,
		PollingPeriod:  time.Second,
	},
	"KAS": &pool.Options{
		Chain:           "KAS",
		WindowSize:      100000,
		ExtraNonceSize:  2,
		JobListSize:     50,
		JobListAgeLimit: 30,
		VarDiffEnabled:  true,
		StreamEnabled:   true,
		PollingPeriod:   time.Millisecond * 100,
		PingingPeriod:   time.Second * 10,
	},
	"NEXA": &pool.Options{
		Chain:          "NEXA",
		WindowSize:     150000,
		ExtraNonceSize: 4,
		JobListSize:    5,
		VarDiffEnabled: false,
		StreamEnabled:  false,
		PollingPeriod:  time.Second,
	},
	"RVN": &pool.Options{
		Chain:          "RVN",
		WindowSize:     300000,
		ExtraNonceSize: 1,
		JobListSize:    5,
		VarDiffEnabled: false,
		StreamEnabled:  false,
		PollingPeriod:  time.Second,
	},
}

func main() {
	argChain := flag.String("chain", "ETC", "The chain to run the pool for")
	argMainnet := flag.Bool("mainnet", true, "Whether or not to run on the mainnet")
	argSecretVar := flag.String("secret", "", "ENV variable defined by ECS")
	argStandardPort := flag.Int("port", 3333, "The pool port to use")
	argHighDiffPort := flag.Int("high-diff-port", -1, "The port for high difficulty (-1 is disabled)")
	argExtraHighDiffPort := flag.Int("extra-high-diff-port", -1, "The port for extra high difficulty (-1 is disabled)")
	argSoloEnabled := flag.Bool("solo-enabled", false, "Whether or not to enable solo")
	argMetricsPort := flag.Int("metrics-port", 6060, "The metrics port to use")

	flag.Parse()

	opts, ok := defaultOptions[strings.ToUpper(*argChain)]
	if !ok {
		panic(fmt.Errorf("invalid chain %s", *argChain))
	} else if *argStandardPort > 10000 {
		panic(fmt.Errorf("port must be less than 10000"))
	}

	portDiffIdx := map[int]int{*argStandardPort: 1}
	if *argHighDiffPort != -1 {
		portDiffIdx[*argHighDiffPort] = 16
	}
	if *argExtraHighDiffPort != -1 {
		portDiffIdx[*argExtraHighDiffPort] = 256
	}

	opts.PortDiffIdx = portDiffIdx
	opts.SoloEnabled = *argSoloEnabled
	secrets, err := svc.ParseSecrets(*argSecretVar)
	if err != nil {
		panic(err)
	}

	metricsClient, err := initMetrics(secrets["ENVIRONMENT"], *argMetricsPort)
	if err != nil {
		panic(err)
	}

	poolServer, logger, err := newPool(secrets, *argMainnet, opts, metricsClient)
	if err != nil {
		panic(err)
	}

	logger.Debug(fmt.Sprintf("running server for %s", opts.Chain))

	runner := svc.NewRunner(logger)
	runner.AddTCPServer(poolServer)
	runner.AddHTTPServer(metricsClient.Server())
	runner.Run()
}
