package auth

import (
	"context"
	"errors"
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/golang-jwt/jwt/v4"
)

// 上下文key类型
type contextKey string

// 用户上下文key
const UserKey contextKey = "user"

var (
	secretKey = os.Getenv("JWT_SECRET")
	issuer    = os.Getenv("JWT_ISSUER")
)

var (
	// 错误定义
	ErrInvalidToken     = errors.New("invalid token")
	ErrExpiredToken     = errors.New("token expired")
	ErrTokenNotProvided = errors.New("token not provided")
	ErrInvalidSignature = errors.New("invalid signature")
	ErrUnsupportedAlg   = errors.New("unsupported algorithm")
)

// TokenType 定义不同类型的token
type TokenType string

const (
	AccessToken  TokenType = "access"
	RefreshToken TokenType = "refresh"
)

// JWTConfig 定义JWT配置
type JWTConfig struct {
	// 签名密钥
	SecretKey []byte
	// 签名方法
	SigningMethod jwt.SigningMethod
	// 颁发者
	Issuer string
	// Token有效期，默认为24小时
	AccessTokenDuration time.Duration
	// 刷新Token有效期，默认为7天
	RefreshTokenDuration time.Duration
}

// DefaultConfig 创建默认配置
func DefaultConfig() JWTConfig {
	return JWTConfig{
		SecretKey:            []byte(secretKey),
		SigningMethod:        jwt.SigningMethodHS256,
		Issuer:               issuer,
		AccessTokenDuration:  time.Hour * 24,
		RefreshTokenDuration: time.Hour * 24 * 7,
	}
}

// JWTService JWT服务接口
type JWTService interface {
	// GenerateToken 生成JWT令牌
	GenerateToken(claims CustomClaims) (string, error)
	// GenerateTokenPair 生成访问令牌和刷新令牌对
	GenerateTokenPair(claims CustomClaims) (accessToken string, refreshToken string, err error)
	// ValidateToken 验证JWT令牌
	ValidateToken(tokenString string) (*CustomClaims, error)
	// RefreshToken 使用刷新令牌生成新的访问令牌
	RefreshToken(refreshTokenString string) (newAccessToken string, err error)
	// ExtractTokenFromHeader 从Authorization头中提取令牌
	ExtractTokenFromHeader(authHeader string) (string, error)
}

// CustomClaims 自定义JWT声明
type CustomClaims struct {
	// 用户唯一标识
	UserID string `json:"user_id"`
	// 用户名
	Username string `json:"username"`
	// 角色
	Role string `json:"role"`
	// 其他数据(可选)
	ExtraData map[string]interface{} `json:"extra_data,omitempty"`
	// 标准声明
	jwt.RegisteredClaims
}

// jwtService JWT服务实现
type jwtService struct {
	config JWTConfig
	mu     sync.RWMutex
}

// NewJWTService 创建新的JWT服务
func NewJWTService(config JWTConfig) JWTService {
	return &jwtService{
		config: config,
	}
}

// NewDefaultJWTService 使用默认配置创建JWT服务
func NewDefaultJWTService() JWTService {
	return NewJWTService(DefaultConfig())
}

// GenerateToken 生成JWT令牌
func (j *jwtService) GenerateToken(claims CustomClaims) (string, error) {
	j.mu.RLock()
	defer j.mu.RUnlock()

	// 使用自定义声明创建JWT
	token := jwt.NewWithClaims(j.config.SigningMethod, claims)

	// 签名并获取完整的编码后的字符串
	tokenString, err := token.SignedString(j.config.SecretKey)
	if err != nil {
		return "", fmt.Errorf("generate token failed: %w", err)
	}

	return tokenString, nil
}

