package security

import (
	"academy/internal/config"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io"
)

type Service struct {
	cipher cipher.Block
}

func NewService(cfg *config.Config) (*Service, error) {
	key, err := hex.DecodeString(cfg.Auth.EncryptionKey)
	if err != nil {
		return nil, fmt.Errorf("failed to decode key: %w", err)
	}

	cipher, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("failed to create AES cipher: %w", err)
	}

	return &Service{
		cipher: cipher,
	}, nil
}

func (s *Service) EncryptString(data string) (string, error) {
	gcm, err := cipher.NewGCM(s.cipher)
	if err != nil {
		return "", fmt.Errorf("failed to create GCM: %w", err)
	}

	nonceSize := gcm.NonceSize()
	nonce := make([]byte, nonceSize)
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", fmt.Errorf("failed to generate nonce: %w", err)
	}

	sealedData := gcm.Seal(nonce, nonce, []byte(data), nil)
	return hex.EncodeToString(sealedData), nil
}

func (s *Service) DecryptString(hash string) (string, error) {
	sealedData, err := hex.DecodeString(hash)
	if err != nil {
		return "", fmt.Errorf("failed to decode hash: %w", err)
	}

	gcm, err := cipher.NewGCM(s.cipher)
	if err != nil {
		return "", fmt.Errorf("failed to create GCM: %w", err)
	}

	nonceSize := gcm.NonceSize()
	if len(sealedData) < nonceSize {
		return "", fmt.Errorf("sealed data is too short")
	}

	nonce := sealedData[:nonceSize]
	ciphertext := sealedData[nonceSize:]

	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return "", fmt.Errorf("failed to decrypt or verify data: %w", err)
	}

	return string(plaintext), nil
}
