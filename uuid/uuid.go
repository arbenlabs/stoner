package uuid

import (
	"crypto/rand"
	"fmt"
	"regexp"
	"strings"
)

// **************************************************
// UUID
// UUID generates and validates UUIDs
// **************************************************

// UUID represents a 128-bit UUID
type UUID [16]byte

// String returns the string representation of the UUID
func (u UUID) String() string {
	return fmt.Sprintf("%08x-%04x-%04x-%04x-%012x",
		u[0:4], u[4:6], u[6:8], u[8:10], u[10:16])
}

// NewV4 generates a new random UUID (version 4)
func NewV4() (UUID, error) {
	var uuid UUID

	// Generate 16 random bytes
	_, err := rand.Read(uuid[:])
	if err != nil {
		return UUID{}, err
	}

	// Set version (4) in the 7th byte
	uuid[6] = (uuid[6] & 0x0f) | 0x40

	// Set variant bits in the 9th byte
	uuid[8] = (uuid[8] & 0x3f) | 0x80

	return uuid, nil
}

// NewUUIDString generates a new UUID string
func NewUUIDString() (string, error) {
	uuid, err := NewV4()
	if err != nil {
		return "", err
	}
	return uuid.String(), nil
}

// NewWithNamespace generates a new UUID string with a namespace
func NewWithNamespace(namespace string) (string, error) {
	uuid, err := NewV4()
	if err != nil {
		return "", err
	}
	id := fmt.Sprintf("%s-%s", uuid.String(), namespace)
	return id, nil
}

// MustNew generates a new UUID string or panics
func MustNewUUIDString() string {
	uuid, err := NewUUIDString()
	if err != nil {
		panic(err)
	}
	return uuid
}

// Parse parses a UUID string into a 16 byte UUID struct
func Parse(s string) (UUID, error) {
	var uuid UUID

	// Validate format
	matched, err := regexp.MatchString(`^[0-9a-f]{8}-[0-9a-f]{4}-4[0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$`, s)
	if err != nil || !matched {
		return UUID{}, fmt.Errorf("invalid UUID format")
	}

	// Remove hyphens and parse
	s = strings.ReplaceAll(s, "-", "")
	if len(s) != 32 {
		return UUID{}, fmt.Errorf("invalid UUID length")
	}

	// Convert hex string to bytes
	for i := 0; i < 16; i++ {
		var b byte
		_, err := fmt.Sscanf(s[i*2:i*2+2], "%02x", &b)
		if err != nil {
			return UUID{}, fmt.Errorf("invalid hex character")
		}
		uuid[i] = b
	}

	return uuid, nil
}

// IsValid checks if a string is a valid UUID format
func IsValid(s string) bool {
	matched, _ := regexp.MatchString(`^[0-9a-f]{8}-[0-9a-f]{4}-4[0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$`, s)
	return matched
}

// MustNewV4 generates a new UUID or panics
func MustNewV4() UUID {
	uuid, err := NewV4()
	if err != nil {
		panic(err)
	}
	return uuid
}
