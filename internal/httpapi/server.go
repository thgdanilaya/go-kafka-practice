package httpapi

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/jackc/pgx/v5"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"wbl0/internal/cache"
	"wbl0/internal/repository"
)

type Server struct {
	repo  *repository.Repository
	cache *cache.Cache
	http  *http.Server
}

func New(addr string, repo *repository.Repository, cache *cache.Cache) *Server {
	router := chi.NewRouter()

	router.Use(func(next http.Handler) http.Handler {
		return http.TimeoutHandler(next, 5*time.Second, "timeout")
	})

	router.Get("/healthz", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})

	router.Get("/api/orders/{id}", func(w http.ResponseWriter, request *http.Request) {
		id := chi.URLParam(request, "id")
		if id == "" {
			http.Error(w, "missing id", http.StatusBadRequest)
			return
		}
		if o, ok := cache.Get(id); ok {
			w.Header().Set("Cache", "HIT")
			w.Header().Set("Server-Timing", `cache;desc="hit";duration=0`)
			_ = json.NewEncoder(w).Encode(o)
			return
		}

		t0 := time.Now()
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		order, err := repo.GetOrder(ctx, id)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				http.Error(w, "server error", http.StatusInternalServerError)
			}
			return
		}
		duration := time.Since(t0)
		cache.Set(order)
		w.Header().Set("Cache", "MISS")
		w.Header().Set("Server-Timing", fmt.Sprintf(`db;duration=%.1f`, float64(duration.Microseconds())/1000.0))
		_ = json.NewEncoder(w).Encode(order)
	})

	router.Get("/debug/cache", func(w http.ResponseWriter, _ *http.Request) {
		hits, misses := cache.Stats()
		_ = json.NewEncoder(w).Encode(map[string]any{
			"size":     cache.Len(),
			"capacity": cache.Capacity(),
			"hits":     hits,
			"misses":   misses,
		})
	})

	router.Get("/", func(w http.ResponseWriter, req *http.Request) {
		http.ServeFile(w, req, "web/static/index.html")
	})

	return &Server{
		repo:  repo,
		cache: cache,
		http: &http.Server{
			Addr:         addr,
			Handler:      router,
			ReadTimeout:  5 * time.Second,
			WriteTimeout: 5 * time.Second,
			IdleTimeout:  120 * time.Second,
		},
	}
}

func (s *Server) Start() error { return s.http.ListenAndServe() }
func (s *Server) Shutdown(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	return s.http.Shutdown(ctx)
}
