package main

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"strings"

	"golang.org/x/crypto/scrypt"
)

type PasswordHash struct {
	SaltByteSize    int
	HashKeySize     int
	HashIterations  int
	HashBlockSize   int
	HashParallelism int
	Salt            string
	Checksum        string
}

func (p *PasswordHash) UnmarshalText(text []byte) error {
	_, err := fmt.Sscanf(strings.ReplaceAll(string(text), "$", " "), "scrypt %d %d %d %d %d %s %s",
		&p.SaltByteSize,
		&p.HashKeySize,
		&p.HashIterations,
		&p.HashBlockSize,
		&p.HashParallelism,
		&p.Salt,
		&p.Checksum,
	)
	return err
}

func (p *PasswordHash) String() string {
	return fmt.Sprintf("scrypt$%d$%d$%d$%d$%d$%s$%s",
		p.SaltByteSize,
		p.HashKeySize,
		p.HashIterations,
		p.HashBlockSize,
		p.HashParallelism,
		p.Salt,
		p.Checksum,
	)
}

func HashPassword(password string) (*PasswordHash, error) {
	result := &PasswordHash{
		SaltByteSize:    16,
		HashKeySize:     32,
		HashIterations:  1 << 15,
		HashBlockSize:   8,
		HashParallelism: 1,
	}

	salt := make([]byte, 16)
	if _, err := rand.Read(salt); err != nil {
		return nil, fmt.Errorf("generating password salt: %w", err)
	}
	result.Salt = base64.StdEncoding.EncodeToString(salt)

	checksum, err := scrypt.Key([]byte(password), salt, result.HashIterations, result.HashBlockSize, result.HashParallelism, result.HashKeySize)
	if err != nil {
		return nil, fmt.Errorf("generating password checksum: %w", err)
	}
	result.Checksum = base64.StdEncoding.EncodeToString(checksum)

	return result, nil
}

func ComparedPasswordHashes(hash string, other string) (bool, error) {
	var a PasswordHash
	var b PasswordHash
	if err := a.UnmarshalText([]byte(hash)); err != nil {
		return false, fmt.Errorf("unmarshalling password hash: %w", err)
	}
	if err := b.UnmarshalText([]byte(hash)); err != nil {
		return false, fmt.Errorf("unmarshalling password hash: %w", err)
	}
	return a.Checksum == b.Checksum, nil
}
