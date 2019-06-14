package main

import (
	"errors"
	"log"

	"github.com/boltdb/bolt"
)

// Blockchain the blockchain
type Blockchain struct {
	db  *bolt.DB
	tip []byte
}

// BlockchainIterator the iterator of a blockchain
type BlockchainIterator struct {
	currentHash []byte
	db          *bolt.DB
}

// AddBlock add a block to blockchain
func (bc *Blockchain) AddBlock(data string) {
	newBlock := NewBlock(data, bc.tip)
	err := bc.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(blocksBucket))
		if b == nil {
			return errors.New("bucket is nil")
		}

		err := b.Put(newBlock.Hash, newBlock.Serialize())
		if err != nil {
			return err
		}
		err = b.Put([]byte("l"), newBlock.Hash)
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		log.Fatal(err)
	}
	bc.tip = newBlock.Hash
}

// NewBlockchain create a block chain with a genesis block
func NewBlockchain() *Blockchain {
	/**
	1.Open a DB file.
	2.Check if there’s a blockchain stored in it.
	3.If there’s a blockchain:
		1.Create a new Blockchain instance.
		2.Set the tip of the Blockchain instance to the last block hash stored in the DB.

	4.If there’s no existing blockchain:
		1.Create the genesis block.
		2.Store in the DB.
		3.Save the genesis block’s hash as the last block hash.
		4.Create a new Blockchain instance with its tip pointing at the genesis block.
	*/
	var tip []byte
	db, err := bolt.Open(dbFile, 0600, nil)
	if err != nil {
		log.Fatal(err)
	}
	err = db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(blocksBucket))
		if b == nil {
			genesis := NewGenesisBlock()
			b, err := tx.CreateBucket([]byte(blocksBucket))
			if err != nil {
				return err
			}
			err = b.Put(genesis.Hash, genesis.Serialize())
			if err != nil {
				return err
			}
			err = b.Put([]byte("l"), genesis.Hash)
			if err != nil {
				return err
			}
			tip = genesis.Hash
		} else {
			tip = b.Get([]byte("l"))
		}
		return nil
	})
	if err != nil {
		log.Fatal(err)
	}

	return &Blockchain{db, tip}
}

// Iterator get a iterator of a block chain
func (bc *Blockchain) Iterator() *BlockchainIterator {
	return &BlockchainIterator{bc.tip, bc.db}
}

// Next get next block of block chain
func (bci *BlockchainIterator) Next() *Block {
	var block *Block
	bci.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(blocksBucket))
		d := b.Get(bci.currentHash)
		block = DeserializeBlock(d)
		return nil
	})
	bci.currentHash = block.PrevBlockHash
	return block
}
