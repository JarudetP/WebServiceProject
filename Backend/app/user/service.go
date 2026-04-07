package user

import (
	"errors"
	"time"
	"os"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)
func getSecret() []byte{
	return []byte(os.Getenv("JWT_SECRET"))
}
func getRefresh() []byte{
	return  []byte(os.Getenv("REFRESH_SECRET"))
}


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
	accessToken, err := jwt.NewWithClaims(jwt.SigningMethodHS256, accessClaims).SignedString(getSecret())
	if err != nil {
		return "", "", err
	}

	refreshExpiration := time.Now().Add(2 * 24 * time.Hour)
	refreshClaims := &CustomClaims{
		UserID:   user.ID,
		Username: user.Username,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(refreshExpiration),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    "travel-api",
		},
	}
	refreshToken, err := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims).SignedString(getRefresh())
	if err != nil {
		return "", "", err
	}

	return accessToken, refreshToken, nil
}
func (s *Service) RefreshToken(tokenString string) (string, string, error) {
	token, err := jwt.ParseWithClaims(tokenString, &CustomClaims{}, func(token *jwt.Token) (interface{}, error) {
		return getRefresh(), nil
	})

	if err != nil || !token.Valid {
		return "", "", errors.New("invalid refresh token")
	}

	claims, ok := token.Claims.(*CustomClaims)
	if !ok {
		return "", "", errors.New("invalid token claims")
	}

	user,err := s.repo.FindByID(claims.UserID)
	if err != nil {
		return "", "", errors.New("User Not Found")
	}
	return s.GenerateTokenPair(user)
}