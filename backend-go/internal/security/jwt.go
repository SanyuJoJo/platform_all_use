package security

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

// JWTClaims 自定义 JWT 载荷。
//
// v1.1（P1-02）：移除自定义 Sub 字段，统一使用 RegisteredClaims.Subject。
//
// v1.4：GenerateToken 自动注入 RegisteredClaims.ID（jti），
// 保证每次签发的 token 唯一。此前缺少 jti 时，
// 同一秒内为同一用户签发 refresh token 会得到完全相同的 JWT
// （sub/exp/iat/type 全相同），token_hash 相同导致
// auth_refresh_token.token_hash 唯一约束冲突。
type JWTClaims struct {
	Username    string   `json:"username,omitempty"`
	Roles       []string `json:"roles,omitempty"`
	Permissions []string `json:"permissions,omitempty"`
	Type        string   `json:"type"`
	jwt.RegisteredClaims
}

// GenerateToken 签发 JWT。
//
// 每次调用都会生成一个随机的 jti（RegisteredClaims.ID），
// 确保即使 sub/exp/iat/type 完全相同的两次签发，
// 得到的 token 字符串也不同（避免 refresh_token.token_hash 冲突）。
func GenerateToken(secret, subject string, claims JWTClaims, expiresAt time.Time) (string, error) {
	claims.RegisteredClaims = jwt.RegisteredClaims{
		ID:        uuid.NewString(), // ★ jti：每次签发唯一
		Subject:   subject,
		ExpiresAt: jwt.NewNumericDate(expiresAt),
		IssuedAt:  jwt.NewNumericDate(time.Now()),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secret))
}

// ParseToken 解析并验证 JWT。
// 调用方通过 `claims.Subject` 读取用户 ID。
func ParseToken(secret, tokenString string) (*JWTClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return []byte(secret), nil
	})
	if err != nil {
		return nil, err
	}
	claims, ok := token.Claims.(*JWTClaims)
	if !ok || !token.Valid {
		return nil, errors.New("invalid token")
	}
	return claims, nil
}
