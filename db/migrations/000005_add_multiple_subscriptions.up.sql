-- Создаем таблицу для хранения подписок
CREATE TABLE subscription (
    id                BIGSERIAL PRIMARY KEY,
    customer_id       BIGINT REFERENCES customer (id) ON DELETE CASCADE,
    subscription_link TEXT NOT NULL,
    expire_at         TIMESTAMP WITH TIME ZONE NOT NULL,
    created_at        TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    is_active         BOOLEAN DEFAULT TRUE,
    name              TEXT DEFAULT '',
    description       TEXT DEFAULT ''
);

-- Создаем индекс для быстрого поиска активных подписок
CREATE INDEX idx_subscription_customer_active ON subscription (customer_id, is_active);
CREATE INDEX idx_subscription_expire_at ON subscription (expire_at);

-- Мигрируем существующие данные из таблицы customer
INSERT INTO subscription (customer_id, subscription_link, expire_at, created_at)
SELECT id, subscription_link, expire_at, created_at
FROM customer
WHERE subscription_link IS NOT NULL;

-- Добавляем поле для подсчета количества подписок в таблице customer
ALTER TABLE customer ADD COLUMN subscription_count INTEGER DEFAULT 0;

-- Обновляем счетчик подписок для существующих клиентов
UPDATE customer 
SET subscription_count = (
    SELECT COUNT(*) 
    FROM subscription 
    WHERE subscription.customer_id = customer.id AND is_active = TRUE
);