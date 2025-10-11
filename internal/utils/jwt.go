package utils

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

// JWT密钥，生产环境应从配置文件读取
var jwtSecret = []byte("ClusterResourceInsight-Secret-Key-2024")

// Claims JWT声明结构
type Claims struct {
	UserID   uint     `json:"user_id"`
	Username string   `json:"username"`
	RealName string   `json:"real_name"` // 真实姓名
	Roles    []string `json:"roles"`     // 角色代码列表
	jwt.RegisteredClaims
}

// HashPassword 使用bcrypt加密密码
func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(bytes), err
}

// CheckPassword 验证密码
func CheckPassword(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

// GenerateToken 生成JWT Token
// userID: 用户ID
// username: 用户名
// realName: 真实姓名
// roles: 用户角色代码列表
// 返回: token字符串和可能的错误
func GenerateToken(userID uint, username string, realName string, roles []string) (string, error) {
	// 设置过期时间为24小时
	expirationTime := time.Now().Add(24 * time.Hour)

	// 创建声明
	claims := &Claims{
		UserID:   userID,
		Username: username,
		RealName: realName,
		Roles:    roles,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    "ClusterResourceInsight",
		},
	}

	// 创建token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// 签名token
	tokenString, err := token.SignedString(jwtSecret)
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

// ParseToken 解析JWT Token
// tokenString: token字符串
// 返回: Claims和可能的错误
func ParseToken(tokenString string) (*Claims, error) {
	// 解析token
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		return jwtSecret, nil
	})

	if err != nil {
		return nil, err
	}

	// 验证token并提取claims
	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims, nil
	}

	return nil, errors.New("无效的token")
}

// RefreshToken 刷新Token
// 如果token即将过期(小于2小时)，生成新token
func RefreshToken(tokenString string) (string, error) {
	claims, err := ParseToken(tokenString)
	if err != nil {
		return "", err
	}

	// 如果token还有超过2小时才过期，不需要刷新
	if time.Until(claims.ExpiresAt.Time) > 2*time.Hour {
		return tokenString, nil
	}

	// 生成新token
	return GenerateToken(claims.UserID, claims.Username, claims.RealName, claims.Roles)
}