package main

import (
	"encoding/hex"
	"fmt"
	"log"
	"math/big"
	"math/rand"
)

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

// SetID generate transaction id
func (t *Transaction) SetID() {
	if len(t.ID) == 0 {
		t.ID = big.NewInt(rand.Int63()).Bytes()
	}
}

// CanUnlockOutputWith ...
func (in *TxInput) CanUnlockOutputWith(unlockDataString string) bool {
	return in.ScriptSig == unlockDataString
}

// CanUnlockedWith ...
func (out *TxOutput) CanUnlockedWith(unlockDataString string) bool {
	return out.ScriptPubKey == unlockDataString
}

// NewCoinbaseTX ...
func NewCoinbaseTX(to, data string) *Transaction {
	if len(data) == 0 {
		data = fmt.Sprintf("Reward to %s", to)
	}
	tin := TxInput{[]byte{}, -1, data}
	tout := TxOutput{subsidy, to}

	tx := Transaction{Vin: []TxInput{tin}, Vout: []TxOutput{tout}}
	tx.SetID()

	return &tx
}

// NewUTXOTransaction ...
func NewUTXOTransaction(from, to string, amount int, bc *Blockchain) *Transaction {
	var inputs []TxInput
	var outputs []TxOutput

	acc, validOutputs := bc.FindSpendableOutputs(from, amount)
	if acc < amount {
		log.Fatalf("Not enough balance: left %d", acc)
	}
	for txid, outs := range validOutputs {
		txID, err := hex.DecodeString(txid)
		if err != nil {
			log.Fatal(err)
		}

		for _, out := range outs {
			inputs = append(inputs, TxInput{txID, out, from})
		}
	}

	outputs = append(outputs, TxOutput{amount, to})
	if acc > amount {
		outputs = append(outputs, TxOutput{acc - amount, from})
	}

	tx := Transaction{nil, inputs, outputs}
	tx.SetID()
	return &tx
}

// IsCoinbase ...
func (t *Transaction) IsCoinbase() bool {
	return len(t.Vin) == 1 && len(t.Vin[0].Txid) == 0 && t.Vin[0].Vout == -1
}
