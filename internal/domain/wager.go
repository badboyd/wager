package domain

import (
	"context"
	"time"

	"github.com/shopspring/decimal"
)

type (
	// Wager ...
	Wager struct {
		ID                  int             `json:"id" db:"id"`
		TotalWagerValue     int             `json:"total_wager_value" db:"total_wager_value" validate:"required,min=1"`
		Odds                int             `json:"odds" db:"odds" validate:"required,min=1"`
		SellingPercentage   int             `json:"selling_percentage" db:"selling_percentage" validate:"required,min=1,max=100"`
		SellingPrice        decimal.Decimal `json:"selling_price" db:"selling_price" validate:"required,v_selling_price"`
		CurrentSellingPrice decimal.Decimal `json:"current_selling_price" db:"current_selling_price"`
		PercentageSold      *int            `json:"percentage_sold" db:"percentage_sold"`
		AmountSold          *int            `json:"amount_sold" db:"amount_sold"`
		PlacedAt            time.Time       `json:"placed_at" db:"placed_at"`
	}

	// Purchase ...
	Purchase struct {
		ID          int             `json:"id"`
		WagerID     int             `json:"wager_id"`
		BuyingPrice decimal.Decimal `json:"buying_price"`
		BoughtAt    time.Time       `json:"bought_at"`
	}
)

// WagerRepository interface
type WagerRepository interface {
	Create(ctx context.Context, wager Wager) (Wager, error)
	Get(ctx context.Context, wagerID, limit int) ([]Wager, int, error)
	Purchase(ctx context.Context, wagerID int, buyingPrice decimal.Decimal) (Purchase, error)
	Close(ctx context.Context) error
}
