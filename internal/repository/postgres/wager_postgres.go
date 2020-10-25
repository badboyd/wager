package repository

import (
	"context"

	"wager/internal/domain"

	"github.com/jmoiron/sqlx"
	"github.com/shopspring/decimal"
)

// PostgresRepository ...
type PostgresRepository struct {
	conn *sqlx.DB
}

// New returns new wager postgres repository
func New(conn *sqlx.DB) *PostgresRepository {
	return &PostgresRepository{
		conn: conn,
	}
}

// Create new wager
func (w *PostgresRepository) Create(ctx context.Context, wager *domain.Wager) error {
	query := `INSERT INTO wagers
		(total_wager_value, selling_price, odds, selling_percentage, current_selling_price)
		VALUES
		($1, $2, $3, $4, $5)
		RETURNING id, placed_at`

	return w.conn.GetContext(ctx, wager, query, wager.TotalWagerValue, wager.SellingPrice,
		wager.Odds, wager.SellingPercentage, wager.SellingPrice)
}

// Get list of wagers
func (w *PostgresRepository) Get(ctx context.Context, wagerID, limit int) ([]domain.Wager, int, error) {
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
func (w *PostgresRepository) Purchase(ctx context.Context, wagerID int, buyingPrice decimal.Decimal) (domain.Purchase, error) {

	return domain.Purchase{}, nil
}
