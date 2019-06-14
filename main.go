package main

const (
	dbFile       = "block-chain.db"
	blocksBucket = "blocksBucket"
)

const (
	cmdAddBlock   = "addblock"
	argsData      = "data"
	cmdPrintChain = "printchain"
)

func main() {
	chain := NewBlockchain()

	// chain.AddBlock("Send 1 BTC to zw")
	// chain.AddBlock("Send 1 more BTC to yy")

	cli := NewCLI(chain)
	cli.Run()
}
