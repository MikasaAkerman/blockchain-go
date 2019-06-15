package main

const (
	dbFile              = "block-chain.db"
	blocksBucket        = "blocksBucket"
	subsidy             = 50
	genesisCoinbaseData = "Genesis data"
)

func main() {
	cli := NewCLI()
	cli.Run()
}
