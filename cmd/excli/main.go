package main

import (
	"flag"
	"log"
	"strings"

	"github.com/magicpool-co/pool/core/trade/binance"
	"github.com/magicpool-co/pool/core/trade/bittrex"
	"github.com/magicpool-co/pool/core/trade/kucoin"
	"github.com/magicpool-co/pool/core/trade/mexc"
	"github.com/magicpool-co/pool/svc"
	"github.com/magicpool-co/pool/types"
)

func main() {
	argExchange := flag.String("ex", "kucoin", "The exchange to use")
	argAction := flag.String("action", "GetAccountStatus", "The action to do")
	argChain := flag.String("chain", "", "The chain")
	argID := flag.String("id", "", "The id")
	argValue := flag.Float64("value", 0, "The value")
	argMarket := flag.String("market", "", "The market")
	argDirection := flag.String("direction", "", "The direction")
	argAddress := flag.String("addr", "", "The address")

	flag.Parse()

	secrets, err := svc.ParseSecrets("")
	if err != nil {
		log.Fatalf("failed to fetch secrets: %v", err)
	}

	var ex types.Exchange
	switch *argExchange {
	case "binance":
		ex = binance.New(secrets["BINANCE_API_KEY"], secrets["BINANCE_API_SECRET"])
	case "bittrex":
		ex = bittrex.New(secrets["BITTREX_API_KEY"], secrets["BITTREX_API_SECRET"])
	case "mexc":
		ex = mexc.New(secrets["MEXC_API_KEY"], secrets["MEXC_API_SECRET"])
	case "kucoin":
		ex = kucoin.New(secrets["KUCOIN_API_KEY"], secrets["KUCOIN_API_SECRET"], secrets["KUCOIN_API_PASSPHRASE"])
	default:
		log.Fatalf("exchange: unsupported exchange %s", *argExchange)
	}

	chain := strings.ToUpper(*argChain)

	switch *argAction {
	case "GetAccountStatus":
		err := ex.GetAccountStatus()
		if err != nil {
			log.Fatalf("account status: %v", err)
		}
		log.Printf("account status: ok")

	case "GetWalletStatus":
		depositsEnabled, withdrawalsEnabled, err := ex.GetWalletStatus(chain)
		if err != nil {
			log.Fatalf("wallet status: %v", err)
		}
		log.Printf("wallet status: deposits enabled - %t, withdrawals enabled - %t", depositsEnabled, withdrawalsEnabled)

	case "GetWalletBalance":
		mainBalance, tradeBalance, err := ex.GetWalletBalance(chain)
		if err != nil {
			log.Fatalf("wallet balance: %v", err)
		}
		log.Printf("wallet balance: main - %f, trade - %f", mainBalance, tradeBalance)

	case "GetDepositAddress":
		address, err := ex.GetDepositAddress(chain)
		if err != nil {
			log.Fatalf("wallet address: %v", err)
		}
		log.Printf("wallet address: %s", address)

	case "GetDepositByTxID":
		deposit, err := ex.GetDepositByTxID(chain, *argID)
		if err != nil {
			log.Fatalf("deposit: %v", err)
		}
		log.Printf("deposit: %v", deposit)

	case "GetDepositByID":
		deposit, err := ex.GetDepositByID(chain, *argID)
		if err != nil {
			log.Fatalf("deposit: %v", err)
		}
		log.Printf("deposit: %v", deposit)

	case "TransferToTradeAccount":
		err := ex.TransferToTradeAccount(chain, *argValue)
		if err != nil {
			log.Fatalf("transfer: %v", err)
		}
		log.Printf("transfer: ok")

	case "TransferToMainAccount":
		err := ex.TransferToMainAccount(chain, *argValue)
		if err != nil {
			log.Fatalf("transfer: %v", err)
		}
		log.Printf("transfer: ok")

	case "CreateTrade":
		var direction types.TradeDirection
		switch strings.ToUpper(*argDirection) {
		case "BUY":
			direction = types.TradeBuy
		case "SELL":
			direction = types.TradeSell
		default:
			log.Fatalf("create trade: unknown trade direction %s", *argDirection)
		}

		id, err := ex.CreateTrade(*argMarket, direction, *argValue)
		if err != nil {
			log.Fatalf("create trade: %v", err)
		}
		log.Printf("create trade: %s", id)

	case "GetTradeByID":
		trade, err := ex.GetTradeByID(*argMarket, *argID, *argValue)
		if err != nil {
			log.Fatalf("trade: %v", err)
		}
		log.Printf("trade: %v", trade)

		// withdrawal
	case "CreateWithdrawal":
		id, err := ex.CreateWithdrawal(chain, *argAddress, *argValue)
		if err != nil {
			log.Fatalf("create withdrawal: %v", err)
		}
		log.Printf("create withdrawal: %s", id)

	case "GetWithdrawalByID":
		withdrawal, err := ex.GetWithdrawalByID(chain, *argID)
		if err != nil {
			log.Fatalf("withdrawal: %v", err)
		}
		log.Printf("withdrawal: %v", withdrawal)

	default:
		log.Fatalf("unknown action %s", *argAction)
	}
}
