package main

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"log"

	"golang.org/x/crypto/ripemd160"
)

// Wallet the wallet of block chain
type Wallet struct {
	PrivateKey ecdsa.PrivateKey
	PublicKey  []byte
}

// NewWallet create a new wallet
func NewWallet() *Wallet {
	privateKey, publicKey := newKeyPair()
	return &Wallet{privateKey, publicKey}
}

func newKeyPair() (ecdsa.PrivateKey, []byte) {
	cureve := elliptic.P256()
	privatekey, err := ecdsa.GenerateKey(cureve, rand.Reader)
	if err != nil {
		log.Fatal(err)
	}
	pubkey := append(privatekey.PublicKey.X.Bytes(), privatekey.PublicKey.Y.Bytes()...)

	return *privatekey, pubkey
}

// Address get address of a wallet
func (w Wallet) Address() []byte {
	hash := HashPublicKey(w.PublicKey)
	versionedPayload := append([]byte{version}, hash...)
	checkSum := checkSum(versionedPayload)

	payload := append(versionedPayload, checkSum...)

	return Base58Encode(payload)
}

// HashPublicKey ...
func HashPublicKey(pk []byte) []byte {
	publicSHA256 := sha256.Sum256(pk)

	RIPEMD160Hasher := ripemd160.New()
	_, err := RIPEMD160Hasher.Write(publicSHA256[:])
	if err != nil {
		log.Fatal(err)
	}

	return RIPEMD160Hasher.Sum(nil)
}

func checkSum(payload []byte) []byte {
	firstSHA := sha256.Sum256(payload)
	secondSHA := sha256.Sum256(firstSHA[:])

	return secondSHA[:addressChecksumLen]
}
