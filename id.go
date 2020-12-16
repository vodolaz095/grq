package grq

import (
	"crypto/rand"
	"encoding/hex"
)

// getRandomID gets random hex encoded id
func getRandomID() (id string, err error) {
	b := make([]byte, 10)
	_, err = rand.Read(b)
	if err != nil {
		return
	}
	id = hex.EncodeToString(b)
	return
}
