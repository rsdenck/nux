package skill

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"golang.org/x/crypto/argon2"
)

const (
	VaultDir    = ".skills"
	VaultFile   = ".nux.json"
	saltLen     = 32
	nonceLen    = 12
	keyLen      = 32 // 256 bits for AES-256
)

// EncryptedVault represents an encrypted vault file
type EncryptedVault struct {
	Version string `json:"version"`
	Salt    string `json:"salt"`
	Nonce   string `json:"nonce"`
	Data    string `json:"data"`
}

// Vault represents the unencrypted vault structure
type Vault struct {
	Version        string            `json:"version"`
	InstalledSkills []string         `json:"installed_skills"`
	EnabledSkills   []string         `json:"enabled_skills"`
	APIKeys         map[string]string `json:"api_keys"`
	Ollama          OllamaConfig     `json:"ollama"`
	VaultMode       bool             `json:"vault_mode"`
}

// OllamaConfig represents Ollama configuration
type OllamaConfig struct {
	Host    string `json:"host"`
	Model   string `json:"model"`
	Enabled bool   `json:"enabled"`
}

// GetVaultPath returns the path to the vault file
func GetVaultPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get home directory: %w", err)
	}
	return filepath.Join(home, VaultDir, VaultFile), nil
}

// DeriveKey derives a key from password using Argon2id
func DeriveKey(password, salt []byte) []byte {
	return argon2.IDKey(password, salt, 1, 64*1024, 4, keyLen)
}

// Encrypt encrypts data with AES-256-GCM
func Encrypt(data []byte, password []byte) (*EncryptedVault, error) {
	// Generate random salt
	salt := make([]byte, saltLen)
	if _, err := io.ReadFull(rand.Reader, salt); err != nil {
		return nil, fmt.Errorf("failed to generate salt: %w", err)
	}

	// Derive key from password
	key := DeriveKey(password, salt)

	// Generate random nonce
	nonce := make([]byte, nonceLen)
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, fmt.Errorf("failed to generate nonce: %w", err)
	}

	// Create AES cipher
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("failed to create cipher: %w", err)
	}

	// Create GCM mode
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("failed to create GCM: %w", err)
	}

	// Encrypt data
	ciphertext := gcm.Seal(nil, nonce, data, nil)

	return &EncryptedVault{
		Version: "1.0",
		Salt:    base64.StdEncoding.EncodeToString(salt),
		Nonce:   base64.StdEncoding.EncodeToString(nonce),
		Data:    base64.StdEncoding.EncodeToString(ciphertext),
	}, nil
}

// Decrypt decrypts data with AES-256-GCM
func Decrypt(encrypted *EncryptedVault, password []byte) ([]byte, error) {
	// Decode salt
	salt, err := base64.StdEncoding.DecodeString(encrypted.Salt)
	if err != nil {
		return nil, fmt.Errorf("failed to decode salt: %w", err)
	}

	// Derive key from password
	key := DeriveKey(password, salt)

	// Decode nonce
	nonce, err := base64.StdEncoding.DecodeString(encrypted.Nonce)
	if err != nil {
		return nil, fmt.Errorf("failed to decode nonce: %w", err)
	}

	// Decode ciphertext
	ciphertext, err := base64.StdEncoding.DecodeString(encrypted.Data)
	if err != nil {
		return nil, fmt.Errorf("failed to decode data: %w", err)
	}

	// Create AES cipher
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("failed to create cipher: %w", err)
	}

	// Create GCM mode
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("failed to create GCM: %w", err)
	}

	// Decrypt data
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt: %w", err)
	}

	return plaintext, nil
}

