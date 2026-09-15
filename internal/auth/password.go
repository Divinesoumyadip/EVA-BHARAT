package auth

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"errors"
	"fmt"
	"strconv"
	"strings"
)

const (
	pbkdf2Iterations = 100000
	pbkdf2KeyLength  = 32
	saltLength       = 16
)

func HashPassword(password string) (string, error) {
	salt := make([]byte, saltLength)
	if _, err := rand.Read(salt); err != nil {
		return "", err
	}
	derived := pbkdf2(password, salt, pbkdf2Iterations, pbkdf2KeyLength)
	encoded := fmt.Sprintf("pbkdf2$%d$%s$%s",
		pbkdf2Iterations,
		base64.RawStdEncoding.EncodeToString(salt),
		base64.RawStdEncoding.EncodeToString(derived),
	)
	return encoded, nil
}

func VerifyPassword(password, encoded string) (bool, error) {
	parts := strings.Split(encoded, "$")
	if len(parts) != 4 || parts[0] != "pbkdf2" {
		return false, errors.New("invalid hash format")
	}
	iterations, err := strconv.Atoi(parts[1])
	if err != nil {
		return false, errors.New("invalid hash format")
	}
	salt, err := base64.RawStdEncoding.DecodeString(parts[2])
	if err != nil {
		return false, errors.New("invalid hash format")
	}
	expected, err := base64.RawStdEncoding.DecodeString(parts[3])
	if err != nil {
		return false, errors.New("invalid hash format")
	}
	derived := pbkdf2(password, salt, iterations, len(expected))
	return subtle.ConstantTimeCompare(derived, expected) == 1, nil
}

func pbkdf2(password string, salt []byte, iterations, keyLength int) []byte {
	hashLength := sha256.Size
	numBlocks := (keyLength + hashLength - 1) / hashLength
	derived := make([]byte, 0, numBlocks*hashLength)

	for block := 1; block <= numBlocks; block++ {
		mac := hmac.New(sha256.New, []byte(password))
		mac.Write(salt)
		blockIndex := []byte{
			byte(block >> 24),
			byte(block >> 16),
			byte(block >> 8),
			byte(block),
		}
		mac.Write(blockIndex)
		u := mac.Sum(nil)
		result := make([]byte, hashLength)
		copy(result, u)

		for i := 1; i < iterations; i++ {
			mac := hmac.New(sha256.New, []byte(password))
			mac.Write(u)
			u = mac.Sum(nil)
			for j := range result {
				result[j] ^= u[j]
			}
		}
		derived = append(derived, result...)
	}

	return derived[:keyLength]
}
