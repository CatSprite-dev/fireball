package session

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"time"

	"github.com/google/uuid"
)

type SessionManager struct {
	store  *Store
	secret []byte
}

type SessionData struct {
	EncryptedToken string    `json:"encrypted_token"`
	AccountID      string    `json:"account_id"`
	OpenedDate     time.Time `json:"opened_date"`
}

func NewManager(store *Store, secret string) (*SessionManager, error) {
	key, err := hex.DecodeString(secret)
	if err != nil || len(key) != 32 {
		return nil, fmt.Errorf("SESSION_SECRET must be a 32-byte hex-encoded string")
	}
	return &SessionManager{
		store:  store,
		secret: key,
	}, nil
}

func (m *SessionManager) encrypt(plaintext string) (string, error) {
	block, err := aes.NewCipher(m.secret)
	if err != nil {
		return "", err
	}
	aesgcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}
	nonce := make([]byte, aesgcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}

	ciphertext := aesgcm.Seal(nonce, nonce, []byte(plaintext), nil)
	return hex.EncodeToString(ciphertext), nil
}

func (m *SessionManager) decrypt(cyphertextHex string) (string, error) {
	data, err := hex.DecodeString(cyphertextHex)
	if err != nil {
		return "", err
	}
	block, err := aes.NewCipher(m.secret)
	if err != nil {
		return "", err
	}
	aesgcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}
	nonceSize := aesgcm.NonceSize()
	if len(data) < nonceSize {
		return "", fmt.Errorf("ciphertext too short")
	}
	nonce, ciphertext := data[:nonceSize], data[nonceSize:]
	plaintext, err := aesgcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return "", fmt.Errorf("failed to decrypt: %w", err)
	}
	return string(plaintext), nil
}

func (m *SessionManager) CreateSession(ctx context.Context, token string, accountID string, openedDate time.Time) (string, error) {
	encrypted, err := m.encrypt(token)
	if err != nil {
		return "", fmt.Errorf("failed to encrypt token: %w", err)
	}

	session := SessionData{
		EncryptedToken: encrypted,
		AccountID:      accountID,
		OpenedDate:     openedDate,
	}
	sessionJSON, err := json.Marshal(session)
	if err != nil {
		return "", err
	}

	sessionID := uuid.New().String()
	if err := m.store.Set(ctx, "session:"+sessionID, sessionJSON); err != nil {
		return "", fmt.Errorf("failed to store session: %w", err)
	}

	return sessionID, nil
}

func (m *SessionManager) DeleteSession(ctx context.Context, sessionID string) error {
	return m.store.Delete(ctx, "session:"+sessionID)
}

func (m *SessionManager) GetSession(ctx context.Context, sessionID string) (SessionData, error) {
	sessionJSON, err := m.store.Get(ctx, "session:"+sessionID)
	if err != nil {
		return SessionData{}, err
	}

	session := SessionData{}
	err = json.Unmarshal([]byte(sessionJSON), &session)
	if err != nil {
		return SessionData{}, err
	}

	token, err := m.decrypt(session.EncryptedToken)
	if err != nil {
		return SessionData{}, err
	}
	session.EncryptedToken = token

	return session, nil
}
