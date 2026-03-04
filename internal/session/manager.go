package session

/*
type Manager struct {
	store  *Store
	secret []byte
}

func NewManager(store *Store, secret string) (*Manager, error) {
	key, err := hex.DecodeString(secret)
	if err != nil || len(key) != 32 {
		return nil, fmt.Errorf("SESSION_SECRET must be a 32-byte hex-encoded string")
	}
	return &Manager{store: store, secret: key}, nil
}

func (m *Manager) encrypt(plaintext string) (string, error) {
	block, err := aes.NewCipher(m.secret)
	if err != nil {
		return "", nil
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", nil
	}
	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", nil
	}
	ciphertext := gcm.Seal(nonce, nonce, []byte(plaintext), nil)
	return hex.EncodeToString(ciphertext), nil
}

func (m *Manager) decrypt(ciphertextHex string) (string, error) {
	data, err := hex.DecodeString(ciphertextHex)
	if err != nil {
		return "", err
	}
	block, err := aes.NewCipher(m.secret)
	if err != nil {
		return "", nil
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", nil
	}
	nonceSize := gcm.NonceSize()
	if len(data) < nonceSize {
		return "", fmt.Errorf("ciphertext too short")
	}
	nonce, ciphertext := data[:nonceSize], data[nonceSize:]
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return "", fmt.Errorf("failed to decrypt: %w", err)
	}

	return string(plaintext), nil
}

func (m *Manager) CreateSession(ctx context.Context, token string) (string, error) {
	encrypted, err := m.encrypt(token)
	if err != nil {
		return "", fmt.Errorf("failed to encrypt token: %w", err)
	}

	sessionID := uuid.New().String()
	if err := m.store.Set(ctx, sessionID, encrypted); err != nil {
		return "", fmt.Errorf("failed to store session: %w", err)
	}

	return sessionID, nil
}

func (m *Manager) DeleteSession(ctx context.Context, sessionID string) error {
	return m.store.Delete(ctx, sessionID)
}

func (m *Manager) GetToken(ctx context.Context, sessionID string) (string, error) {
	encrypted, err := m.store.Get(ctx, sessionID)
	if err != nil {
		return "", err
	}
	token, err := m.decrypt(encrypted)
	if err != nil {
		return "", err
	}
	return token, nil
}
*/
