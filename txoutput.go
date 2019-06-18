package main

import (
	"bytes"
	"encoding/gob"
	"log"
)

// TxOutput output of transactions
type TxOutput struct {
	Value      int
	PubKeyHash []byte
}

// CanUnlockedWith ...
func (out *TxOutput) CanUnlockedWith(unlockData []byte) bool {
	return bytes.Compare(out.PubKeyHash, unlockData) == 0
}

// Lock lock the output by given address
// set output's publickey
func (out *TxOutput) Lock(address []byte) {
	payload := Base58Decode(address)
	pubKeyHash := payload[len([]byte{version}) : len(payload)-addressChecksumLen]
	out.PubKeyHash = pubKeyHash
}

// NewTxOutput ...
func NewTxOutput(value int, address string) *TxOutput {
	txo := TxOutput{value, nil}
	txo.Lock([]byte(address))

	return &txo
}

// TxOutputs ...
type TxOutputs []TxOutput

// Serialize ...
func (out TxOutputs) Serialize() []byte {
	var buf bytes.Buffer
	encoder := gob.NewEncoder(&buf)
	err := encoder.Encode(out)
	if err != nil {
		log.Fatal(err)
	}

	return buf.Bytes()
}

// DeserializeOutputs ...
func DeserializeOutputs(d []byte) *TxOutputs {
	var b TxOutputs
	decoder := gob.NewDecoder(bytes.NewReader(d))

	err := decoder.Decode(&b)
	if err != nil {
		log.Fatal(err)
	}
	return &b
}
