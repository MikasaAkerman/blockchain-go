package main

import (
	"bytes"
	"crypto/ecdsa"
	"encoding/hex"
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
func (bc *Blockchain) AddBlock(trans []*Transaction) *Block {
	for _, tx := range trans {
		if !bc.VerifyTransaction(tx) {
			log.Fatal("AddBlock() Invalid transaction")
		}
	}

	newBlock := NewBlock(trans, bc.tip)
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

		bc.tip = newBlock.Hash
		return nil
	})
	if err != nil {
		log.Fatal(err)
	}

	return newBlock
}

// NewBlockchain create a block chain with a genesis block
func NewBlockchain(address string) (*Blockchain, bool) {
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
	var created bool
	db, err := bolt.Open(dbFile, 0600, nil)
	if err != nil {
		log.Fatal(err)
	}
	err = db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(blocksBucket))
		if b == nil {
			cbTX := NewCoinbaseTX(address, genesisCoinbaseData)
			genesis := NewGenesisBlock(cbTX)
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
			created = true
		} else {
			tip = b.Get([]byte("l"))
		}
		return nil
	})
	if err != nil {
		log.Fatal(err)
	}

	return &Blockchain{db, tip}, created
}

// FindTransaction ...
func (bc *Blockchain) FindTransaction(id []byte) (Transaction, error) {
	bci := bc.Iterator()

	for {
		block := bci.Next()

		for _, tx := range block.Transactions {
			if bytes.Compare(tx.ID, id) == 0 {
				return *tx, nil
			}
		}

		if len(block.PrevBlockHash) == 0 {
			break
		}
	}

	return Transaction{}, errors.New("Transaction not found")
}

func (bc *Blockchain) findPrevTx(tx *Transaction) map[string]Transaction {
	prevTXs := make(map[string]Transaction)

	if tx.IsCoinbase() {
		return prevTXs
	}

	for _, in := range tx.Vin {
		prevTX, err := bc.FindTransaction(in.Txid)
		if err != nil {
			log.Fatal(err)
		}
		prevTXs[hex.EncodeToString(in.Txid)] = prevTX
	}
	return prevTXs
}

// SignTransactioin ...
func (bc *Blockchain) SignTransactioin(tx *Transaction, privKey ecdsa.PrivateKey) {

	tx.Sign(privKey, bc.findPrevTx(tx))
}

// VerifyTransaction ...
func (bc *Blockchain) VerifyTransaction(tx *Transaction) bool {
	return tx.Verify(bc.findPrevTx(tx))
}

// FindUnspentTransactions ...
func (bc *Blockchain) FindUnspentTransactions() []Transaction {
	var uTxs []Transaction             // unspent transactions
	var sTxOs = make(map[string][]int) // spented transaction outputs
	iter := bc.Iterator()

	for {
		block := iter.Next()

		for _, tx := range block.Transactions {
			txID := hex.EncodeToString(tx.ID)

		loop1:
			for index := range tx.Vout {
				if sTxOs[txID] != nil {
					for _, spentOut := range sTxOs[txID] {
						if spentOut == index {
							continue loop1
						}
					}
				}
				uTxs = append(uTxs, *tx)
			}

			if !tx.IsCoinbase() {
				for _, in := range tx.Vin {
					inTxID := hex.EncodeToString(in.Txid)
					sTxOs[inTxID] = append(sTxOs[inTxID], in.Vout)
				}
			}
		}

		if len(block.PrevBlockHash) == 0 {
			break
		}
	}

	return uTxs
}

// FindUTXO find unspent transaction outputs
func (bc *Blockchain) FindUTXO() map[string]TxOutputs {
	var txos = make(map[string]TxOutputs)
	utxs := bc.FindUnspentTransactions()
	for _, tx := range utxs {
		txID := hex.EncodeToString(tx.ID)
		txos[txID] = tx.Vout
	}
	return txos
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
