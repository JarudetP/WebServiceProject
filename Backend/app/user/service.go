package user

import (
	"errors"

	"golang.org/x/crypto/bcrypt"
)

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
