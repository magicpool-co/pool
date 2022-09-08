package mock

func GetNextNonce(address string) []byte {
	return nil
}

func GetBalance(address string) []byte {
	return nil
}

func GetStatus() []byte {
	return []byte(`{"difficulty":1448983380448,"genesis_key_block_hash":"kh_pbtwgLrNu23k9PA6XCZnUbtsvEFeQGgavY4FS2do3QP8kcp2z","listening":true,"network_id":"ae_mainnet","node_revision":"a01662672b33feab273af79eed5e98535eb74308","node_version":"6.5.1","peer_connections":{"inbound":0,"outbound":10},"peer_count":77,"peer_pubkey":"pp_vPnVXVwxyQFGY1JeHjFqge1Zq5asBPNHSXk6AhEZhFeyHFX96","pending_transactions_count":5,"protocols":[{"effective_at_height":441444,"version":5},{"effective_at_height":161150,"version":4},{"effective_at_height":90800,"version":3},{"effective_at_height":47800,"version":2},{"effective_at_height":0,"version":1}],"solutions":0,"sync_progress":100.0,"syncing":false,"top_block_height":649328,"top_key_block_hash":"kh_QQ612ruuWUKLBmaub4Uwe37MWQ6n4dqsW5KnRfXFiXf4Nu8ev"}`)
}

func GetBlock(height uint64) []byte {
	return nil
}

func GetPendingBlock() []byte {
	return []byte(`{"beneficiary":"ak_2Tf5zs6vginLjcqksfwRi2WNkL6YZsytgN1npXUdTz4ZeirfFz","hash":"kh_eMExHL7CxGeLwRF7U65MtmsUnssa9djzY1CEGoWEUPPZKbj4X","height":649329,"info":"cb_AAACi/oayNs=","miner":"ak_fpBfNohXL7YGFn3XQCyusTM2ojdguS5AAbKUvy7ZGkb2vkkfH","prev_hash":"mh_2C3XtPHoG5474sstTr5ByBKcRAKvHALmU6jBB2wBUsYpCwWw56","prev_key_hash":"kh_QQ612ruuWUKLBmaub4Uwe37MWQ6n4dqsW5KnRfXFiXf4Nu8ev","state_hash":"bs_2dv3wfcEDTLKCqUuLvL31hTeYBPzA33MHgJM4igHopqpD77bk8","target":553713663,"time":1661924424713,"version":5}`)
}

func PostBlock(hostID string, block interface{}) []byte {
	return []byte(`{}`)
}

func PostTransaction(tx string) []byte {
	return []byte(`{"tx_hash": ""}`)
}
