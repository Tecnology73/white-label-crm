package hash

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"fmt"
	"golang.org/x/crypto/argon2"
	"strings"
)

// Tweak Memory & Threads based on the system running this app.
// With those configured, benchmark how long the Hash function
// takes, increasing Time to the highest possible value that
// you find acceptable for how long it takes.
const (
	PasswordTime       = 1
	PasswordMemory     = 64 * 1024
	PasswordThreads    = 4
	PasswordSaltLength = 16
	PasswordKeyLength  = 32
)

type Argon2Options struct {
	Time       uint32
	Memory     uint32
	Threads    uint8
	SaltLength uint32
	KeyLength  uint32
}

// Hash takes in a raw password (typically user-provided) and hashes the
// password using argon2id - returning the encoded version.
func Hash(password string, opts *Argon2Options) (string, error) {
	salt := make([]byte, opts.SaltLength)
	if _, err := rand.Read(salt); err != nil {
		return "", err
	}

	hash := argon2.IDKey(
		[]byte(password),
		salt,
		opts.Time,
		opts.Memory,
		opts.Threads,
		opts.KeyLength,
	)

	return encodeArgon2(hash, salt, opts), nil
}

// Compare takes in a password (typically user-provided) and an
// encodedPassword (a password that's previously hashed with Hash
// - typically stored in the database) and returns nil if the passwords match.
func Compare(password string, encodedPassword string) error {
	// Decode the encoded hash (i.e. password from database)
	hash, salt, opts, err := decodeArgon2(encodedPassword)
	if err != nil {
		return err
	}

	// Hash the raw password (i.e. password provided by the user)
	hashedPassword := argon2.IDKey(
		[]byte(password),
		salt,
		opts.Time,
		opts.Memory,
		opts.Threads,
		opts.KeyLength,
	)

	// Check if the hashes are the same
	if subtle.ConstantTimeCompare(hash, hashedPassword) != 1 {
		return fmt.Errorf("invalid password")
	}

	return nil
}

func encodeArgon2(hash []byte, salt []byte, opts *Argon2Options) string {
	return fmt.Sprintf(
		"$argon2id$v=%d$m=%d,t=%d,p=%d$%s$%s",
		argon2.Version,
		opts.Memory,
		opts.Time,
		opts.Threads,
		base64.RawStdEncoding.EncodeToString(salt),
		base64.RawStdEncoding.EncodeToString(hash),
	)
}

func decodeArgon2(encodedPassword string) ([]byte, []byte, *Argon2Options, error) {
	parts := strings.Split(encodedPassword, "$")
	if len(parts) != 6 {
		return nil, nil, nil, fmt.Errorf("invalid hash")
	}

	// Version
	var version int
	if _, err := fmt.Sscanf(parts[2], "v=%d", &version); err != nil {
		return nil, nil, nil, err
	}
	if version != argon2.Version {
		return nil, nil, nil, fmt.Errorf("incompatible version")
	}

	// Memory + Time + Threads
	opts := &Argon2Options{}
	if _, err := fmt.Sscanf(parts[3], "m=%d,t=%d,p=%d", &opts.Memory, &opts.Time, &opts.Threads); err != nil {
		return nil, nil, nil, fmt.Errorf("cannot parse options")
	}

	// Salt
	salt, err := base64.RawStdEncoding.Strict().DecodeString(parts[4])
	if err != nil {
		return nil, nil, nil, fmt.Errorf("cannot decode salt")
	}
	opts.SaltLength = uint32(len(salt))

	// Hash
	hash, err := base64.RawStdEncoding.Strict().DecodeString(parts[5])
	if err != nil {
		return nil, nil, nil, fmt.Errorf("cannot decode hash")
	}
	opts.KeyLength = uint32(len(hash))

	return hash, salt, opts, nil
}
