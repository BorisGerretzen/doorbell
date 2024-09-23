package common

import (
	"crypto/hmac"
	"crypto/sha256"
	"fmt"
	"golang.org/x/crypto/bcrypt"
)

func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	return string(bytes), err
}

// VerifyPassword verifies if the given password matches the stored hash.
func VerifyPassword(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

func Sha256(data string) string {
	return fmt.Sprintf("%x", sha256.Sum256([]byte(data)))
}

func HmacSha256(data, secret []byte) string {
	h := hmac.New(sha256.New, secret)
	h.Write(data)
	return fmt.Sprintf("%x", h.Sum(nil))
}
