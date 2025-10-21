package database

import (
	"context"
)

// UpdateSubscriptionName обновляет имя подписки
func (sr *SubscriptionRepository) UpdateSubscriptionName(ctx context.Context, subscriptionID int64, newName string) error {
	updates := map[string]interface{}{"name": newName}
	return sr.UpdateSubscription(ctx, subscriptionID, updates)
}
