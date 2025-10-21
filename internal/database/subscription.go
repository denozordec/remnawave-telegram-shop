package database

import (
	"context"
	"errors"
	"fmt"
	sq "github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
	"time"
)

type Subscription struct {
	ID               int64      `db:"id"`
	CustomerID       int64      `db:"customer_id"`
	SubscriptionLink string     `db:"subscription_link"`
	ExpireAt         time.Time  `db:"expire_at"`
	CreatedAt        time.Time  `db:"created_at"`
	IsActive         bool       `db:"is_active"`
	Name             string     `db:"name"`
	Description      string     `db:"description"`
}

type SubscriptionRepository struct {
	pool *pgxpool.Pool
}

func NewSubscriptionRepository(pool *pgxpool.Pool) *SubscriptionRepository {
	return &SubscriptionRepository{pool: pool}
}

// CreateSubscription создает новую подписку для клиента
func (sr *SubscriptionRepository) CreateSubscription(ctx context.Context, subscription *Subscription) (*Subscription, error) {
	buildInsert := sq.Insert("subscription").
		Columns("customer_id", "subscription_link", "expire_at", "is_active", "name", "description").
		Values(subscription.CustomerID, subscription.SubscriptionLink, subscription.ExpireAt, subscription.IsActive, subscription.Name, subscription.Description).
		Suffix("RETURNING id, created_at").
		PlaceholderFormat(sq.Dollar)

	sqlStr, args, err := buildInsert.ToSql()
	if err != nil {
		return nil, fmt.Errorf("failed to build insert query: %w", err)
	}

	var id int64
	var createdAt time.Time
	err = sr.pool.QueryRow(ctx, sqlStr, args...).Scan(&id, &createdAt)
	if err != nil {
		return nil, fmt.Errorf("failed to insert subscription: %w", err)
	}

	subscription.ID = id
	subscription.CreatedAt = createdAt

	// Увеличиваем счетчик подписок у клиента
	err = sr.updateCustomerSubscriptionCount(ctx, subscription.CustomerID)
	if err != nil {
		return nil, fmt.Errorf("failed to update customer subscription count: %w", err)
	}

	return subscription, nil
}

// GetActiveSubscriptions возвращает все активные подписки клиента
func (sr *SubscriptionRepository) GetActiveSubscriptions(ctx context.Context, customerID int64) ([]Subscription, error) {
	buildSelect := sq.Select("id", "customer_id", "subscription_link", "expire_at", "created_at", "is_active", "name", "description").
		From("subscription").
		Where(sq.And{
			sq.Eq{"customer_id": customerID},
			sq.Eq{"is_active": true},
		}).
		OrderBy("created_at DESC").
		PlaceholderFormat(sq.Dollar)

	sqlStr, args, err := buildSelect.ToSql()
	if err != nil {
		return nil, fmt.Errorf("failed to build select query: %w", err)
	}

	rows, err := sr.pool.Query(ctx, sqlStr, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query subscriptions: %w", err)
	}
	defer rows.Close()

	var subscriptions []Subscription
	for rows.Next() {
		var sub Subscription
		err := rows.Scan(
			&sub.ID,
			&sub.CustomerID,
			&sub.SubscriptionLink,
			&sub.ExpireAt,
			&sub.CreatedAt,
			&sub.IsActive,
			&sub.Name,
			&sub.Description,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan subscription row: %w", err)
		}
		subscriptions = append(subscriptions, sub)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating over subscription rows: %w", err)
	}

	return subscriptions, nil
}

// GetAllSubscriptions возвращает все подписки клиента (включая неактивные)
func (sr *SubscriptionRepository) GetAllSubscriptions(ctx context.Context, customerID int64) ([]Subscription, error) {
	buildSelect := sq.Select("id", "customer_id", "subscription_link", "expire_at", "created_at", "is_active", "name", "description").
		From("subscription").
		Where(sq.Eq{"customer_id": customerID}).
		OrderBy("created_at DESC").
		PlaceholderFormat(sq.Dollar)

	sqlStr, args, err := buildSelect.ToSql()
	if err != nil {
		return nil, fmt.Errorf("failed to build select query: %w", err)
	}

	rows, err := sr.pool.Query(ctx, sqlStr, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query subscriptions: %w", err)
	}
	defer rows.Close()

	var subscriptions []Subscription
	for rows.Next() {
		var sub Subscription
		err := rows.Scan(
			&sub.ID,
			&sub.CustomerID,
			&sub.SubscriptionLink,
			&sub.ExpireAt,
			&sub.CreatedAt,
			&sub.IsActive,
			&sub.Name,
			&sub.Description,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan subscription row: %w", err)
		}
		subscriptions = append(subscriptions, sub)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating over subscription rows: %w", err)
	}

	return subscriptions, nil
}

