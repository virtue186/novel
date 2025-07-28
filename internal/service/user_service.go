package service

import (
	"errors"
	"fmt"
	"github.com/golang-jwt/jwt/v5"
	"github.com/novel/internal/model"
	"github.com/novel/internal/pkg/config"
	"github.com/novel/internal/repository"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
	"time"
)

// 定义一些业务相关的错误，方便上层进行判断
var (
	ErrUserAlreadyExists  = errors.New("username is already registered")
	ErrInvalidCredentials = errors.New("invalid username or password")
	ErrInvalidToken       = errors.New("invalid or expired token")
)

// UserService 定义了用户认证相关的核心业务逻辑接口
type UserService interface {
	Register(username, password string) (*model.User, error)
	Login(username, password string) (string, error)
	ParseToken(tokenString string) (uint, error)
}

// userService 结构体实现了 UserService 接口
type userService struct {
	repo   repository.UserRepository
	jwtCfg *config.JWTConfig
}

// NewUserService 是 userService 的构造函数，负责依赖注入
func NewUserService(repo repository.UserRepository, jwtCfg *config.JWTConfig) UserService {
	return &userService{
		repo:   repo,
		jwtCfg: jwtCfg,
	}
}

// Register 负责处理用户注册逻辑
func (s *userService) Register(username, password string) (*model.User, error) {
	// 1. 业务校验：检查用户名是否已被占用
	_, err := s.repo.FindByUsername(username)
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		// 如果错误不是 "未找到"，说明发生了其他错误或用户已存在
		if err == nil {
			return nil, ErrUserAlreadyExists
		}
		return nil, err // 其他数据库错误
	}

	// 2. 密码加密：使用 bcrypt 对密码进行安全的哈希处理
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	// 3. 创建用户模型
	user := &model.User{
		Username:     username,
		PasswordHash: string(hashedPassword),
	}

	// 4. 持久化到数据库
	if err := s.repo.Create(user); err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	// 5. 返回用户信息时，清除敏感的密码哈希值
	user.PasswordHash = ""
	return user, nil
}

// Login 负责处理用户登录逻辑
func (s *userService) Login(username, password string) (string, error) {
	// 1. 根据用户名查找用户
	user, err := s.repo.FindByUsername(username)
	if err != nil {
		// 统一返回“无效凭证”错误，避免泄露“用户不存在”的信息
		return "", ErrInvalidCredentials
	}

	// 2. 验证密码哈希与提供的密码是否匹配
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		// 密码不匹配
		return "", ErrInvalidCredentials
	}

	// 3. 创建 JWT (JSON Web Token)
	duration, err := time.ParseDuration(s.jwtCfg.ExpiryTime)
	if err != nil {
		return "", errors.New("系统配置的过期时间无效")
	}
	claims := jwt.MapClaims{
		"user_id": user.ID,
		"exp":     time.Now().Add(duration).Unix(),
		"iat":     time.Now().Unix(),
	}

	// 使用 HS256 签名算法创建一个新的 Token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// 使用我们在配置中定义的密钥，对 Token 进行签名，生成最终的 Token 字符串
	tokenString, err := token.SignedString([]byte(s.jwtCfg.SecretKey))
	if err != nil {
		return "", fmt.Errorf("failed to sign token: %w", err)
	}

	return tokenString, nil
}

// ParseToken 负责解析和验证 JWT
func (s *userService) ParseToken(tokenString string) (uint, error) {
	// 1. 解析 Token 字符串
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		// 2. 校验签名算法是否是我们预期的 HS256
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		// 3. 返回我们的密钥供库进行签名验证
		return []byte(s.jwtCfg.SecretKey), nil
	})

	if err != nil {
		return 0, ErrInvalidToken // 解析或签名验证失败
	}

	// 4. 验证 Token 是否有效，并提取 Claims
	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		// 5. 从 Claims 中获取 user_id
		userIDFloat, ok := claims["user_id"].(float64)
		if !ok {
			return 0, ErrInvalidToken // user_id 类型不正确
		}
		return uint(userIDFloat), nil
	}

	return 0, ErrInvalidToken
}
