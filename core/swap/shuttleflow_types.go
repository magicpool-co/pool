package swap

type ShuttleflowSwap struct {
	FromChain   string `json:"from_chain"`
	ToChain     string `json:"to_chain"`
	Token       string `json:"token"`
	InOrOut     string `json:"in_or_out"`
	Type        string `json:"type"`
	NonceOrTxID string `json:"nonce_or_txid"`
	Amount      string `json:"amount"`
	UserAddr    string `json:"user_addr"`
	DefiAddr    string `json:"defi"`
	ToAddr      string `json:"to_addr"`
	Status      string `json:"finished"`
	Timestamp   int    `json:"timestamp"`
	SettledTx   string `json:"settled_tx"`
	TxTo        string `json:"tx_to"`
	TxInput     string `json:"tx_input"`
}

type ShuttleflowToken struct {
	ID               int    `json:"id"`
	Origin           string `json:"origin"`
	ToChain          string `json:"to_chain"`
	Reference        string `json:"reference"`
	ReferenceName    string `json:"reference_name"`
	ReferenceSymbol  string `json:"reference_symbol"`
	BurnFee          string `json:"burn_fee"`
	MintFee          string `json:"mint_fee"`
	WalletFee        string `json:"wallet_fee"`
	MinimalMintValue string `json:"minimal_mint_value"`
	MinimalBurnValue string `json:"minimal_burn_value"`
	Name             string `json:"name"`
	Symbol           string `json:"symbol"`
	Decimals         int    `json:"decimals"`
	TotalSupply      string `json:"total_supply"`
	Sponsor          string `json:"sponsor"`
	SponsorValue     string `json:"sponsor_value"`
	CToken           string `json:"ctoken"`
	IsAdmin          int    `json:"is_admin"`
	InTokenList      int    `json:"in_token_list"`
	Icon             string `json:"in_token_list"`
	Supported        int    `json:"supported"`
}
