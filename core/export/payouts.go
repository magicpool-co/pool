package export

import (
	"fmt"
	"sort"

	"github.com/magicpool-co/pool/core/stats"
)

func generatePayoutCols(inputChains []string, outputChain string) []string {
	cols := []string{
		"Chain",
		"TxID",
		"Date",
		"Confirmed",
	}

	if len(inputChains) == 0 || (len(inputChains) == 1 && inputChains[0] == outputChain) {
		return append(cols, []string{
			"Value",
			"Pool Fees",
			"Tx Fees",
			"Total Fees",
		}...)
	}

	for _, chain := range inputChains {
		cols = append(cols, chain+" Input Value")
	}

	cols = append(cols, []string{
		"Value",
		"Pool Fees",
		"Exchange Fees",
		"Tx Fees",
		"Total Fees",
	}...)

	return cols
}

func generatePayoutRow(payout *stats.Payout, inputChains []string) []string {
	row := []string{
		payout.Chain,
		payout.TxID,
		formatTimestamp(payout.Timestamp),
		formatBool(payout.Confirmed),
	}

	if len(inputChains) == 0 || (len(inputChains) == 1 && inputChains[0] == payout.Chain) {
		return append(row, []string{
			formatFloat64(payout.Value.Value),
			formatFloat64(payout.PoolFees.Value),
			formatFloat64(payout.TxFees.Value),
			formatFloat64(payout.TotalFees.Value),
		}...)
	}

	for _, chain := range inputChains {
		var value float64
		if number, ok := payout.Inputs[chain]; ok {
			value = number.Value
		}

		row = append(row, formatFloat64(value))
	}

	row = append(row, []string{
		formatFloat64(payout.Value.Value),
		formatFloat64(payout.PoolFees.Value),
		formatFloat64(payout.ExchangeFees.Value),
		formatFloat64(payout.TxFees.Value),
		formatFloat64(payout.TotalFees.Value),
	}...)

	return row
}

func exportPayouts(payouts []*stats.Payout) ([]string, [][]string, error) {
	sort.Slice(payouts, func(i, j int) bool {
		return payouts[i].Timestamp < payouts[j].Timestamp
	})

	inputChainIdx := make(map[string]bool)
	var outputChain string
	for _, payout := range payouts {
		for chain := range payout.Inputs {
			inputChainIdx[chain] = true
		}

		if outputChain == "" {
			outputChain = payout.Chain
		} else if outputChain != payout.Chain {
			return nil, nil, fmt.Errorf("multiple output chains for payout export")
		}
	}

	inputChains := make([]string, 0)
	for chain := range inputChainIdx {
		inputChains = append(inputChains, chain)
	}
	sort.Strings(inputChains)

	cols := generatePayoutCols(inputChains, outputChain)
	rows := make([][]string, len(payouts))
	for i, payout := range payouts {
		rows[i] = generatePayoutRow(payout, inputChains)
	}

	return cols, rows, nil
}

func ExportPayoutsAsCSV(payouts []*stats.Payout) ([]byte, error) {
	cols, rows, err := exportPayouts(payouts)
	if err != nil {
		return nil, err
	}

	return writeAsCSV(cols, rows)
}
