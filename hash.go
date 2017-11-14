package main

import (
	"crypto/rand"
	"crypto/sha256"
)

const (
	HASH_SIZE   = 256
	BUCKET_SIZE = HASH_SIZE / 8
)

func NewHash(val []byte) []byte {
	h := sha256.New()

	h.Write(val)

	return h.Sum(nil)
}

func NewRandomHash() []byte {
	res := make([]byte, BUCKET_SIZE)

	rand.Read(res)

	return res
}
