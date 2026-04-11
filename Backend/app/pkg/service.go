package pkg

import (
	"errors"
	"time"
)

type Service struct {
	repo     *Repository
	userRepo userBalanceRepo
}

// userBalanceRepo is an interface to deduct balance from user service
type userBalanceRepo interface {
	DeductBalance(userID int, amount float64) error
}

func NewService(repo *Repository, userRepo userBalanceRepo) *Service {
	return &Service{repo: repo, userRepo: userRepo}
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

	// Check if user already has an active subscription
	activeSub, _ := s.repo.GetActiveSubscription(userID)

	var amountToPay float64
	var expiresAt time.Time
	var isUpgrade bool
	var isRenewal bool

	// Default duration is 1 day for dev
	const devDuration = 24 * time.Hour

	if activeSub == nil {
		// 1. New Purchase
		amountToPay = newPkg.Price
		expiresAt = time.Now().Add(devDuration)
	} else if activeSub.PackageID == packageID {
		// 2. Renewal (Same package)
		isRenewal = true
		amountToPay = newPkg.Price
		expiresAt = activeSub.ExpiresAt.Add(devDuration)
	} else {
		// 3. Different package (Check for upgrade)
		currentPkg, err := s.repo.FindPackageByID(activeSub.PackageID)
		if err != nil {
			return nil, err
		}

		if newPkg.Price > currentPkg.Price {
			// Upgrade: Pay difference
			isUpgrade = true
			amountToPay = newPkg.Price - currentPkg.Price
			expiresAt = time.Now().Add(devDuration)
		} else {
			return nil, errors.New("you already have an active subscription with same or higher tier")
		}
	}

	// Deduct balance
	if err := s.userRepo.DeductBalance(userID, amountToPay); err != nil {
		return nil, errors.New("insufficient balance: " + err.Error())
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

func (s *Service) GetUsageStats(userID int) ([]UsageStat, error) {
	// Default to last 30 days
	return s.repo.GetUsageStats(userID, 30)
}
