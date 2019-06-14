package main

// Transaction stores inputs and outputs
type Transaction struct {
	ID   []byte
	Vin  []TxInput
	Vout []TxOutput
}

// TxInput input of transactions
type TxInput struct {
	Txid      []byte
	Vout      int
	ScriptSig string
}

// TxOutput output of transactions
type TxOutput struct {
	Value        int
	ScriptPubKey string
}
