package handlers

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"math/big"
	"os"
	"time"
)

// Simply removes a file
func removeFile(p string) error {
	if err := os.Remove(p); err != nil {
		return err
	}

	return nil
}

// Generates a hash using the given string, the actual unix time and a random 3 sequence number
func genFileNameHash(fn string) string {
	unixTime := fmt.Sprint(time.Now().Unix())
	salt := randStr(3)
	algo := sha256.New()
	algo.Write([]byte(fn + unixTime + salt))
	return hex.EncodeToString(algo.Sum(nil))
}

// Generates a random stringth of the length n
func randStr(n int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

	result := make([]byte, n)
	for i := range result {
		idx, _ := rand.Int(rand.Reader, big.NewInt(int64(len(charset))))
		result[i] = charset[idx.Int64()]
	}

	return string(result)
}
