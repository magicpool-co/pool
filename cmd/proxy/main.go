package main

import (
	"flag"
)

func main() {
	argHost := flag.String("h", "", "The host")
	argPort := flag.Int("p", 3333, "The port")

	flag.Parse()

	proxy, err := newProxy(*argHost)
	if err != nil {
		logPanic(err)
	}
	defer proxy.close()

	server, err := newServer(*argPort)
	if err != nil {
		logPanic(err)
	}

	go server.listen()
	proxyMsgs, proxyErrs := proxy.recieve()

	for {
		select {
		case msg := <-server.msgs:
			logClient(string(msg))
			if err := proxy.send(msg); err != nil {
				logError(err)
			}
		case msg := <-proxyMsgs:
			logServer(string(msg))
			errs := server.broadcast(msg)
			for _, err := range errs {
				if err != nil {
					logError(err)
				}
			}
		case err := <-proxyErrs:
			logPanic(err)
		}
	}
}
