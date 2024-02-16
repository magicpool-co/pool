package main

import (
	"context"
	"flag"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/magicpool-co/pool/pkg/stratum"
	"github.com/magicpool-co/pool/pkg/stratum/rpc"
)

var (
	btcStratumHandshake = []*rpc.Request{
		rpc.MustNewRequest("mining.subscribe"),
		rpc.MustNewRequest("mining.authorize", "0x0000000000000000000000000000000000000000:ETH", "x"),
	}

	cfxStratumHandshake = []*rpc.Request{
		rpc.MustNewRequest("mining.subscribe", "0x0000000000000000000000000000000000000000:ETH", "x"),
	}

	ethStratumHandshake = []*rpc.Request{
		rpc.MustNewRequest("eth_submitLogin", "0x0000000000000000000000000000000000000000:ETH", "x"),
	}
)

func main() {
	argURL := flag.String("u", "localhost:3333", "The URL to use")
	argChain := flag.String("c", "ETC", "The chain to use")
	argCount := flag.Int("n", 10, "The number of clients to create")

	flag.Parse()

	chain := strings.ToUpper(*argChain)

	var handshakeReqs []*rpc.Request
	switch chain {
	case "ERG", "FIRO", "FLUX", "RVN":
		handshakeReqs = btcStratumHandshake
	case "CFX":
		handshakeReqs = cfxStratumHandshake
	case "ETC":
		handshakeReqs = ethStratumHandshake
	default:
		log.Fatalf("chain not supported")
	}

	log.Printf("running load test with %d clients", *argCount)
	for i := 0; i < *argCount; i++ {
		go func(i int) {
			ctx := context.Background()
			client := stratum.NewClient(ctx, *argURL, time.Second, time.Second*5)
			client.Start(handshakeReqs)
		}(i)
	}

	exit := make(chan os.Signal)
	signal.Notify(exit, syscall.SIGTERM)
	signal.Notify(exit, syscall.SIGINT)

	<-exit

	log.Printf("exiting load test")

}
