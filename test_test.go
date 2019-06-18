package main

import (
	"log"
	"testing"
)

func TestSinVerify(t *testing.T) {
	bc := NewBlockchain("19KM6QZTCNZiQnDXt5MsCVS9KxqM6UBHJd")
	defer bc.db.Close()

	tx := NewUTXOTransaction("1377khvXDZ2vemhCYSuD1ShbNFT5Dc6DCq", "19KM6QZTCNZiQnDXt5MsCVS9KxqM6UBHJd", 10, bc)

	if !bc.VerifyTransaction(tx) {
		log.Fatal("verify failed")
	}
}
