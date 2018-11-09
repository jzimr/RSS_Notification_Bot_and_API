package main

import (
	"log"
	"math/rand"
	"strconv"
	"time"
)

// Valid characters for our keygen: 0-9, a-z, A-Z, -, _
var chars = [64]string{"0", "1", "2", "3", "4", "5", "6", "7", "8", "9", "a", "b",
	"c", "d", "e", "f", "g", "h", "i", "j", "k", "l", "m", "n",
	"o", "p", "q", "r", "s", "t", "u", "v", "w", "x", "y", "z",
	"A", "B", "C", "D", "E", "F", "G", "H", "I", "J", "K", "L",
	"M", "N", "O", "P", "Q", "R", "S", "T", "U", "V", "W", "X",
	"Y", "Z", "-", "_"}

/*
generateKey generates an 25 character long key consisting of numbers and letters
*/
func generateNewKey() (key string) {
	// New seed for every 1/1,000,000,000 second passed in real time
	rand.Seed(time.Now().UTC().UnixNano())

	var newKey string

	for i := 0; i < 20; i++ {
		newKey += strconv.Itoa(genRandInt())
	}

	log.Println("Generated a new key: " + newKey)

	return newKey
}

/*
Generate a random int between 0 and 63
*/
func genRandInt() (num int) {
	return rand.Intn(63)
}

func isKeyTaken() (isTaken bool) {
	// TODO: Check if key is not taken yet
	return false
}
