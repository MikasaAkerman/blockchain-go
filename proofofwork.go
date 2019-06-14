package main

import (
	"bytes"
	"crypto/sha256"
	"fmt"
	"math"
	"math/big"
	"time"
)

const targetBits = 24

// ProofOfWork the proof of work
type ProofOfWork struct {
	block  *Block
	target *big.Int
}

// NewProofOfWork create a proof of work
func NewProofOfWork(b *Block) *ProofOfWork {
	target := big.NewInt(1)
	target.Lsh(target, uint(256-targetBits))
	return &ProofOfWork{b, target}
}

func (pow *ProofOfWork) prepartData(nonce int) []byte {
	data := bytes.Join(
		[][]byte{
			pow.block.PrevBlockHash,
			pow.block.Data,
			IntToHex(pow.block.Timestamp),
			IntToHex(targetBits),
			IntToHex(int64(nonce)),
		},
		[]byte{},
	)
	return data
}

// IntToHex conver int to bytes
func IntToHex(i int64) []byte {
	return big.NewInt(i).Bytes()
}

// Run run the mining action
func (pow *ProofOfWork) Run() (int, []byte) {
	var hashInt big.Int
	var hash [32]byte
	nonce := 0

	fmt.Printf("Mining the block containing \"%s\"\n", pow.block.Data)
	start := time.Now()
	for nonce < math.MaxInt64 {
		data := pow.prepartData(nonce)
		hash = sha256.Sum256(data)
		fmt.Printf("\r%x", hash)
		hashInt.SetBytes(hash[:])
		if hashInt.Cmp(pow.target) == -1 {
			break
		} else {
			nonce++
		}
	}
	fmt.Println(" Mining costs: ", time.Now().Sub(start))
	fmt.Println()

	return nonce, hash[:]
}

// Validate verify proof of work
func (pow *ProofOfWork) Validate() bool {
	var hashInt big.Int
	data := pow.prepartData(pow.block.Nonce)
	hash := sha256.Sum256(data)
	hashInt.SetBytes(hash[:])
	isValid := hashInt.Cmp(pow.target) == -1
	return isValid
}