// GenerateTokenPair 生成访问令牌和刷新令牌对
func (j *jwtService) GenerateTokenPair(claims CustomClaims) (string, string, error) {
	j.mu.RLock()
	defer j.mu.RUnlock()

	// 准备访问令牌声明
	accessClaims := claims
	accessClaims.RegisteredClaims.ExpiresAt = jwt.NewNumericDate(time.Now().Add(j.config.AccessTokenDuration))
	accessClaims.RegisteredClaims.IssuedAt = jwt.NewNumericDate(time.Now())
	accessClaims.RegisteredClaims.NotBefore = jwt.NewNumericDate(time.Now())
	accessClaims.RegisteredClaims.Issuer = j.config.Issuer

	// 生成访问令牌
	accessToken, err := j.GenerateToken(accessClaims)
	if err != nil {
		return "", "", err
	}

	// 准备刷新令牌声明 (通常包含较少信息)
	refreshClaims := CustomClaims{
		UserID: claims.UserID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(j.config.RefreshTokenDuration)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    j.config.Issuer,
			Subject:   claims.UserID,
		},
	}

	// 生成刷新令牌
	refreshToken, err := j.GenerateToken(refreshClaims)
	if err != nil {
		return "", "", err
	}

	return accessToken, refreshToken, nil
}

// ValidateToken 验证JWT令牌
func (j *jwtService) ValidateToken(tokenString string) (*CustomClaims, error) {
	j.mu.RLock()
	defer j.mu.RUnlock()

	if tokenString == "" {
		return nil, ErrTokenNotProvided
	}

	// 解析令牌
	token, err := jwt.ParseWithClaims(
		tokenString,
		&CustomClaims{},
		func(token *jwt.Token) (interface{}, error) {
			// 验证签名算法
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, ErrUnsupportedAlg
			}
			return j.config.SecretKey, nil
		},
	)

	if err != nil {
		ve, ok := err.(*jwt.ValidationError)
		if ok {
			if ve.Errors&jwt.ValidationErrorExpired != 0 {
				return nil, ErrExpiredToken
			} else if ve.Errors&(jwt.ValidationErrorSignatureInvalid) != 0 {
				return nil, ErrInvalidSignature
			}
		}
		return nil, ErrInvalidToken
	}

	claims, ok := token.Claims.(*CustomClaims)
	if !ok || !token.Valid {
		return nil, ErrInvalidToken
	}

	return claims, nil
}

// RefreshToken 使用刷新令牌生成新的访问令牌
func (j *jwtService) RefreshToken(refreshTokenString string) (string, error) {
	// 验证刷新令牌
	refreshClaims, err := j.ValidateToken(refreshTokenString)
	if err != nil {
		return "", err
	}

	// 创建新的访问令牌
	accessClaims := CustomClaims{
		UserID:   refreshClaims.UserID,
		Username: refreshClaims.Username,
		Role:     refreshClaims.Role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(j.config.AccessTokenDuration)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    j.config.Issuer,
			Subject:   refreshClaims.UserID,
		},
	}

	// 如果有额外数据，也一并复制
	if refreshClaims.ExtraData != nil {
		accessClaims.ExtraData = refreshClaims.ExtraData
	}

	return j.GenerateToken(accessClaims)
}

// ExtractTokenFromHeader 从Authorization头中提取令牌
func (j *jwtService) ExtractTokenFromHeader(authHeader string) (string, error) {
	if authHeader == "" {
		return "", ErrTokenNotProvided
	}

	// 检查格式是否为 "Bearer <token>"
	if len(authHeader) > 7 && authHeader[:7] == "Bearer " {
		return authHeader[7:], nil
	}

	return "", ErrInvalidToken
}

// WithUser 将用户信息添加到上下文中
func WithUser(ctx context.Context, claims *CustomClaims) context.Context {
	return context.WithValue(ctx, UserKey, claims)
}

// FromContext 从上下文中获取用户信息
func FromContext(ctx context.Context) (*CustomClaims, bool) {
	claims, ok := ctx.Value(UserKey).(*CustomClaims)
	return claims, ok
}

func GetSecretKey() string {
	return secretKey
}

func GetIssuer() string {
	return issuer
}
