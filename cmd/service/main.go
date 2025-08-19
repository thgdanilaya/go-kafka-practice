package main

//TODO Prometheus + поправить проверку транзакций
import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	kafka "github.com/segmentio/kafka-go"

	"wbl0/internal/cache"
	"wbl0/internal/httpapi"
	"wbl0/internal/model"
	"wbl0/internal/repository"
)

func env(key, defaultValue string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return defaultValue
}

func envInt(key string, def int) int {
	if v := os.Getenv(key); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			return n
		}
	}
	return def
}

func main() {
	httpAddr := env("HTTP_ADDR", ":8080")
	brokers := strings.Split(env("KAFKA_BROKER", "localhost:9094"), ",")
	topic := env("KAFKA_TOPIC", "orders")
	groupID := env("KAFKA_GROUP_ID", "orders-svc")
	pgDSN := env("PG_DSN", "postgres://app:app_pass@localhost:5432/orders?sslmode=disable")
	capacity := envInt("CACHE_CAPACITY", 10)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	repo, err := repository.New(ctx, pgDSN)
	if err != nil {
		log.Fatalf("pg connect %v", err)
	}
	defer repo.Close()

	c := cache.New(3)
	{
		pctx, cancel := context.WithTimeout(ctx, 5*time.Second)
		defer cancel()

		orders, err := repo.ListRecent(pctx, capacity)
		if err == nil {
			for i := len(orders) - 1; i >= 0; i-- {
				c.Set(orders[i])
			}
			log.Printf("cache preloaded: %d/%d", c.Len(), c.Capacity())
		} else {
			log.Printf("cache preload failed: %v", err)
		}
	}

	srv := httpapi.New(httpAddr, repo, c)

	dialer := &kafka.Dialer{
		Timeout:   10 * time.Second,
		DualStack: true,
		ClientID:  "orders-svc",
	}

	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:           brokers,
		GroupID:           groupID,
		Topic:             topic,
		Dialer:            dialer,
		StartOffset:       kafka.FirstOffset,
		MinBytes:          1,
		MaxBytes:          10e6,
		HeartbeatInterval: 3 * time.Second,
		SessionTimeout:    30 * time.Second,
		RebalanceTimeout:  30 * time.Second,
		Logger:            log.New(os.Stdout, "kafka-reader ", log.LstdFlags),
		ErrorLogger:       log.New(os.Stderr, "kafka-reader ERR ", log.LstdFlags),
	})
	defer reader.Close()

	go func() {
		log.Printf("consumer starting: brokers=%v topic=%s group=%s", brokers, topic, groupID)
		for {
			msg, err := reader.FetchMessage(ctx)
			if err != nil {
				if errors.Is(err, context.Canceled) {
					break
				}
				log.Printf("fetch error: %v", err)
				continue
			}
			var o model.Order
			if err := json.Unmarshal(msg.Value, &o); err != nil || o.OrderUID == "" {
				log.Printf("invalid message partition=%d offset=%d: %v", msg.Partition, msg.Offset, err)
				_ = reader.CommitMessages(ctx, msg)
				continue
			}

			dbCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
			err = repo.UpsertOrder(dbCtx, o)
			cancel()
			if err != nil {
				log.Printf("db upsert failed partition=%d offset=%d: %v", msg.Partition, msg.Offset, err)
				time.Sleep(500 * time.Millisecond)
				continue
			}

			c.Set(o)
			if err := reader.CommitMessages(ctx, msg); err != nil {
				log.Printf("commit failed partition=%d offset=%d: %v", msg.Partition, msg.Offset, err)
				continue
			}
			log.Printf("OK saved order_uid=%s partition=%d offset=%d", o.OrderUID, msg.Partition, msg.Offset)
		}
		log.Println("consumer stopped")
	}()

	go func() {
		log.Printf("http listening on %s", httpAddr)
		if err := srv.Start(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("http server: %v", err)
		}
	}()

	<-ctx.Done()
	log.Println("shutting down...")

	sdCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(sdCtx); err != nil {
		log.Printf("http shutdown error: %v", err)
	}
}
