package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strconv"
)

const (
	cmdPrintChain = "printchain"
	cmdGetBalance = "getbalance"
	cmdSend       = "send"
)

// CLI the command-line interface of blockchain
type CLI struct{}

// NewCLI ...
func NewCLI() *CLI {
	return &CLI{}
}

// Run start cli
func (cli *CLI) Run() {
	printChainCmd := flag.NewFlagSet(cmdPrintChain, flag.ExitOnError)
	getBalanceCmd := flag.NewFlagSet(cmdGetBalance, flag.ExitOnError)
	sendCmd := flag.NewFlagSet(cmdSend, flag.ExitOnError)

	getBalanceAddress := getBalanceCmd.String("address", "", "The address to get balance for")
	sendFrom := sendCmd.String("from", "", "The origin address of BTC")
	sendTo := sendCmd.String("to", "", "The remote address of BTC")
	sendAmount := sendCmd.Int("amount", 0, "The amount of BTC")

	switch os.Args[1] {
	case cmdPrintChain:
		err := printChainCmd.Parse(os.Args[2:])
		if err != nil {
			log.Fatal(err)
		}
	case cmdGetBalance:
		err := getBalanceCmd.Parse(os.Args[2:])
		if err != nil {
			log.Fatal(err)
		}
	case cmdSend:
		err := sendCmd.Parse(os.Args[2:])
		if err != nil {
			log.Fatal(err)
		}
	default:
		os.Exit(1)
	}

	if printChainCmd.Parsed() {
		cli.printChain()
	}
	if getBalanceCmd.Parsed() {
		cli.getBalance(*getBalanceAddress)
	}
	if sendCmd.Parsed() {
		if len(*sendFrom) == 0 {
			log.Fatal("from cannot be nil")
		}
		if len(*sendTo) == 0 {
			log.Fatal("to cannot be nil")
		}
		if *sendAmount <= 0 {
			log.Fatal("amount must greater than 0")
		}
		cli.send(*sendFrom, *sendTo, *sendAmount)
	}
}

func (cli *CLI) printChain() {
	chain := NewBlockchain("")
	defer chain.db.Close()

	iter := chain.Iterator()

	for {
		block := iter.Next()

		fmt.Printf("Prev.hash : %x\n", block.PrevBlockHash)
		fmt.Printf("Hash : %x\n", block.Hash)
		pow := NewProofOfWork(block)
		fmt.Printf("PoW: %s\n", strconv.FormatBool(pow.Validate()))
		fmt.Println()

		if len(iter.currentHash) == 0 {
			break
		}
	}
}

func (cli *CLI) getBalance(address string) {
	bc := NewBlockchain(address)
	defer bc.db.Close()

	balance := 0

	utxos := bc.FindUTXO(address)
	for _, out := range utxos {
		balance += out.Value
	}

	fmt.Printf("Balance of '%v' : %d\n", address, balance)
}

func (cli *CLI) send(from, to string, amount int) {
	bc := NewBlockchain(from)
	defer bc.db.Close()

	tx := NewUTXOTransaction(from, to, amount, bc)

	bc.AddBlock([]*Transaction{tx})

	fmt.Println("success")
}
