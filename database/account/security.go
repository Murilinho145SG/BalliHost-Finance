package account

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"io"
	"os"
	"prodata/logs"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

func Encrypt(text string) string {
	logger := logs.NewSistemLogger()

	block, err := aes.NewCipher([]byte(os.Getenv("CRIKEY_ACC")))
	if err != nil {
		logger.LogAndSendSystemMessage(err.Error())
		return ""
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		logger.LogAndSendSystemMessage(err.Error())
		return ""
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		logger.LogAndSendSystemMessage(err.Error())
		return ""
	}

	encrypted := gcm.Seal(nonce, nonce, []byte(text), nil)
	return base64.StdEncoding.EncodeToString(encrypted)
}

func Decrypt(encrypted string) string {
	logger := logs.NewSistemLogger()

	data, err := base64.StdEncoding.DecodeString(encrypted)
	if err != nil {
		logger.LogAndSendSystemMessage(err.Error())
		return ""
	}

	block, err := aes.NewCipher([]byte(os.Getenv("CRIKEY_ACC")))
	if err != nil {
		logger.LogAndSendSystemMessage(err.Error())
		return ""
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		logger.LogAndSendSystemMessage(err.Error())
		return ""
	}

	nonceSize := gcm.NonceSize()
	nonce, ciphertext := data[:nonceSize], data[nonceSize:]

	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		logger.LogAndSendSystemMessage(err.Error())
		return ""
	}

	return string(plaintext)
}

func GenerateJWT(userID, deviceId string) string {
	claims := jwt.MapClaims{
		"userId":   userID,
		"DeviceId": deviceId,
		"exp":      time.Now().Add(7 * 24 * time.Hour).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(os.Getenv("JWTKEY")))
	if err != nil {
		logs.NewSistemLogger().LogAndSendSystemMessage(err.Error())
		return ""
	}
	return tokenString
}

func GenerateJWTRole(userID, deviceId string) string {
	claims := jwt.MapClaims{
		"userId":   userID,
		"DeviceId": deviceId,
		"exp":      time.Now().Add(7 * 24 * time.Hour).Unix(),
		"admin":    true,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(os.Getenv("JWTKEY")))
	if err != nil {
		logs.NewSistemLogger().LogAndSendSystemMessage(err.Error())
		return ""
	}
	return tokenString
}

func HashPassword(password string) string {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		logs.NewSistemLogger().LogAndSendSystemMessage(err.Error())
		return ""
	}

	return string(hash)
}

func ComparePassword(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}
