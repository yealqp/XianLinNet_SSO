// Copyright 2024 OAuth Server Authors.
// Licensed under the Apache License, Version 2.0

package services

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"fmt"
	"os"
	"path/filepath"
)

const (
	RSAKeySize       = 2048
	PrivateKeyFile   = "keys/private.pem"
	PublicKeyFile    = "keys/public.pem"
	EncryptedDataDir = "data/encrypted"
)

var (
	privateKey *rsa.PrivateKey
	publicKey  *rsa.PublicKey
)

// InitRSAKeys 初始化 RSA 密钥对
func InitRSAKeys() error {
	// 确保密钥目录存在
	keyDir := filepath.Dir(PrivateKeyFile)
	if err := os.MkdirAll(keyDir, 0700); err != nil {
		return fmt.Errorf("failed to create key directory: %v", err)
	}

	// 确保加密数据目录存在
	if err := os.MkdirAll(EncryptedDataDir, 0700); err != nil {
		return fmt.Errorf("failed to create encrypted data directory: %v", err)
	}

	// 检查密钥文件是否存在
	if _, err := os.Stat(PrivateKeyFile); os.IsNotExist(err) {
		// 生成新的密钥对
		fmt.Println("[RSA] Generating new RSA key pair...")
		if err := generateRSAKeyPair(); err != nil {
			return err
		}
	}

	// 加载密钥
	if err := loadRSAKeys(); err != nil {
		return err
	}

	fmt.Println("[RSA] RSA keys initialized successfully")
	return nil
}

// generateRSAKeyPair 生成 RSA 密钥对
func generateRSAKeyPair() error {
	// 生成私钥
	privKey, err := rsa.GenerateKey(rand.Reader, RSAKeySize)
	if err != nil {
		return fmt.Errorf("failed to generate private key: %v", err)
	}

	// 保存私钥
	privKeyBytes := x509.MarshalPKCS1PrivateKey(privKey)
	privKeyPEM := &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: privKeyBytes,
	}

	privKeyFile, err := os.OpenFile(PrivateKeyFile, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return fmt.Errorf("failed to create private key file: %v", err)
	}
	defer privKeyFile.Close()

	if err := pem.Encode(privKeyFile, privKeyPEM); err != nil {
		return fmt.Errorf("failed to write private key: %v", err)
	}

	// 保存公钥
	pubKeyBytes, err := x509.MarshalPKIXPublicKey(&privKey.PublicKey)
	if err != nil {
		return fmt.Errorf("failed to marshal public key: %v", err)
	}

	pubKeyPEM := &pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: pubKeyBytes,
	}

	pubKeyFile, err := os.OpenFile(PublicKeyFile, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		return fmt.Errorf("failed to create public key file: %v", err)
	}
	defer pubKeyFile.Close()

	if err := pem.Encode(pubKeyFile, pubKeyPEM); err != nil {
		return fmt.Errorf("failed to write public key: %v", err)
	}

	fmt.Println("[RSA] RSA key pair generated and saved")
	return nil
}

// loadRSAKeys 加载 RSA 密钥
func loadRSAKeys() error {
	// 加载私钥
	privKeyData, err := os.ReadFile(PrivateKeyFile)
	if err != nil {
		return fmt.Errorf("failed to read private key: %v", err)
	}

	privKeyBlock, _ := pem.Decode(privKeyData)
	if privKeyBlock == nil {
		return fmt.Errorf("failed to decode private key PEM")
	}

	privKey, err := x509.ParsePKCS1PrivateKey(privKeyBlock.Bytes)
	if err != nil {
		return fmt.Errorf("failed to parse private key: %v", err)
	}

	privateKey = privKey

	// 加载公钥
	pubKeyData, err := os.ReadFile(PublicKeyFile)
	if err != nil {
		return fmt.Errorf("failed to read public key: %v", err)
	}

	pubKeyBlock, _ := pem.Decode(pubKeyData)
	if pubKeyBlock == nil {
		return fmt.Errorf("failed to decode public key PEM")
	}

	pubKeyInterface, err := x509.ParsePKIXPublicKey(pubKeyBlock.Bytes)
	if err != nil {
		return fmt.Errorf("failed to parse public key: %v", err)
	}

	pubKey, ok := pubKeyInterface.(*rsa.PublicKey)
	if !ok {
		return fmt.Errorf("not an RSA public key")
	}

	publicKey = pubKey

	return nil
}

// EncryptData 使用 RSA 公钥加密数据
func EncryptData(plaintext string) (string, error) {
	if publicKey == nil {
		return "", fmt.Errorf("public key not initialized")
	}

	// 使用 OAEP 填充方式加密
	ciphertext, err := rsa.EncryptOAEP(
		sha256.New(),
		rand.Reader,
		publicKey,
		[]byte(plaintext),
		nil,
	)
	if err != nil {
		return "", fmt.Errorf("encryption failed: %v", err)
	}

	// Base64 编码
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

// DecryptData 使用 RSA 私钥解密数据
func DecryptData(ciphertext string) (string, error) {
	if privateKey == nil {
		return "", fmt.Errorf("private key not initialized")
	}

	// Base64 解码
	ciphertextBytes, err := base64.StdEncoding.DecodeString(ciphertext)
	if err != nil {
		return "", fmt.Errorf("base64 decode failed: %v", err)
	}

	// 使用 OAEP 填充方式解密
	plaintext, err := rsa.DecryptOAEP(
		sha256.New(),
		rand.Reader,
		privateKey,
		ciphertextBytes,
		nil,
	)
	if err != nil {
		return "", fmt.Errorf("decryption failed: %v", err)
	}

	return string(plaintext), nil
}

// GetPublicKeyPEM 获取公钥的 PEM 格式（用于前端加密）
func GetPublicKeyPEM() (string, error) {
	if publicKey == nil {
		return "", fmt.Errorf("public key not initialized")
	}

	pubKeyBytes, err := x509.MarshalPKIXPublicKey(publicKey)
	if err != nil {
		return "", fmt.Errorf("failed to marshal public key: %v", err)
	}

	pubKeyPEM := &pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: pubKeyBytes,
	}

	return string(pem.EncodeToMemory(pubKeyPEM)), nil
}

// GetPublicKeyJWK 获取公钥的 JWK 格式（用于 OIDC JWKS 端点）
func GetPublicKeyJWK() (map[string]interface{}, error) {
	if publicKey == nil {
		return nil, fmt.Errorf("public key not initialized")
	}

	// 将 RSA 公钥的 N 和 E 转换为 Base64 URL 编码
	nBytes := publicKey.N.Bytes()
	eBytes := make([]byte, 4)
	// E 通常是 65537 (0x010001)
	eBytes[0] = byte(publicKey.E >> 24)
	eBytes[1] = byte(publicKey.E >> 16)
	eBytes[2] = byte(publicKey.E >> 8)
	eBytes[3] = byte(publicKey.E)

	// 去掉前导零
	for len(eBytes) > 1 && eBytes[0] == 0 {
		eBytes = eBytes[1:]
	}

	// Base64 URL 编码（无填充）
	n := base64.RawURLEncoding.EncodeToString(nBytes)
	e := base64.RawURLEncoding.EncodeToString(eBytes)

	jwk := map[string]interface{}{
		"kty": "RSA",         // Key Type
		"use": "sig",         // Public Key Use (signature)
		"alg": "RS256",       // Algorithm
		"kid": "default-key", // Key ID
		"n":   n,             // Modulus
		"e":   e,             // Exponent
	}

	return jwk, nil
}
