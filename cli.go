package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strconv"
)

const (
	cmdPrintChain    = "printchain"
	cmdGetBalance    = "getbalance"
	cmdSend          = "send"
	cmdCreateWallet  = "createwallet"
	cmdListAddresses = "listaddresses"
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
	createWalletCmd := flag.NewFlagSet(cmdCreateWallet, flag.ExitOnError)
	listAddressesCmd := flag.NewFlagSet(cmdListAddresses, flag.ExitOnError)

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
	case cmdCreateWallet:
		err := createWalletCmd.Parse(os.Args[2:])
		if err != nil {
			log.Fatal(err)
		}
	case cmdListAddresses:
		err := listAddressesCmd.Parse(os.Args[2:])
		if err != nil {
			log.Fatal(err)
		}
	default:
		log.Printf("unkown cmd: %v", os.Args[1])
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
	if createWalletCmd.Parsed() {
		cli.createWallet()
	}
	if listAddressesCmd.Parsed() {
		cli.listAddresses()
	}
}

func (cli *CLI) printChain() {
	chain := NewBlockchain("")
	defer chain.db.Close()

	iter := chain.Iterator()

	for {
		block := iter.Next()

		fmt.Printf("============ Block %x ============\n", block.Hash)
		fmt.Printf("Prev. block: %x\n", block.PrevBlockHash)
		pow := NewProofOfWork(block)
		fmt.Printf("PoW: %s\n\n", strconv.FormatBool(pow.Validate()))
		for _, tx := range block.Transactions {
			fmt.Println(tx)
		}
		fmt.Printf("\n\n")

		if len(iter.currentHash) == 0 {
			break
		}
	}
}

func (cli *CLI) getBalance(address string) {
	if !ValidateAddress(address) {
		log.Fatal("not valid address")
	}
	bc := NewBlockchain(address)
	defer bc.db.Close()

	pubKeyHash := Base58Decode([]byte(address))
	pubKeyHash = pubKeyHash[1 : len(pubKeyHash)-4]

	balance := 0

	utxos := bc.FindUTXO(pubKeyHash)
	for _, out := range utxos {
		balance += out.Value
	}

	fmt.Printf("Balance of '%v' : %d\n", address, balance)
}

func (cli *CLI) send(from, to string, amount int) {
	if !ValidateAddress(from) {
		log.Fatal("ERROR: Sender address is not valid")
	}
	if !ValidateAddress(to) {
		log.Fatal("ERROR: Recipient address is not valid")
	}

	bc := NewBlockchain(from)
	defer bc.db.Close()

	tx := NewUTXOTransaction(from, to, amount, bc)

	bc.AddBlock([]*Transaction{tx})

	fmt.Println("success")
}

func (cli *CLI) createWallet() {
	wallets := NewWallets()
	address := wallets.CreateWallet()
	wallets.SaveToFile()

	fmt.Printf("Your new address: %s\n", address)
}

func (cli *CLI) listAddresses() {
	wallets := NewWallets()
	addresses := wallets.Addresses()

	fmt.Println("Your wallets address list:")
	for _, address := range addresses {
		fmt.Println("		", address)
	}
}
