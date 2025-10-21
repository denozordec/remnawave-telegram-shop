package database

import (
	"context"
	"fmt"
)

type Subscription struct {
	ID               int64
	CustomerID       int64
	SubscriptionLink string
	ExpireAt         Time
	IsActive         bool
	Name             string
	Description      string
}

type SubscriptionRepository struct { pool *Pool }

func NewSubscriptionRepository(pool *Pool) *SubscriptionRepository { return &SubscriptionRepository{pool: pool} }

// UpdateSubscriptionName обновляет имя подписки
func (r *SubscriptionRepository) UpdateSubscriptionName(ctx context.Context, subscriptionID int64, newName string) error {
	query := `UPDATE subscriptions SET name = $1 WHERE id = $2`
	_, err := r.pool.Exec(ctx, query, newName, subscriptionID)
	if err != nil { return fmt.Errorf("update subscription name: %w", err) }
	return nil
}
