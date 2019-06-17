package main

const (
	dbFile              = "block-chain.db"
	blocksBucket        = "blocksBucket"
	subsidy             = 50
	genesisCoinbaseData = "Genesis data"
	version             = byte(0x00)
	addressChecksumLen  = 4
)

func main() {
	cli := NewCLI()
	cli.Run()
}
