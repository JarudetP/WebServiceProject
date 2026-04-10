package user

import (
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"errors"
)

type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) Create(u *User) (*User, error) {
	query := `
		INSERT INTO users (username, email, password_hash, full_name, company, role, balance)
		VALUES ($1, $2, $3, $4, $5, 'user', 0.00)
		RETURNING id, username, email, full_name, company, role, balance, is_active, created_at, updated_at`

	row := r.db.QueryRow(query, u.Username, u.Email, u.PasswordHash, u.FullName, u.Company)
	result := &User{}
	err := row.Scan(
		&result.ID, &result.Username, &result.Email,
		&result.FullName, &result.Company, &result.Role, &result.Balance,
		&result.IsActive, &result.CreatedAt, &result.UpdatedAt,
	)
	return result, err
}

func (r *Repository) FindByEmail(email string) (*User, error) {
	query := `SELECT id, username, email, password_hash, full_name, company, role, balance, is_active FROM users WHERE email = $1`
	row := r.db.QueryRow(query, email)
	u := &User{}
	err := row.Scan(&u.ID, &u.Username, &u.Email, &u.PasswordHash, &u.FullName, &u.Company, &u.Role, &u.Balance, &u.IsActive)
	if err == sql.ErrNoRows {
		return nil, errors.New("user not found")
	}
	return u, err
}

func (r *Repository) FindByID(id int) (*User, error) {
	query := `SELECT id, username, email, full_name, company, role, balance, is_active, created_at, updated_at FROM users WHERE id = $1`
	row := r.db.QueryRow(query, id)
	u := &User{}
	err := row.Scan(&u.ID, &u.Username, &u.Email, &u.FullName, &u.Company, &u.Role, &u.Balance, &u.IsActive, &u.CreatedAt, &u.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, errors.New("user not found")
	}
	return u, err
}

func (r *Repository) TopUp(userID int, amount float64) (float64, error) {
	query := `UPDATE users SET balance = balance + $1 WHERE id = $2 RETURNING balance`
	var newBalance float64
	err := r.db.QueryRow(query, amount, userID).Scan(&newBalance)
	return newBalance, err
}

func (r *Repository) DeductBalance(userID int, amount float64) error {
	query := `UPDATE users SET balance = balance - $1 WHERE id = $2 AND balance >= $1`
	result, err := r.db.Exec(query, amount, userID)
	if err != nil {
		return err
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return errors.New("insufficient balance")
	}
	return nil
}

func (r *Repository) FindByAPIKey(apiKey string) (int, int, string, error) {
	query := `
		SELECT k.user_id, k.id, u.role 
		FROM api_keys k
		JOIN users u ON k.user_id = u.id
		WHERE k.key = $1 AND k.is_active = TRUE`
	var userID, apiKeyID int
	var role string
	err := r.db.QueryRow(query, apiKey).Scan(&userID, &apiKeyID, &role)
	if err == sql.ErrNoRows {
		return 0, 0, "", errors.New("invalid or inactive API key")
	}
	return userID, apiKeyID, role, err
}

func (r *Repository) CreateAPIKey(userID int) (string, error) {
	// Generate a 32-character random key
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	key := hex.EncodeToString(b)

	query := `INSERT INTO api_keys (user_id, key, is_active) VALUES ($1, $2, TRUE)`
	_, err := r.db.Exec(query, userID, key)
	return key, err
}

func (r *Repository) GetAPIKeys(userID int) ([]string, error) {
	query := `SELECT key FROM api_keys WHERE user_id = $1 AND is_active = TRUE`
	rows, err := r.db.Query(query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var keys []string
	for rows.Next() {
		var k string
		if err := rows.Scan(&k); err != nil {
			return nil, err
		}
		keys = append(keys, k)
	}
	return keys, nil
}

func (r *Repository) DeleteAPIKey(userID int, key string) error {
	query := `DELETE FROM api_keys WHERE user_id = $1 AND key = $2`
	res, err := r.db.Exec(query, userID, key)
	if err != nil {
		return err
	}
	rows, _ := res.RowsAffected()
	if rows == 0 {
		return errors.New("key not found or unauthorized")
	}
	return nil
}