// GetSubscriptionByID возвращает подписку по ID
func (sr *SubscriptionRepository) GetSubscriptionByID(ctx context.Context, id int64) (*Subscription, error) {
	buildSelect := sq.Select("id", "customer_id", "subscription_link", "expire_at", "created_at", "is_active", "name", "description").
		From("subscription").
		Where(sq.Eq{"id": id}).
		PlaceholderFormat(sq.Dollar)

	sqlStr, args, err := buildSelect.ToSql()
	if err != nil {
		return nil, fmt.Errorf("failed to build select query: %w", err)
	}

	var sub Subscription
	err = sr.pool.QueryRow(ctx, sqlStr, args...).Scan(
		&sub.ID,
		&sub.CustomerID,
		&sub.SubscriptionLink,
		&sub.ExpireAt,
		&sub.CreatedAt,
		&sub.IsActive,
		&sub.Name,
		&sub.Description,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to query subscription: %w", err)
	}

	return &sub, nil
}

// UpdateSubscription обновляет подписку
func (sr *SubscriptionRepository) UpdateSubscription(ctx context.Context, id int64, updates map[string]interface{}) error {
	if len(updates) == 0 {
		return nil
	}

	buildUpdate := sq.Update("subscription").
		PlaceholderFormat(sq.Dollar).
		Where(sq.Eq{"id": id})

	for field, value := range updates {
		buildUpdate = buildUpdate.Set(field, value)
	}

	sqlStr, args, err := buildUpdate.ToSql()
	if err != nil {
		return fmt.Errorf("failed to build update query: %w", err)
	}

	result, err := sr.pool.Exec(ctx, sqlStr, args...)
	if err != nil {
		return fmt.Errorf("failed to update subscription: %w", err)
	}

	rowsAffected := result.RowsAffected()
	if rowsAffected == 0 {
		return fmt.Errorf("no subscription found with id: %d", id)
	}

	return nil
}

// DeactivateSubscription деактивирует подписку
func (sr *SubscriptionRepository) DeactivateSubscription(ctx context.Context, id int64) error {
	// Получаем информацию о подписке для обновления счетчика
	sub, err := sr.GetSubscriptionByID(ctx, id)
	if err != nil {
		return err
	}
	if sub == nil {
		return fmt.Errorf("subscription with id %d not found", id)
	}

	updates := map[string]interface{}{
		"is_active": false,
	}

	err = sr.UpdateSubscription(ctx, id, updates)
	if err != nil {
		return err
	}

	// Обновляем счетчик подписок у клиента
	return sr.updateCustomerSubscriptionCount(ctx, sub.CustomerID)
}

// FindExpiredSubscriptions находит просроченные подписки
func (sr *SubscriptionRepository) FindExpiredSubscriptions(ctx context.Context) ([]Subscription, error) {
	buildSelect := sq.Select("id", "customer_id", "subscription_link", "expire_at", "created_at", "is_active", "name", "description").
		From("subscription").
		Where(sq.And{
			sq.Eq{"is_active": true},
			sq.Lt{"expire_at": time.Now()},
		}).
		PlaceholderFormat(sq.Dollar)

	sqlStr, args, err := buildSelect.ToSql()
	if err != nil {
		return nil, fmt.Errorf("failed to build select query: %w", err)
	}

	rows, err := sr.pool.Query(ctx, sqlStr, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query expired subscriptions: %w", err)
	}
	defer rows.Close()

	var subscriptions []Subscription
	for rows.Next() {
		var sub Subscription
		err := rows.Scan(
			&sub.ID,
			&sub.CustomerID,
			&sub.SubscriptionLink,
			&sub.ExpireAt,
			&sub.CreatedAt,
			&sub.IsActive,
			&sub.Name,
			&sub.Description,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan subscription row: %w", err)
		}
		subscriptions = append(subscriptions, sub)
	}

	return subscriptions, nil
}

// updateCustomerSubscriptionCount обновляет счетчик активных подписок у клиента
func (sr *SubscriptionRepository) updateCustomerSubscriptionCount(ctx context.Context, customerID int64) error {
	// Подсчитываем активные подписки
	buildCount := sq.Select("COUNT(*)").
		From("subscription").
		Where(sq.And{
			sq.Eq{"customer_id": customerID},
			sq.Eq{"is_active": true},
		}).
		PlaceholderFormat(sq.Dollar)

	countSql, countArgs, err := buildCount.ToSql()
	if err != nil {
		return fmt.Errorf("failed to build count query: %w", err)
	}

	var count int
	err = sr.pool.QueryRow(ctx, countSql, countArgs...).Scan(&count)
	if err != nil {
		return fmt.Errorf("failed to count subscriptions: %w", err)
	}

	// Обновляем счетчик в таблице customer
	buildUpdate := sq.Update("customer").
		Set("subscription_count", count).
		Where(sq.Eq{"id": customerID}).
		PlaceholderFormat(sq.Dollar)

	updateSql, updateArgs, err := buildUpdate.ToSql()
	if err != nil {
		return fmt.Errorf("failed to build update query: %w", err)
	}

	_, err = sr.pool.Exec(ctx, updateSql, updateArgs...)
	if err != nil {
		return fmt.Errorf("failed to update customer subscription count: %w", err)
	}

	return nil
}