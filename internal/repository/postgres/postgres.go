package postgres

import (
	"context"
	"fmt"

	"github.com/jmoiron/sqlx"
	"github.com/shopspring/decimal"

	"wager/internal/domain"
)

// Repository ...
type Repository struct {
	conn *sqlx.DB
}

// New returns new wager postgres repository
func New(conn *sqlx.DB) *Repository {
	return &Repository{
		conn: conn,
	}
}

// Create new wager
func (w *Repository) Create(ctx context.Context, wager domain.Wager) (domain.Wager, error) {
	query := `INSERT INTO wagers
		(total_wager_value, selling_price, odds, selling_percentage, current_selling_price)
		VALUES
		($1, $2, $3, $4, $5)
		RETURNING *`

	res := domain.Wager{}
	err := w.conn.GetContext(ctx, &res, query, wager.TotalWagerValue, wager.SellingPrice,
		wager.Odds, wager.SellingPercentage, wager.SellingPrice)

	return res, err
}

// Get list of wagers
func (w *Repository) Get(ctx context.Context, wagerID, limit int) ([]domain.Wager, int, error) {
	wagers := []domain.Wager{}

	query := `SELECT * FROM wagers WHERE ID > $1 ORDER BY ID LIMIT $2`
	if err := w.conn.SelectContext(ctx, &wagers, query, wagerID, limit); err != nil {
		return nil, 0, err
	}

	if len(wagers) == 0 {
		return nil, 0, nil
	}

	return wagers, wagers[len(wagers)-1].ID, nil
}

// Purchase a wager
func (w *Repository) Purchase(ctx context.Context, wagerID int, buyingPrice decimal.Decimal) (domain.Purchase, error) {
	purchase := domain.Purchase{}

	tx, err := w.conn.BeginTx(ctx, nil)
	if err != nil {
		return purchase, err
	}

	defer func() {
		if err != nil {
			err = tx.Rollback()
		} else {
			err = tx.Commit()
		}

		if err != nil {
			fmt.Printf("Rollback or commit error here: %s", err.Error())
		}
	}()

	lockQuery := `
		SELECT id, current_selling_price, total_wager_value, amount_sold
		FROM wagers
		WHERE id = $1 FOR UPDATE`

	wager := domain.Wager{}
	err = tx.QueryRowContext(ctx, lockQuery, wagerID).Scan(
		&wager.ID,
		&wager.CurrentSellingPrice,
		&wager.TotalWagerValue,
		&wager.AmountSold,
	)
	if err != nil {
		return purchase, err
	}

	if wager.ID == 0 {
		return purchase, fmt.Errorf("Wager %d is not found", wagerID)
	}

	if buyingPrice.GreaterThan(wager.CurrentSellingPrice) {
		return purchase, fmt.Errorf("buying_price must be less than current_selling_price")
	}

	var amountSold, percentageSold int
	if wager.AmountSold != nil {
		amountSold = *wager.AmountSold + 1
	}
	percentageSold = amountSold * 100 / wager.TotalWagerValue

	updateWagerQuery := `UPDATE wagers
		SET (current_selling_price, amount_sold, percentage_sold) = ($1, $2, $3)
		WHERE ID = $4`

	_, err = tx.ExecContext(ctx, updateWagerQuery, buyingPrice,
		amountSold, percentageSold, wagerID)
	if err != nil {
		return purchase, err
	}

	insertPurchaseQuery := `INSERT INTO purchases
		(wager_id, buying_price)
		VALUES
		($1, $2)
		RETURNING id, wager_id, buying_price, bought_at`

	err = tx.QueryRowContext(ctx, insertPurchaseQuery, wagerID, buyingPrice).Scan(
		&purchase.ID,
		&purchase.WagerID,
		&purchase.BuyingPrice,
		&purchase.BoughtAt,
	)
	return purchase, err
}

// Close the repository
func (w *Repository) Close(ctx context.Context) error {
	return w.conn.DB.Close()
}
