package feed

import (
	"crypto/rand"
	"math/big"
	"regexp"
)

// IDPrefix is the prefix for all post IDs
const IDPrefix = "smk-"

// IDLength is the length of the random portion of the ID
const IDLength = 6

// base62Chars are the characters used for ID generation
const base62Chars = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"

// idPattern is the regex pattern for valid post IDs
var idPattern = regexp.MustCompile(`^smk-[a-zA-Z0-9]{6}$`)

// GenerateID creates a new unique post ID in the format smk-<6 base62 chars>
func GenerateID() (string, error) {
	result := make([]byte, IDLength)
	base := big.NewInt(int64(len(base62Chars)))

	for i := 0; i < IDLength; i++ {
		n, err := rand.Int(rand.Reader, base)
		if err != nil {
			return "", err
		}
		result[i] = base62Chars[n.Int64()]
	}

	return IDPrefix + string(result), nil
}

// ValidateID checks if a string is a valid post ID
func ValidateID(id string) bool {
	return idPattern.MatchString(id)
}
