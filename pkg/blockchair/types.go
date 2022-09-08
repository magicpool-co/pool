package blockchair

/* context objects */

type RawContext struct {
	Code           int       `json:"code"`
	Error          string    `json:"error,omitempty"`
	Source         string    `json:"source"`
	Limit          string    `json:"limit"`
	Offset         string    `json:"offset"`
	Results        int       `json:"results"`
	State          int       `json:"state"`
	MarketPriceUsd float64   `json:"market_price_usd"`
	Cache          *RawCache `json:"cache"`
	API            *RawAPI   `json:"api"`
	Server         string    `json:"server"`
	Time           float32   `json:"time"`
	RenderTime     float32   `json:"render_time"`
	FullTime       float32   `json:"full_time"`
	RequestCost    float32   `json:"request_cost"`
}

type RawCache struct {
	Live     bool    `json:"live"`
	Duration int     `json:"duration"`
	Since    string  `json:"since"`
	Until    string  `json:"until"`
	Time     float32 `json:"time,omitempty"`
}

type RawAPI struct {
	Version         string `json:"version"`
	LastMajorUpdate string `json:"last_major_update"`
	NextMajorUpdate string `json:"next_major_update,omitempty"`
	Documentation   string `json:"documentation"`
	Notice          string `json:"notice"`
}

/* transaction objects */

type TxResponse struct {
	Data    map[string]*TxInfo `json:"data"`
	Context *RawContext        `json:"context"`
}

type TxInfo struct {
	Tx      *RawTx         `json:"transaction"`
	Inputs  []*RawTxInput  `json:"inputs"`
	Outputs []*RawTxOutput `json:"outputs"`
}

type RawTx struct {
	BlockID        int     `json:"block_id"`
	ID             int     `json:"id"`
	Hash           string  `json:"hash"`
	Date           string  `json:"date"`
	Time           string  `json:"time"`
	Size           int     `json:"size"`
	Weight         int     `json:"weight"`
	Version        int     `json:"version"`
	LockTime       int     `json:"lock_time"`
	IsCoinbase     bool    `json:"is_coinbase"`
	HasWitness     bool    `json:"has_witness"`
	InputCount     int     `json:"input_count"`
	OutputCount    int     `json:"output_count"`
	InputTotal     uint64  `json:"input_total"`
	InputTotalUsd  float32 `json:"input_total_usd"`
	OutputTotal    uint64  `json:"output_total"`
	OutputTotalUsd float32 `json:"output_total_usd"`
	Fee            uint64  `json:"fee"`
	FeeUsd         float32 `json:"fee_usd"`
	FeePerKb       float32 `json:"fee_per_kb"`
	FeePerKbUsd    float32 `json:"fee_per_kb_usd"`
	FeePerKwu      float32 `json:"fee_per_kwu"`
	FeePerKwuUsd   float32 `json:"fee_per_kwu_usd"`
	CddTotal       float32 `json:"cdd_total"`
	IsRbf          bool    `json:"is_rbf"`
}

type RawTxInput struct {
	BlockID                 int     `json:"block_id"`
	TransactionID           int     `json:"transaction_id"`
	Index                   uint32  `json:"index"`
	TransactionHash         string  `json:"transaction_hash"`
	Date                    string  `json:"date"`
	Time                    string  `json:"time"`
	Value                   int64   `json:"value"`
	ValueUsd                float32 `json:"value_usd"`
	Recipient               string  `json:"recipient"`
	Type                    string  `json:"type"`
	ScriptHex               string  `json:"script_hex"`
	IsFromCoinbase          bool    `json:"is_from_coinbase"`
	IsSpendable             bool    `json:"is_spendable,omitempty"`
	IsSpent                 bool    `json:"is_spent"`
	SpendingBlockID         int     `json:"spending_block_id"`
	SpendingTransactionID   int     `json:"spending_transaction_id"`
	SpendingIndex           int     `json:"spending_index"`
	SpendingTransactionHash string  `json:"spending_transaction_hash"`
	SpendingDate            string  `json:"spending_date"`
	SpendingTime            string  `json:"spending_time"`
	SpendingValueUsd        float32 `json:"spending_value_usd"`
	SpendingSequence        int64   `json:"spending_sequence"`
	SpendingSignatureHex    string  `json:"spending_signature_hex"`
	SpendingWitness         string  `json:"spending_witness"`
	Lifespan                int     `json:"lifespan"`
	Cdd                     float32 `json:"cdd"`
}

