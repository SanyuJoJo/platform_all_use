package security
import (
	"errors"
	"time"
	"github.com/golang-jwt/jwt/v5"
)
// JWTClaims 自定义 JWT 载荷。
//
// v1.1（P1-02）修复：
//   移除自定义的 Sub 字段。`jwt.RegisteredClaims` 已包含
//   `Subject string json:"sub,omitempty"`，两者 JSON 名冲突会导致
//   编码/解码行为不确定。现统一使用 `RegisteredClaims.Subject`。
type JWTClaims struct {
	Username    string   `json:"username,omitempty"`
	Roles       []string `json:"roles,omitempty"`
	Permissions []string `json:"permissions,omitempty"`
	Type        string   `json:"type"`
	jwt.RegisteredClaims
}
// GenerateToken 签发 JWT。
//
// v1.1（P1-02）：新增 subject 参数，由调用方显式传入，
//   内部写入 `RegisteredClaims.Subject`，避免与自定义字段冲突。
func GenerateToken(secret, subject string, claims JWTClaims, expiresAt time.Time) (string, error) {
	claims.RegisteredClaims = jwt.RegisteredClaims{
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
