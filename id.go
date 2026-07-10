package grq

import (
	"crypto/rand"
	"encoding/hex"
)

const keyLength = 10

// getRandomID gets random hex encoded id
func getRandomID() (id string, err error) {
	b := make([]byte, keyLength)
	_, err = rand.Read(b)
	if err != nil {
		return
	}
	id = hex.EncodeToString(b)
	return
}
