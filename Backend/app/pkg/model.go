package pkg

import "time"

type Package struct {
	ID                      int       `json:"id"`
	Name                    string    `json:"name"`
	Description             string    `json:"description"`
	Price                   float64   `json:"price"`
	RequestLimit            int       `json:"request_limit"`
	RefreshIntervalMinutes  int       `json:"refresh_interval_minutes"`
	HistoricalDataDays      int       `json:"historical_data_days"`
	HasGenreAnalytics       bool      `json:"has_genre_analytics"`
	HasRevenueAnalytics     bool      `json:"has_revenue_analytics"`
	HasRegionBreakdown      bool      `json:"has_region_breakdown"`
	HasWebhook              bool      `json:"has_webhook"`
	HasBulkExport           bool      `json:"has_bulk_export"`
	HasCustomReports        bool      `json:"has_custom_reports"`
	HasDedicatedSupport     bool      `json:"has_dedicated_support"`
	HasSlaGuarantee         bool      `json:"has_sla_guarantee"`
	HasRealtimeStream       bool      `json:"has_realtime_stream"`
	CreatedAt               time.Time `json:"created_at"`
}

type Subscription struct {
	ID        int       `json:"id"`
	UserID    int       `json:"user_id"`
	PackageID int       `json:"package_id"`
	Status    string    `json:"status"`
	StartedAt time.Time `json:"started_at"`
	ExpiresAt time.Time `json:"expires_at"`
	CreatedAt time.Time `json:"created_at"`
}

type PurchaseRequest struct {
	PackageID int `json:"package_id" binding:"required"`
}
