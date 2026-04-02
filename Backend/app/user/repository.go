package user

import (
	"database/sql"
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
		INSERT INTO users (username, email, password_hash, full_name, company, balance)
		VALUES ($1, $2, $3, $4, $5, 0.00)
		RETURNING id, username, email, full_name, company, balance, is_active, created_at, updated_at`

	row := r.db.QueryRow(query, u.Username, u.Email, u.PasswordHash, u.FullName, u.Company)
	result := &User{}
	err := row.Scan(
		&result.ID, &result.Username, &result.Email,
		&result.FullName, &result.Company, &result.Balance,
		&result.IsActive, &result.CreatedAt, &result.UpdatedAt,
	)
	return result, err
}

func (r *Repository) FindByEmail(email string) (*User, error) {
	query := `SELECT id, username, email, password_hash, full_name, company, balance, is_active FROM users WHERE email = $1`
	row := r.db.QueryRow(query, email)
	u := &User{}
	err := row.Scan(&u.ID, &u.Username, &u.Email, &u.PasswordHash, &u.FullName, &u.Company, &u.Balance, &u.IsActive)
	if err == sql.ErrNoRows {
		return nil, errors.New("user not found")
	}
	return u, err
}

func (r *Repository) FindByID(id int) (*User, error) {
	query := `SELECT id, username, email, full_name, company, balance, is_active, created_at, updated_at FROM users WHERE id = $1`
	row := r.db.QueryRow(query, id)
	u := &User{}
	err := row.Scan(&u.ID, &u.Username, &u.Email, &u.FullName, &u.Company, &u.Balance, &u.IsActive, &u.CreatedAt, &u.UpdatedAt)
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
