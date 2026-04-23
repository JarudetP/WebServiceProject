package pkg

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"
)

type Service struct {
	repo     *Repository
	userSvcURL string
}

func NewService(repo *Repository, userSvcURL string) *Service {
	return &Service{repo: repo, userSvcURL: userSvcURL}
}

func (s *Service) ListPackages() ([]Package, error) {
	return s.repo.ListPackages()
}

func (s *Service) GetPackage(id int) (*Package, error) {
	return s.repo.FindPackageByID(id)
}

func (s *Service) Purchase(userID, packageID int) (*Subscription, error) {
	newPkg, err := s.repo.FindPackageByID(packageID)
	if err != nil {
		return nil, err
	}

	activeSub, _ := s.repo.GetActiveSubscription(userID)

	var amountToPay float64
	var expiresAt time.Time
	var isUpgrade bool
	var isRenewal bool

	// Default duration is 1 day for dev
	const devDuration = 24 * time.Hour

	if activeSub == nil {
		amountToPay = newPkg.Price
		expiresAt = time.Now().Add(devDuration)
	} else if activeSub.PackageID == packageID {
		isRenewal = true
		amountToPay = newPkg.Price
		expiresAt = activeSub.ExpiresAt.Add(devDuration)
	} else {
		currentPkg, err := s.repo.FindPackageByID(activeSub.PackageID)
		if err != nil {
			return nil, err
		}

		if newPkg.Price > currentPkg.Price {
			isUpgrade = true
			amountToPay = newPkg.Price - currentPkg.Price
			expiresAt = time.Now().Add(devDuration)
		} else {
			return nil, errors.New("you already have an active subscription with same or higher tier")
		}
	}

	// Deduct balance via User Service (Inter-service call)
	if err := s.deductUserBalance(userID, amountToPay); err != nil {
		return nil, fmt.Errorf("insufficient balance or user-service error: %v", err)
	}

	var sub *Subscription
	if isUpgrade {
		err = s.repo.UpdateSubscription(activeSub.ID, packageID, expiresAt)
		if err != nil {
			return nil, err
		}
		sub = activeSub
		sub.PackageID = packageID
		sub.ExpiresAt = expiresAt
	} else if isRenewal {
		err = s.repo.ExtendSubscription(activeSub.ID, expiresAt)
		if err != nil {
			return nil, err
		}
		sub = activeSub
		sub.ExpiresAt = expiresAt
	} else {
		sub, err = s.repo.CreateSubscription(userID, packageID, expiresAt)
		if err != nil {
			return nil, err
		}
	}

	// Record payment
	method := "wallet_purchase"
	if isUpgrade {
		method = "wallet_upgrade"
	} else if isRenewal {
		method = "wallet_renewal"
	}
	_ = s.repo.RecordPayment(userID, sub.ID, amountToPay, method, "success")

	return sub, nil
}

func (s *Service) GetActiveSubscription(userID int) (*Subscription, error) {
	return s.repo.GetActiveSubscription(userID)
}

// deductUserBalance calls user-service internal API
func (s *Service) deductUserBalance(userID int, amount float64) error {
	url := fmt.Sprintf("%s/internal/users/%d/deduct", s.userSvcURL, userID)
	
	reqBody, _ := json.Marshal(map[string]float64{
		"amount": amount,
	})

	resp, err := http.Post(url, "application/json", bytes.NewBuffer(reqBody))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		var errResp map[string]string
		json.NewDecoder(resp.Body).Decode(&errResp)
		return errors.New(errResp["error"])
	}

	return nil
}
