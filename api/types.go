package main

type answer struct {
	Addr        string `json:"Addr"`
	Txhash      string `json:"Txhash"`
	Raw         string `json:"Raw"`
	Block       string `json:"Block"`
	Blockhash   string `json:"Blockhash"`
	Blockheight string `json:"Blockheight"`
	Blocktime   string `json:"Blocktime"`
}

type toDB struct {
	*Block
	*Tx
}

type Block struct {
	Hash         string `json:"hash"`
	Ver          int    `json:"ver"`
	PrevBlock    string `json:"prev_block"`
	MrklRoot     string `json:"mrkl_root"`
	Time         int    `json:"time"`
	Bits         int    `json:"bits"`
	Nonce        int    `json:"nonce"`
	NTx          int    `json:"n_tx"`
	Size         int    `json:"size"`
	BlockIndex   int    `json:"block_index"`
	MainChain    bool   `json:"main_chain"`
	Height       int    `json:"height"`
	ReceivedTime int    `json:"received_time"`
	RelayedBy    string `json:"relayed_by"`
	Tx           []*Tx  `json:"tx"`
	TxIndexes    []int  `json:"txIndexes"`
}

type Tx struct {
	Result      int       `json:"result"`
	Ver         int       `json:"ver"`
	Size        int       `json:"size"`
	Inputs      []*Inputs `json:"inputs"`
	Time        int       `json:"time"`
	BlockHeight int       `json:"block_height"`
	TxIndex     int       `json:"tx_index"`
	VinSz       int       `json:"vin_sz"`
	Hash        string    `json:"hash"`
	VoutSz      int       `json:"vout_sz"`
	RelayedBy   string    `json:"relayed_by"`
	Out         []*Out    `json:"out"`
}

type Inputs struct {
	Sequence int      `json:"sequence"`
	Script   string   `json:"script"`
	PrevOut  *PrevOut `json:"prev_out"`
}

type PrevOut struct {
	Spent   bool   `json:"spent"`
	TxIndex int    `json:"tx_index"`
	Type    int    `json:"type"`
	Addr    string `json:"addr"`
	Value   int    `json:"value"`
	N       int    `json:"n"`
	Script  string `json:"script"`
}

type Out struct {
	Spent   bool   `json:"spent"`
	TxIndex int    `json:"tx_index"`
	Type    int    `json:"type"`
	Addr    string `json:"addr"`
	Value   int    `json:"value"`
	N       int    `json:"n"`
	Script  string `json:"script"`
}
