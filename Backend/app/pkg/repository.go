package pkg

import (
	"database/sql"
	"errors"
	"time"
)

type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) ListPackages() ([]Package, error) {
	query := `SELECT id, name, description, price, request_limit, refresh_interval_minutes,
		historical_data_days, has_genre_analytics, has_revenue_analytics, has_region_breakdown,
		has_webhook, has_bulk_export, has_custom_reports, has_dedicated_support,
		has_sla_guarantee, has_realtime_stream, created_at FROM packages ORDER BY price ASC`

	rows, err := r.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var packages []Package
	for rows.Next() {
		var p Package
		err := rows.Scan(
			&p.ID, &p.Name, &p.Description, &p.Price,
			&p.RequestLimit, &p.RefreshIntervalMinutes, &p.HistoricalDataDays,
			&p.HasGenreAnalytics, &p.HasRevenueAnalytics, &p.HasRegionBreakdown,
			&p.HasWebhook, &p.HasBulkExport, &p.HasCustomReports,
			&p.HasDedicatedSupport, &p.HasSlaGuarantee, &p.HasRealtimeStream, &p.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		packages = append(packages, p)
	}
	return packages, nil
}

func (r *Repository) FindPackageByID(id int) (*Package, error) {
	query := `SELECT id, name, description, price, request_limit, refresh_interval_minutes,
		historical_data_days, has_genre_analytics, has_revenue_analytics, has_region_breakdown,
		has_webhook, has_bulk_export, has_custom_reports, has_dedicated_support,
		has_sla_guarantee, has_realtime_stream, created_at FROM packages WHERE id = $1`

	p := &Package{}
	err := r.db.QueryRow(query, id).Scan(
		&p.ID, &p.Name, &p.Description, &p.Price,
		&p.RequestLimit, &p.RefreshIntervalMinutes, &p.HistoricalDataDays,
		&p.HasGenreAnalytics, &p.HasRevenueAnalytics, &p.HasRegionBreakdown,
		&p.HasWebhook, &p.HasBulkExport, &p.HasCustomReports,
		&p.HasDedicatedSupport, &p.HasSlaGuarantee, &p.HasRealtimeStream, &p.CreatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, errors.New("package not found")
	}
	return p, err
}

func (r *Repository) CreateSubscription(userID, packageID int, expiresAt time.Time) (*Subscription, error) {
	query := `
		INSERT INTO subscriptions (user_id, package_id, status, started_at, expires_at)
		VALUES ($1, $2, 'active', NOW(), $3)
		RETURNING id, user_id, package_id, status, started_at, expires_at, created_at`

	sub := &Subscription{}
	err := r.db.QueryRow(query, userID, packageID, expiresAt).Scan(
		&sub.ID, &sub.UserID, &sub.PackageID, &sub.Status,
		&sub.StartedAt, &sub.ExpiresAt, &sub.CreatedAt,
	)
	return sub, err
}

func (r *Repository) GetActiveSubscription(userID int) (*Subscription, error) {
	query := `SELECT id, user_id, package_id, status, started_at, expires_at, created_at
		FROM subscriptions WHERE user_id = $1 AND status = 'active' AND expires_at > NOW()
		ORDER BY created_at DESC LIMIT 1`

	sub := &Subscription{}
	err := r.db.QueryRow(query, userID).Scan(
		&sub.ID, &sub.UserID, &sub.PackageID, &sub.Status,
		&sub.StartedAt, &sub.ExpiresAt, &sub.CreatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, errors.New("no active subscription")
	}
	return sub, err
}

func (r *Repository) RecordPayment(userID, subscriptionID int, amount float64, method, status string) error {
	query := `INSERT INTO payments (user_id, subscription_id, amount, currency, payment_method, status, paid_at)
		VALUES ($1, $2, $3, 'USD', $4, $5, NOW())`
	_, err := r.db.Exec(query, userID, subscriptionID, amount, method, status)
	return err
}

func (r *Repository) UpdateSubscription(subID int, packageID int, expiresAt time.Time) error {
	query := `UPDATE subscriptions SET package_id = $1, expires_at = $2, started_at = NOW() WHERE id = $3`
	_, err := r.db.Exec(query, packageID, expiresAt, subID)
	return err
}

func (r *Repository) ExtendSubscription(subID int, expiresAt time.Time) error {
	query := `UPDATE subscriptions SET expires_at = $1 WHERE id = $2`
	_, err := r.db.Exec(query, expiresAt, subID)
	return err
}

func (r *Repository) GetUsageCountInInterval(userID int, minutes int) (int, error) {
	query := `SELECT COUNT(*) FROM api_usage_logs 
			  WHERE user_id = $1 AND requested_at > NOW() - (INTERVAL '1 minute' * $2)`
	var count int
	err := r.db.QueryRow(query, userID, minutes).Scan(&count)
	return count, err
}

func (r *Repository) LogAPIUsage(userID, apiKeyID int, endpoint, method string, statusCode int) error {
	query := `INSERT INTO api_usage_logs (user_id, api_key_id, endpoint, method, status_code)
			  VALUES ($1, $2, $3, $4, $5)`
	_, err := r.db.Exec(query, userID, apiKeyID, endpoint, method, statusCode)
	return err
}

func (r *Repository) GetUsageStats(userID int, days int) ([]UsageStat, error) {
	query := `
		SELECT 
			TO_CHAR(date_series, 'YYYY-MM-DD') as date,
			COALESCE(usage.count, 0) as count
		FROM (
			SELECT generate_series(
				CURRENT_DATE - (INTERVAL '1 day' * ($2 - 1)), 
				CURRENT_DATE, 
				'1 day'::interval
			)::date as date_series
		) d
		LEFT JOIN (
			SELECT 
				requested_at::date as usage_date, 
				COUNT(*) as count
			FROM api_usage_logs
			WHERE user_id = $1 AND requested_at >= CURRENT_DATE - (INTERVAL '1 day' * ($2 - 1))
			GROUP BY requested_at::date
		) usage ON d.date_series = usage.usage_date
		ORDER BY d.date_series ASC
	`
	rows, err := r.db.Query(query, userID, days)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var stats []UsageStat
	for rows.Next() {
		var s UsageStat
		if err := rows.Scan(&s.Date, &s.Count); err != nil {
			return nil, err
		}
		stats = append(stats, s)
	}
	return stats, nil
}
