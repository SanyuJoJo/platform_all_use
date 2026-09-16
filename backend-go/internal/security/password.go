package security
import (
	"golang.org/x/crypto/bcrypt"
)
// HashPassword 使用 bcrypt 哈希密码。
// 与 Python bcrypt 兼容，生成 $2b$12$... 格式。
func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}
// VerifyPassword 验证明文密码与哈希是否匹配。
func VerifyPassword(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}
