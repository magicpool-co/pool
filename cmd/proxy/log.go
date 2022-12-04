package main

import (
	"log"
)

const (
	clientColor = "\033[1;34m%s\033[0m"
	serverColor = "\033[1;33m%s\033[0m"
	debugColor  = "\033[0;36m%s\033[0m"
	errorColor  = "\033[1;31m%s\033[0m"
)

func sanitizeMessage(msg string) (string, bool) {
	if length := len(msg); length == 0 {
		return "", false
	} else {
		if msg[length-1] == '\n' {
			msg = msg[:length-1]
		}

		return msg, true
	}
}

func logServer(raw string) {
	if msg, ok := sanitizeMessage(raw); ok {
		log.Printf(serverColor, msg)
	}
}

func logClient(raw string) {
	if msg, ok := sanitizeMessage(raw); ok {
		log.Printf(clientColor, msg)
	}
}

func logDebug(msg string) {
	log.Printf(debugColor, msg)
}

func logError(err error) {
	log.Printf(errorColor, err)
}

func logPanic(err error) {
	log.Panicf(errorColor, err)
}
