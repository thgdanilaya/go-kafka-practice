package main

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	kafka "github.com/segmentio/kafka-go"
	"wbl0/internal/model"
	"wbl0/internal/repository"
)

func readEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func main() {
	broker := strings.Split(readEnv("KAFKA_BROKER", "localhost:9094"), ",")
	topic := readEnv("KAFKA_TOPIC", "orders")
	group := readEnv("KAFKA_GROUP_ID", "orders-demo")
	pgDSN := readEnv("PG_DSN", "postgres://app:app_pass@localhost:5432/orders?sslmode=disable")

	//connection to BD
	rp, err := repository.New(context.Background(), pgDSN)
	if err != nil {
		log.Fatalf("failed to connect pg: %v", err)
	}
	defer rp.Close()

	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:     broker,
		GroupID:     group,
		Topic:       topic,
		MinBytes:    1,
		MaxBytes:    10e6,
		StartOffset: kafka.FirstOffset,
	})
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	for {
		msg, err := reader.FetchMessage(ctx)
		if err != nil {
			if errors.Is(err, context.Canceled) {
				break
			}
			log.Printf("fetch error %v", err)
			continue
		}

		//start := time.Now()

		var order model.Order
		if err := json.Unmarshal(msg.Value, &order); err != nil || order.OrderUID == "" {
			log.Printf("invalid message partition=%d offset=%d: %v", msg.Partition, msg.Offset, err)
			// Возможно тут надо сделать дохлые сообщеня
			_ = reader.CommitMessages(ctx, msg)
			continue
		}

		dbCtx, cancelDB := context.WithTimeout(ctx, 5*time.Second)
		err = rp.UpsertOrder(dbCtx, order)
		cancelDB()
		if err != nil {
			// Если фэйл то будет повтор чтения с кафки
			log.Printf("db upsert failed (offset=%d): %v", msg.Offset, err)
			time.Sleep(500 * time.Millisecond)
			continue
		}

		// Если ок то коммитим
		if err := reader.CommitMessages(ctx, msg); err != nil {
			log.Printf("commit: %v", err)
		}
		log.Printf("OK saved order_uid=%s key=%s partition=%d offset=%d",
			order.OrderUID, string(msg.Key), msg.Partition, msg.Offset)
	}
	log.Println("Consumer stop")
}
