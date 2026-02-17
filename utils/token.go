package utils

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/leleo886/lopic/internal/config"
	cerrors "github.com/leleo886/lopic/internal/error"
	"github.com/leleo886/lopic/internal/log"
)

// HMAC 签名令牌

func GenerateSignedToken(email string, timestamp int64, prefix string, cfg *config.JWTConfig) string {
	data := fmt.Sprintf("%s:%s:%d", prefix, email, timestamp)
	signature := generateSignature(data, cfg.TokenSecret)
	return fmt.Sprintf("%s:%s", data, signature)
}

func ValidateSignedToken(token string, prefix string, cfg *config.JWTConfig) (string, int64, error) {
	parts := strings.Split(token, ":")
	if len(parts) < 3 {
		log.Errorf("invalid token format: token=%s", token)
		return "", 0, cerrors.ErrForbidden
	}

	prefixCheck := parts[0]
	if prefixCheck != prefix {
		log.Errorf("invalid token prefix: token=%s", token)
		return "", 0, cerrors.ErrForbidden
	}

	email := parts[1]
	timestampStr := parts[2]
	signature := parts[3]

	timestamp, err := strconv.ParseInt(timestampStr, 10, 64)
	if err != nil {
		log.Errorf("invalid timestamp: token=%s", token)
		return "", 0, cerrors.ErrForbidden
	}

	data := fmt.Sprintf("%s:%s:%d", prefix, email, timestamp)
	expectedSignature := generateSignature(data, cfg.TokenSecret)

	if signature != expectedSignature {
		log.Errorf("invalid signature: token=%s", token)
		return "", 0, cerrors.ErrForbidden
	}

	return email, timestamp, nil
}

func IsTokenExpired(timestamp int64, expirationHours int) bool {
	expirationTime := time.Unix(timestamp, 0).Add(time.Duration(expirationHours) * time.Hour)
	return time.Now().After(expirationTime)
}

func generateSignature(data string, secret string) string {
	h := hmac.New(sha256.New, []byte(secret))
	h.Write([]byte(data))
	// 使用 URL 安全的 Base64 编码，避免 + 和 / 在 URL 中引起问题
	return base64.URLEncoding.EncodeToString(h.Sum(nil))
}