type RawTxOutput struct {
	BlockID                 int     `json:"block_id"`
	TransactionID           int     `json:"transaction_id"`
	Index                   uint32  `json:"index"`
	TransactionHash         string  `json:"transaction_hash"`
	Date                    string  `json:"date"`
	Time                    string  `json:"time"`
	Value                   uint64  `json:"value"`
	ValueUsd                float32 `json:"value_usd"`
	Recipient               string  `json:"recipient"`
	Type                    string  `json:"type"`
	ScriptHex               string  `json:"script_hex"`
	IsFromCoinbase          bool    `json:"is_from_coinbase"`
	IsSpendable             bool    `json:"is_spendable,omitempty"`
	IsSpent                 bool    `json:"is_spent"`
	SpendingBlockID         int     `json:"spending_block_id,omitempty"`
	SpendingTransactionID   int     `json:"spending_transaction_id,omitempty"`
	SpendingIndex           int     `json:"spending_index,omitempty"`
	SpendingTransactionHash string  `json:"spending_transaction_hash,omitempty"`
	SpendingDate            string  `json:"spending_date,omitempty"`
	SpendingTime            string  `json:"spending_time,omitempty"`
	SpendingValueUsd        float32 `json:"spending_value_usd,omitempty"`
	SpendingSequence        int64   `json:"spending_sequence,omitempty"`
	SpendingSignatureHex    string  `json:"spending_signature_hex,omitempty"`
	SpendingWitness         string  `json:"spending_witness,omitempty"`
	Lifespan                int     `json:"lifespan,omitempty"`
	Cdd                     float32 `json:"cdd,omitempty"`
}

/* address objects */

type AddressResponse struct {
	Data    map[string]*AddressInfo `json:"data"`
	Context *RawContext             `json:"context"`
}

type AddressInfo struct {
	Address *RawAddress `json:"address"`
	// @NOTE: we're excluding transactions bc we don't need them
	UTXO []*RawUTXO `json:"utxo"`
}

type RawAddress struct {
	Path               string  `json:"path,omitempty"`
	Type               string  `json:"type"`
	ScriptHex          string  `json:"script_hex"`
	Balance            uint64  `json:"balance"`
	BalanceUsd         float32 `json:"balance_usd"`
	Received           uint64  `json:"received"`
	ReceivedUsd        float32 `json:"received_usd"`
	Spent              float32 `json:"spent"`
	SpentUsd           float32 `json:"spent_usd"`
	OutputCount        int     `json:"output_count"`
	UnspentOutputCount int     `json:"unspent_output_count"`
	FirstSeenReceiving string  `json:"first_seen_receiving,omitempty"`
	LastSeenReceiving  string  `json:"last_seen_receiving,omitempty"`
	FirstSeenSpending  string  `json:"first_seen_spending,omitempty"`
	LastSeenSpending   string  `json:"last_seen_spending,omitempty"`
	ScripthashType     string  `json:"scripthash_type,omitempty"`
	TransactionCount   int     `json:"transaction_count,omitempty"`
}

type RawUTXO struct {
	BlockID         int    `json:"block_id"`
	TransactionHash string `json:"transaction_hash"`
	Index           int    `json:"index"`
	Value           int    `json:"value"`
}

/* broadcast objects */

type BroadcastResponse struct {
	Data    *BroadcastInfo `json:"data"`
	Context *RawContext    `json:"context"`
}

type BroadcastInfo struct {
	TxID string `json:"transaction_hash"`
}
