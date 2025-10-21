-- Отмена миграции множественных подписок
ALTER TABLE customer DROP COLUMN subscription_count;
DROP TABLE subscription;