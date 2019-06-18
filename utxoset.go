package main

import (
	"encoding/hex"
	"log"

	"github.com/boltdb/bolt"
)

// UTxOSet ...
type UTxOSet struct {
	BC *Blockchain
}

// Reindex ...
func (u UTxOSet) Reindex() {
	db := u.BC.db
	bucketName := []byte(utxoBucket)

	err := db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(bucketName)
		if b != nil {
			err := tx.DeleteBucket(bucketName)
			if err != nil {
				return err
			}
		}
		_, err := tx.CreateBucket(bucketName)
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		log.Fatal(err)
	}

	utxos := u.BC.FindUTXO()

	err = db.Update(func(tx *bolt.Tx) error {

		b := tx.Bucket(bucketName)
		for txID, outs := range utxos {
			key, err := hex.DecodeString(txID)
			if err != nil {
				return err
			}
			err = b.Put(key, outs.Serialize())
			if err != nil {
				return err
			}
		}

		return nil
	})
	if err != nil {
		log.Fatal(err)
	}
}

// FindSpendableOutputs ...
func (u UTxOSet) FindSpendableOutputs(address []byte, amount int) (int, map[string][]int) {
	utxos := make(map[string][]int)
	accumulate := 0
	db := u.BC.db

	err := db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(utxoBucket))
		c := b.Cursor()

		for k, v := c.First(); k != nil; k, v = c.Next() {
			txID := hex.EncodeToString(k)
			outs := DeserializeOutputs(v)

			for index, out := range *outs {
				if out.CanUnlockedWith(address) && accumulate < amount {
					accumulate += out.Value
					utxos[txID] = append(utxos[txID], index)
				}
			}
		}

		return nil
	})
	if err != nil {
		log.Fatal(err)
	}

	return accumulate, utxos
}

// FindUTXO ...
func (u UTxOSet) FindUTXO(address []byte) []TxOutput {
	var utxos []TxOutput
	db := u.BC.db

	err := db.View(func(tx *bolt.Tx) error {

		b := tx.Bucket([]byte(utxoBucket))
		c := b.Cursor()

		for k, v := c.First(); k != nil; k, v = c.Next() {
			outs := DeserializeOutputs(v)

			for _, out := range []TxOutput(*outs) {
				if out.CanUnlockedWith(address) {
					utxos = append(utxos, out)
				}
			}
		}

		return nil
	})

	if err != nil {
		log.Fatal(err)
	}

	return utxos
}

// Update ...
func (u UTxOSet) Update(block *Block) {
	db := u.BC.db
	err := db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(utxoBucket))

		for _, tx := range block.Transactions {
			if tx.IsCoinbase() {
				continue
			}

			for _, in := range tx.Vin {
				updatedOuts := TxOutputs{}
				outsBytes := b.Get(in.Txid)

				outs := DeserializeOutputs(outsBytes)
				for index, out := range *outs {
					if in.Vout != index {
						updatedOuts = append(updatedOuts, out)
					}
					if len(updatedOuts) == 0 {
						err := b.Delete(in.Txid)
						if err != nil {
							log.Fatal(err)
						}
					} else {
						err := b.Put(in.Txid, updatedOuts.Serialize())
						if err != nil {
							log.Fatal(err)
						}
					}
				}
			}

			err := b.Put(tx.ID, TxOutputs(tx.Vout).Serialize())
			if err != nil {
				log.Fatal(err)
			}
		}

		return nil
	})

	if err != nil {
		log.Fatal(err)
	}
}
