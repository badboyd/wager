package domain

import (
	"context"
	"errors"
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
		WagerID     int             `json:"wager_id" validate:"required"`
		BuyingPrice decimal.Decimal `json:"buying_price" validate:"required"`
		BoughtAt    time.Time       `json:"bought_at"`
	}
)

const (
	sellingPriceScale = 2
)

const (
	ErrInvalidWagerID           = "wager_id is required and must be greater than 0"
	ErrInvalidBuyingPrice       = "buying_price is required and must be greater than 0"
	ErrInvalidTotalWagerValue   = "total_wager_value is required and must be greater than 0"
	ErrInvalidOdds              = "odds is required and must be greater than 0"
	ErrInvalidSellingPercentage = "selling_percentage is invalid"
	ErrInvalidSellingPrice      = "selling_price is required with scale 2 and must be greater than total_wager_value * selling_percentage/100"
)

// Validate wager
func (w *Wager) Validate(ctx context.Context) error {
	if w.TotalWagerValue <= 0 {
		return errors.New(ErrInvalidTotalWagerValue)
	}

	if w.Odds <= 0 {
		return errors.New(ErrInvalidOdds)
	}

	if w.SellingPercentage < 0 || w.SellingPercentage > 100 {
		return errors.New(ErrInvalidSellingPercentage)
	}

	// we only allow decimal value with sellingPriceScale
	if w.SellingPrice.Exponent() > sellingPriceScale {
		return errors.New(ErrInvalidSellingPrice)
	}

	if w.SellingPrice.LessThan(decimal.NewFromInt(int64(w.TotalWagerValue * w.SellingPercentage / 100))) {
		return errors.New(ErrInvalidSellingPrice)
	}

	return nil
}

// Validate purchase
func (p *Purchase) Validate(cxt context.Context) error {
	if p.WagerID <= 0 {
		return errors.New(ErrInvalidWagerID)
	}

	if p.BuyingPrice.LessThanOrEqual(decimal.Zero) {
		return errors.New(ErrInvalidBuyingPrice)
	}

	return nil
}

// WagerRepository interface
type WagerRepository interface {
	Create(ctx context.Context, wager Wager) (Wager, error)
	Get(ctx context.Context, wagerID, limit int) ([]Wager, int, error)
	Purchase(ctx context.Context, wagerID int, buyingPrice decimal.Decimal) (Purchase, error)
	Close(ctx context.Context) error
}
