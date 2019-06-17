package main

import (
	"bytes"
	"crypto/elliptic"
	"encoding/gob"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"sync"
)

const walletFile = "wallet.db"

// Wallets ...
type Wallets struct {
	Wallets map[string]*Wallet
	mu      *sync.RWMutex
}

// NewWallets ...
func NewWallets() *Wallets {
	wallets := Wallets{}
	wallets.mu = new(sync.RWMutex)
	wallets.Wallets = make(map[string]*Wallet)

	err := wallets.LoadFromFile()
	if err != nil {
		log.Fatal(err)
	}

	return &wallets
}

// CreateWallet ..
func (ws *Wallets) CreateWallet() string {
	wallet := NewWallet()
	address := fmt.Sprintf("%s", wallet.Address())

	ws.mu.Lock()
	ws.Wallets[address] = wallet
	ws.mu.Unlock()

	return address
}

// Addresses ...
func (ws *Wallets) Addresses() []string {
	var addresses []string

	ws.mu.RLock()
	for addr := range ws.Wallets {
		addresses = append(addresses, addr)
	}
	ws.mu.RUnlock()

	return addresses
}

// Wallet get a wallet from wallets
func (ws *Wallets) Wallet(address string) Wallet {
	var wallet *Wallet

	ws.mu.RLock()
	wallet = ws.Wallets[address]
	ws.mu.RUnlock()

	return *wallet
}

// LoadFromFile loads wallets from the file
func (ws *Wallets) LoadFromFile() error {
	if _, err := os.Stat(walletFile); os.IsNotExist(err) {
		return nil
	}

	fileContent, err := ioutil.ReadFile(walletFile)
	if err != nil {
		return err
	}

	var wallets Wallets
	gob.Register(elliptic.P256())
	decoder := gob.NewDecoder(bytes.NewReader(fileContent))

	err = decoder.Decode(&wallets)
	if err != nil {
		return err
	}

	ws.mu.Lock()
	ws.Wallets = wallets.Wallets
	ws.mu.Unlock()

	return nil
}

// SaveToFile saves wallets to a file
func (ws Wallets) SaveToFile() {
	var content bytes.Buffer

	gob.Register(elliptic.P256())

	encoder := gob.NewEncoder(&content)

	ws.mu.RLock()
	err := encoder.Encode(ws)
	ws.mu.RUnlock()

	if err != nil {
		log.Panic(err)
	}

	err = ioutil.WriteFile(walletFile, content.Bytes(), 0644)
	if err != nil {
		log.Panic(err)
	}
}
