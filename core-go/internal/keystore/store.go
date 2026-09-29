package keystore
import (
    "crypto/aes"
    "crypto/cipher"
    "crypto/rand"
    "encoding/hex"
    "errors"
    "io"
    "strings"
)
// Store AES-256-GCM 密钥加解密。
//
// 文件格式（与旧 keycrypt.c 二进制兼容）：
//   offset 0      : magic "KC01" (4 bytes)
//   offset 4      : IV           (12 bytes)
//   offset 16     : ciphertext   (N bytes)
//   offset 16 + N : GCM tag      (16 bytes，由 Seal 自动附加)
type Store struct {
    masterKey []byte
}
const (
    magicBytes = "KC01"
    ivLen      = 12
    tagLen     = 16
    keyLen     = 32
)
// NewStore 用 32 字节主密钥创建 Store。
func NewStore(masterKeyHex string) (*Store, error) {
    key, err := hex.DecodeString(strings.TrimSpace(masterKeyHex))
    if err != nil || len(key) != keyLen {
        return nil, errors.New("master key must be 64 hex chars (32 bytes)")
    }
    return &Store{masterKey: key}, nil
}
// Encrypt 加密明文。
func (s *Store) Encrypt(plaintext []byte) ([]byte, error) {
    block, err := aes.NewCipher(s.masterKey)
    if err != nil {
        return nil, err
    }
    gcm, err := cipher.NewGCM(block)
    if err != nil {
        return nil, err
    }
    iv := make([]byte, ivLen)
    if _, err := io.ReadFull(rand.Reader, iv); err != nil {
        return nil, err
    }
    ct := gcm.Seal(nil, iv, plaintext, nil)
    out := make([]byte, 0, len(magicBytes)+ivLen+len(ct))
    out = append(out, []byte(magicBytes)...)
    out = append(out, iv...)
    out = append(out, ct...)
    return out, nil
}
// Decrypt 解密。
func (s *Store) Decrypt(data []byte) ([]byte, error) {
    if len(data) < len(magicBytes)+ivLen+tagLen {
        return nil, errors.New("ciphertext too short")
    }
    if string(data[:len(magicBytes)]) != magicBytes {
        return nil, errors.New("bad magic, not a keycrypt file")
    }
    iv := data[len(magicBytes) : len(magicBytes)+ivLen]
    ct := data[len(magicBytes)+ivLen:]
    block, err := aes.NewCipher(s.masterKey)
    if err != nil {
        return nil, err
    }
    gcm, err := cipher.NewGCM(block)
    if err != nil {
        return nil, err
    }
    return gcm.Open(nil, iv, ct, nil)
}
