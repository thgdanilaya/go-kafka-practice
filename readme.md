```wbl0/
├─ cmd/service/               # основной сервис
├─ cmd/producer/              # утилита отправки тестового заказа в Kafka
├─ cmd/consumer/              # тестовая утилита получения заказа из Kafka
├─ internal/                  # cache, httpapi, repository, model, queries
├─ migrations/001_init.sql    # схема Postgres
├─ web/static/index.html      # страница поиска заказа
└─ deploy/compose/docker-compose.yml
```

# Настройка .env
```
POSTGRES_PASSWORD=localroot
APP_DB_USER=app
APP_DB_PASSWORD=app_pass
APP_DB_NAME=orders
KAFKA_EXTERNAL_PORT=9094
HTTP_ADDR=:8080
KAFKA_BROKER=127.0.0.1:9094
KAFKA_TOPIC=orders
KAFKA_GROUP_ID=orders-svc
PG_DSN=postgres://app:app_pass@localhost:5432/orders?sslmode=disable
CACHE_CAPACITY=10
```

# Запуск контейнеров
```
docker compose --env-file .env -f deploy/compose/docker-compose.yml up -d
```
# Создание Kafka-topic
```
docker compose -f deploy/compose/docker-compose.yml exec kafka kafka-topics.sh --create --topic orders --bootstrap-server localhost:9092 --partitions 3 --replication-factor 1
```
# Запуск сервисa
```
go run .\cmd\service\
```
# Добавление заказа
```
go run .\cmd\producer\
```
# API
```
Invoke-RestMethod -Uri "http://127.0.0.1:8080/api/orders/testorder"
```
# Web страница
находится на локалхосте