// LoadVault loads and decrypts the vault
// If password is empty, tries to load as plaintext (backward compatibility)
func LoadVault(password ...string) (*Vault, error) {
	path, err := GetVaultPath()
	if err != nil {
		return nil, err
	}

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return defaultVault(), nil
		}
		return nil, fmt.Errorf("failed to read vault: %w", err)
	}

	// Try to parse as encrypted vault first
	var encrypted EncryptedVault
	if err := json.Unmarshal(data, &encrypted); err == nil && encrypted.Version != "" && encrypted.Data != "" {
		// It's an encrypted vault
		if len(password) == 0 {
			return nil, fmt.Errorf("vault is encrypted, password required")
		}
		plaintext, err := Decrypt(&encrypted, []byte(password[0]))
		if err != nil {
			return nil, fmt.Errorf("failed to decrypt vault: %w", err)
		}
		var v Vault
		if err := json.Unmarshal(plaintext, &v); err != nil {
			return nil, fmt.Errorf("failed to parse vault: %w", err)
		}
		return &v, nil
	}

	// Try plaintext (backward compatibility)
	var v Vault
	if err := json.Unmarshal(data, &v); err != nil {
		return nil, fmt.Errorf("failed to parse vault: %w", err)
	}
	return &v, nil
}

// SaveVault saves and optionally encrypts the vault
// If password is provided, encrypts the vault
func SaveVault(v *Vault, password ...string) error {
	path, err := GetVaultPath()
	if err != nil {
		return err
	}

	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0700); err != nil {
		return fmt.Errorf("failed to create vault directory: %w", err)
	}

	var data []byte
	if len(password) > 0 && password[0] != "" {
		// Encrypt the vault
		jsonData, err := json.Marshal(v)
		if err != nil {
			return fmt.Errorf("failed to marshal vault: %w", err)
		}

		encrypted, err := Encrypt(jsonData, []byte(password[0]))
		if err != nil {
			return fmt.Errorf("failed to encrypt vault: %w", err)
		}

		data, err = json.MarshalIndent(encrypted, "", "  ")
		if err != nil {
			return fmt.Errorf("failed to marshal encrypted vault: %w", err)
		}
	} else {
		// Save as plaintext (backward compatibility)
		data, err = json.MarshalIndent(v, "", "  ")
		if err != nil {
			return fmt.Errorf("failed to marshal vault: %w", err)
		}
	}

	// G301: Expect file permissions to be 0600 or less
	if err := os.WriteFile(path, data, 0600); err != nil {
		return fmt.Errorf("failed to write vault: %w", err)
	}
	return nil
}

// VerifyPassword verifies if a password can decrypt the vault
func VerifyPassword(password string) error {
	path, err := GetVaultPath()
	if err != nil {
		return err
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("failed to read vault: %w", err)
	}

	var encrypted EncryptedVault
	if err := json.Unmarshal(data, &encrypted); err != nil {
		return fmt.Errorf("vault is not encrypted or invalid format")
	}

	// Try to decrypt with provided password
	_, err = Decrypt(&encrypted, []byte(password))
	return err
}

// ChangePassword changes the vault password
func ChangePassword(oldPassword, newPassword string) error {
	// Load with old password
	v, err := LoadVault(oldPassword)
	if err != nil {
		return fmt.Errorf("failed to load vault: %w", err)
	}

	// Save with new password
	return SaveVault(v, newPassword)
}

// defaultVault returns a new default vault
func defaultVault() *Vault {
	return &Vault{
		Version:         "1.0.0",
		InstalledSkills: []string{},
		EnabledSkills:   []string{},
		APIKeys:         make(map[string]string),
		Ollama: OllamaConfig{
			Host:    "http://localhost:11434",
			Model:   "qwen3-coder",
			Enabled: false,
		},
		VaultMode: true,
	}
}

// HashPassword creates a secure hash of a password for storage
func HashPassword(password string) string {
	hash := sha256.Sum256([]byte(password))
	return base64.StdEncoding.EncodeToString(hash[:])
}

// CompareHash compares a password with its hash
func CompareHash(password, hash string) bool {
	return HashPassword(password) == hash
}