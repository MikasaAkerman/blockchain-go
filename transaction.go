package main

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/elliptic"
	crand "crypto/rand"
	"crypto/sha256"
	"encoding/gob"
	"encoding/hex"
	"fmt"
	"log"
	"math/big"
	"strings"
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
	Signature []byte
	PubKey    []byte
}

// TxOutput output of transactions
type TxOutput struct {
	Value      int
	PubKeyHash []byte
}

// SetID generate transaction id
func (t *Transaction) SetID() {
	t.ID = t.Hash()
}

// CanUnlockOutputWith ...
func (in *TxInput) CanUnlockOutputWith(unlockData []byte) bool {
	hash := HashPublicKey(in.PubKey)
	return bytes.Compare(unlockData, hash) == 0
}

// CanUnlockedWith ...
func (out *TxOutput) CanUnlockedWith(unlockData []byte) bool {
	return bytes.Compare(out.PubKeyHash, unlockData) == 0
}

// func NewTxInput()

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

// NewCoinbaseTX ...
func NewCoinbaseTX(to, data string) *Transaction {
	if len(data) == 0 {
		data = fmt.Sprintf("Reward to %s", to)
	}
	tin := TxInput{[]byte{}, -1, nil, []byte(data)}
	tout := NewTxOutput(subsidy, to)

	tx := Transaction{Vin: []TxInput{tin}, Vout: []TxOutput{*tout}}
	tx.SetID()

	return &tx
}

// NewUTXOTransaction ...
func NewUTXOTransaction(from, to string, amount int, bc *Blockchain) *Transaction {
	var inputs []TxInput
	var outputs []TxOutput

	wallets := NewWallets()
	wallet := wallets.Wallet(from)

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
			inputs = append(inputs, TxInput{txID, out, nil, []byte(from)})
		}
	}

	outputs = append(outputs, *NewTxOutput(amount, to))
	if acc > amount {
		outputs = append(outputs, *NewTxOutput(acc-amount, from))
	}

	tx := Transaction{nil, inputs, outputs}
	tx.SetID()
	bc.SignTransactioin(&tx, wallet.PrivateKey)

	return &tx
}

// IsCoinbase ...
func (t *Transaction) IsCoinbase() bool {
	return len(t.Vin) == 1 && len(t.Vin[0].Txid) == 0 && t.Vin[0].Vout == -1
}

// TrimmedCopy ...
func (t *Transaction) TrimmedCopy() *Transaction {
	var inputs []TxInput
	var outputs []TxOutput
	for _, input := range t.Vin {
		inputs = append(inputs, TxInput{input.Txid, input.Vout, nil, nil})
	}
	for _, output := range t.Vout {
		outputs = append(outputs, TxOutput{output.Value, output.PubKeyHash})
	}

	return &Transaction{t.ID, inputs, outputs}
}

// Hash ...
func (t *Transaction) Hash() []byte {
	copyTx := *t
	copyTx.ID = []byte{}

	hash := sha256.Sum256(copyTx.Serialize())

	return hash[:]
}

// Sign ...
func (t *Transaction) Sign(privKey ecdsa.PrivateKey, prevTXs map[string]Transaction) {
	if t.IsCoinbase() {
		return
	}

	copyTX := t.TrimmedCopy()

	for index, in := range copyTX.Vin {
		prevTX := prevTXs[hex.EncodeToString(in.Txid)]
		copyTX.Vin[index].Signature = nil
		copyTX.Vin[index].PubKey = prevTX.Vout[in.Vout].PubKeyHash
		copyTX.ID = copyTX.Hash()
		copyTX.Vin[index].PubKey = nil

		r, s, err := ecdsa.Sign(crand.Reader, &privKey, copyTX.ID)
		if err != nil {
			log.Fatal(err)
		}

		sig := append(r.Bytes(), s.Bytes()...)

		t.Vin[index].Signature = sig
	}
}

// Verify ...
func (t *Transaction) Verify(prevTXs map[string]Transaction) bool {
	copyTX := t.TrimmedCopy()
	curve := elliptic.P256()

	for index, in := range copyTX.Vin {
		prevTX := prevTXs[hex.EncodeToString(in.Txid)]
		copyTX.Vin[index].Signature = nil
		copyTX.Vin[index].PubKey = prevTX.Vout[in.Vout].PubKeyHash
		copyTX.ID = copyTX.Hash()
		copyTX.Vin[index].PubKey = nil

		r := big.Int{}
		s := big.Int{}

		sigLen := len(in.Signature)
		r.SetBytes(in.Signature[:(sigLen / 2)])
		s.SetBytes(in.Signature[(sigLen / 2):])

		x := big.Int{}
		y := big.Int{}
		keyLen := len(in.PubKey)
		x.SetBytes(in.PubKey[:(keyLen / 2)])
		y.SetBytes(in.PubKey[(keyLen / 2):])

		rawPubKey := ecdsa.PublicKey{curve, &x, &y}

		if !ecdsa.Verify(&rawPubKey, copyTX.ID, &r, &s) {
			return false
		}
	}

	return true
}

// Serialize ...
func (t *Transaction) Serialize() []byte {
	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)
	err := enc.Encode(t)
	if err != nil {
		log.Fatal(err)
	}
	return buf.Bytes()
}

// String returns a human-readable representation of a transaction
func (t Transaction) String() string {
	var lines []string

	lines = append(lines, fmt.Sprintf("--- Transaction %x:", t.ID))

	for i, input := range t.Vin {

		lines = append(lines, fmt.Sprintf("     Input %d:", i))
		lines = append(lines, fmt.Sprintf("       TXID:      %x", input.Txid))
		lines = append(lines, fmt.Sprintf("       Out:       %d", input.Vout))
		lines = append(lines, fmt.Sprintf("       Signature: %x", input.Signature))
		lines = append(lines, fmt.Sprintf("       PubKey:    %x", input.PubKey))
	}

	for i, output := range t.Vout {
		lines = append(lines, fmt.Sprintf("     Output %d:", i))
		lines = append(lines, fmt.Sprintf("       Value:  %d", output.Value))
		lines = append(lines, fmt.Sprintf("       Script: %x", output.PubKeyHash))
	}

	return strings.Join(lines, "\n")
}
