package main

import (
	"context"
	"encoding/json"
	"log"
	"os"
	"strings"
	"time"
	"wbl0/internal/model"

	kafka "github.com/segmentio/kafka-go"
)

func read_env(key, default_value string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return default_value
}

func main() {
	brokers := strings.Split(read_env("KAFKA_BROKERS", "localhost:9094"), ",")
	topic := read_env("KAFKA_TOPIC", "orders")

	writer := &kafka.Writer{
		Addr:                   kafka.TCP(brokers...),
		Topic:                  topic,
		Balancer:               &kafka.Hash{},
		AllowAutoTopicCreation: false,
		RequiredAcks:           kafka.RequireAll,
		BatchTimeout:           10 * time.Second,
		BatchSize:              100,
	}
	defer writer.Close()

	order := model.Order{
		OrderUID:    "livetest",
		TrackNumber: "WBILMTESTTRACK",
		Entry:       "WBIL",
		Delivery:    model.Delivery{"Test Testov", "+9720000000", "2639809", "Kiryat Mozkin", "Ploshad Mira 15", "Kraiot", "test@gmail.com"},
		Payment:     model.Payment{"livetest", "", "USD", "wbpay", 1817, 1637907727, "alpha", 1500, 317, 0},
		Items: []model.Item{
			{ChrtID: 9934930, TrackNumber: "WBILMTESTTRACK", Price: 453, RID: "ab4219087a764ae0btest", Name: "Mascaras", Sale: 30, Size: "0", TotalPrice: 317, NmID: 2389212, Brand: "Vivienne Sabo", Status: 202},
		},
		Locale:            "en",
		InternalSignature: "",
		CustomerID:        "test",
		DeliveryService:   "meest",
		ShardKey:          "9",
		SmID:              99,
		DateCreated:       time.Now(),
		OofShard:          "1",
	}

	payload, _ := json.Marshal(order)

	message := kafka.Message{
		Key:   []byte(order.OrderUID),
		Value: payload,
		Headers: []kafka.Header{
			{Key: "content-type", Value: []byte("application/json")},
			{Key: "source", Value: []byte("kgo-demo")},
		},
	}

	ctx, cancel := context.WithTimeout(context.Background(), 25*time.Second)
	defer cancel()

	if err := writer.WriteMessages(ctx, message); err != nil {
		log.Fatal("write: %v", err)
	}
	log.Printf("write %d orders to kafka\n", order.OrderUID)
}
