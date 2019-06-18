package main

import "bytes"

// TxInput input of transactions
type TxInput struct {
	Txid      []byte
	Vout      int
	Signature []byte
	PubKey    []byte
}

// CanUnlockOutputWith ...
func (in *TxInput) CanUnlockOutputWith(unlockData []byte) bool {
	hash := HashPublicKey(in.PubKey)
	return bytes.Compare(unlockData, hash) == 0
}

// func NewTxInput()
