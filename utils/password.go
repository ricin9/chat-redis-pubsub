package utils

/* thanks https://www.alexedwards.net/blog/how-to-hash-and-verify-passwords-with-argon2-in-go */

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"errors"
	"fmt"
	"strings"

	"golang.org/x/crypto/argon2"
)

type hashParams struct {
	time    uint32
	memory  uint32
	threads uint8
	keyLen  uint32
	saltLen uint32
}

func HashPassword(password string) (string, error) {
	params := &hashParams{
		time:    1,
		memory:  64 * 1024,
		threads: 4,
		keyLen:  32,
		saltLen: 16,
	}

	salt, err := generateSalt(params.saltLen)
	if err != nil {
		return "", err
	}

	hash := argon2.IDKey([]byte(password), salt, params.time, params.memory, params.threads, params.keyLen)

	encoded := encodePassword(hash, salt, params)
	return encoded, nil
}

func ComparePassword(encoded, password string) (bool, error) {
	hash, salt, params, err := decodePassword(encoded)
	if err != nil {
		return false, err
	}

	comparisonHash := argon2.IDKey([]byte(password), salt, params.time, params.memory, params.threads, params.keyLen)

	if subtle.ConstantTimeCompare(hash, comparisonHash) == 1 {
		return true, nil
	} else {
		return false, nil
	}
}
func encodePassword(hash, salt []byte, params *hashParams) string {

	b64hash := base64.RawStdEncoding.EncodeToString(hash)
	b64salt := base64.RawStdEncoding.EncodeToString(salt)

	encoded := fmt.Sprintf("$argod2id$%d$%d$%d$%d$%s$%s", params.time, params.memory, params.threads, params.keyLen, b64salt, b64hash)
	return encoded
}

func generateSalt(len uint32) ([]byte, error) {
	b := make([]byte, len)

	_, err := rand.Read(b)
	if err != nil {
		return nil, err
	}
	return b, nil
}

func decodePassword(encoded string) (hash, salt []byte, params *hashParams, err error) {
	var b64hash, b64salt string
	params = &hashParams{}

	vals := strings.Split(encoded, "$")
	if len(vals) != 8 {
		return nil, nil, nil, errors.New("invalid encoded hash")
	}

	_, err = fmt.Sscanf(encoded, "$argod2id$%d$%d$%d$%d$", &params.time, &params.memory, &params.threads, &params.keyLen)
	if err != nil {
		return nil, nil, params, err
	}

	b64salt = vals[6]
	b64hash = vals[7]

	hash, err = base64.RawStdEncoding.DecodeString(b64hash)
	if err != nil {
		return nil, nil, params, err
	}

	salt, err = base64.RawStdEncoding.DecodeString(b64salt)
	if err != nil {
		return nil, nil, params, err
	}

	return hash, salt, params, nil
}
