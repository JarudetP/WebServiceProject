package user

import (
	"errors"

	"golang.org/x/crypto/bcrypt"
)
var jwtSecret = []byte("your_access_secret")
var refreshSecret = []byte("your_refresh_secret")

type Service struct {
	repo *Repository
}

func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) Register(req *RegisterRequest) (*User, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}
	u := &User{
		Username:     req.Username,
		Email:        req.Email,
		PasswordHash: string(hash),
		FullName:     req.FullName,
		Company:      req.Company,
	}
	return s.repo.Create(u)
}

func (s *Service) Login(req *LoginRequest) (*User, error) {
	u, err := s.repo.FindByEmail(req.Email)
	if err != nil {
		return nil, errors.New("invalid email or password")
	}
	if err := bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(req.Password)); err != nil {
		return nil, errors.New("invalid email or password")
	}
	return u, nil
}

func (s *Service) GetProfile(userID int) (*User, error) {
	return s.repo.FindByID(userID)
}

func (s *Service) TopUp(userID int, amount float64) (float64, error) {
	return s.repo.TopUp(userID, amount)
}
func (s *Service) GenerateTokenPair(user *User) (string, string, error) {
	// 1. สร้าง Access Token (อายุสั้น เช่น 15 นาที)
	accessExpiration := time.Now().Add(15 * time.Minute)
	accessClaims := &CustomClaims{
		UserID:   user.ID,
		Username: user.Username,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(accessExpiration),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    "travel-api",
		},
	}
	accessToken, err := jwt.NewWithClaims(jwt.SigningMethodHS256, accessClaims).SignedString(jwtSecret)
	if err != nil {
		return "", "", err
	}

	// 2. สร้าง Refresh Token (อายุยาว เช่น 7 วัน)
	refreshExpiration := time.Now().Add(7 * 24 * time.Hour)
	refreshClaims := &CustomClaims{
		UserID:   user.ID,
		Username: user.Username,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(refreshExpiration),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    "travel-api",
		},
	}
	refreshToken, err := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims).SignedString(refreshSecret)
	if err != nil {
		return "", "", err
	}

	return accessToken, refreshToken, nil
}
func (s *Service) RefreshToken(tokenString string) (string, string, error) {
	token, err := jwt.ParseWithClaims(tokenString, &CustomClaims{}, func(token *jwt.Token) (interface{}, error) {
		return refreshSecret, nil
	})

	if err != nil || !token.Valid {
		return "", "", errors.New("invalid refresh token")
	}

	claims, ok := token.Claims.(*CustomClaims)
	if !ok {
		return "", "", errors.New("invalid token claims")
	}

	// สร้าง user จำลองเพื่อส่งไป Gen Token ใหม่ (หรือจะ query จาก DB ใหม่ก็ได้เพื่อความชัวร์)
	user := &User{ID: claims.UserID, Username: claims.Username}
	return s.GenerateTokenPair(user)
}