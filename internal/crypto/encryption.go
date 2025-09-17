package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"io"
)

const (
	// EncryptionKey AES加密密钥，在生产环境中应该从环境变量或配置文件中读取
	// 这里为了演示使用固定密钥，实际部署时需要更改
	DefaultEncryptionKey = "myclustermonitor32bytesecretkey!!" // 32字节密钥用于AES-256
)

// EncryptionService 加密服务结构
type EncryptionService struct {
	key []byte
}

// NewEncryptionService 创建新的加密服务实例
func NewEncryptionService(key string) *EncryptionService {
	if key == "" {
		key = DefaultEncryptionKey
	}
	
	// 确保密钥长度为32字节（AES-256）
	keyBytes := []byte(key)
	if len(keyBytes) > 32 {
		keyBytes = keyBytes[:32]
	} else if len(keyBytes) < 32 {
		// 如果密钥长度不足32字节，用0填充
		for len(keyBytes) < 32 {
			keyBytes = append(keyBytes, 0)
		}
	}
	
	return &EncryptionService{
		key: keyBytes,
	}
}

// Encrypt 加密字符串数据
func (e *EncryptionService) Encrypt(plaintext string) (string, error) {
	if plaintext == "" {
		return "", nil
	}

	// 创建AES加密器
	block, err := aes.NewCipher(e.key)
	if err != nil {
		return "", err
	}

	// 创建GCM模式
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	// 生成随机nonce
	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}

	// 加密数据
	ciphertext := gcm.Seal(nonce, nonce, []byte(plaintext), nil)

	// 返回base64编码的结果
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

// Decrypt 解密字符串数据
func (e *EncryptionService) Decrypt(ciphertext string) (string, error) {
	if ciphertext == "" {
		return "", nil
	}

	// 解码base64
	data, err := base64.StdEncoding.DecodeString(ciphertext)
	if err != nil {
		return "", err
	}

	// 创建AES解密器
	block, err := aes.NewCipher(e.key)
	if err != nil {
		return "", err
	}

	// 创建GCM模式
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	// 检查数据长度
	nonceSize := gcm.NonceSize()
	if len(data) < nonceSize {
		return "", errors.New("密文数据长度不足")
	}

	// 分离nonce和密文
	nonce, cipherBytes := data[:nonceSize], data[nonceSize:]

	// 解密数据
	plaintext, err := gcm.Open(nil, nonce, cipherBytes, nil)
	if err != nil {
		return "", err
	}

	return string(plaintext), nil
}

// 全局加密服务实例
var defaultEncryptionService *EncryptionService

// 初始化默认加密服务
func init() {
	defaultEncryptionService = NewEncryptionService(DefaultEncryptionKey)
}

// EncryptData 使用默认加密服务加密数据
func EncryptData(plaintext string) (string, error) {
	return defaultEncryptionService.Encrypt(plaintext)
}

// DecryptData 使用默认加密服务解密数据  
func DecryptData(ciphertext string) (string, error) {
	return defaultEncryptionService.Decrypt(ciphertext)
}