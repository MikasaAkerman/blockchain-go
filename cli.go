package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strconv"
)

// CLI the command-line interface of blockchain
type CLI struct {
	bc *Blockchain
}

// NewCLI ...
func NewCLI(bc *Blockchain) *CLI {
	return &CLI{bc}
}

// Run start cli
func (cli *CLI) Run() {
	addBlockCmd := flag.NewFlagSet(cmdAddBlock, flag.ExitOnError)
	printChainCmd := flag.NewFlagSet(cmdPrintChain, flag.ExitOnError)
	addBlockData := addBlockCmd.String(argsData, "", "Block data")

	switch os.Args[1] {
	case cmdAddBlock:
		err := addBlockCmd.Parse(os.Args[2:])
		if err != nil {
			log.Fatal(err)
		}
	case cmdPrintChain:
		err := printChainCmd.Parse(os.Args[2:])
		if err != nil {
			log.Fatal(err)
		}
	default:
		// TODO print usage
		os.Exit(1)
	}

	if addBlockCmd.Parsed() {
		if *addBlockData == "" {
			addBlockCmd.Usage()
			os.Exit(1)
		}
		// TODO ...
		cli.addBlock(*addBlockData)
	}
	if printChainCmd.Parsed() {
		// TODO print chain
		cli.printChain()
	}
}

func (cli *CLI) addBlock(data string) {
	cli.bc.AddBlock(data)
	log.Println("success")
}

func (cli *CLI) printChain() {
	iter := cli.bc.Iterator()

	for {
		block := iter.Next()

		fmt.Printf("Prev.hash : %x\n", block.PrevBlockHash)
		fmt.Printf("Data : %s\n", block.Data)
		fmt.Printf("Hash : %x\n", block.Hash)
		pow := NewProofOfWork(block)
		fmt.Printf("PoW: %s\n", strconv.FormatBool(pow.Validate()))
		fmt.Println()

		if len(iter.currentHash) == 0 {
			break
		}
	}
}
