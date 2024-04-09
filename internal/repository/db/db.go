package db

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/ArtemZ007/wb-l0/internal/cache"
	_ "github.com/lib/pq"
)

type DBService struct {
	db     *sql.DB
	cache  *cache.Cache
	ctx    context.Context
	cancel context.CancelFunc
	wg     sync.WaitGroup
}

// NewDBService создает новый сервис для работы с базой данных и кэшем.
func NewDBService(cfg *DBConfig, cache *cache.Cache) (*DBService, error) {
	dsn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		cfg.Host, cfg.Port, cfg.User, cfg.Password, cfg.DBName)
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("db connect error: %w", err)
	}
	if err = db.Ping(); err != nil {
		return nil, fmt.Errorf("db ping error: %w", err)
	}
	log.Println("Database connection established")

	ctx, cancel := context.WithCancel(context.Background())
	service := &DBService{
		db:     db,
		cache:  cache,
		ctx:    ctx,
		cancel: cancel,
	}

	return service, nil
}

// Start запускает процесс синхронизации данных между кэшем и базой данных.
func (s *DBService) Start() {
	s.wg.Add(1)
	go s.syncCacheToDB()
}

// Stop останавливает процесс синхронизации и ожидает его завершения.
func (s *DBService) Stop() {
	s.cancel()
	s.wg.Wait()
}

// syncCacheToDB выполняет периодическую синхронизацию данных из кэша в базу данных.
func (s *DBService) syncCacheToDB() {
	defer s.wg.Done()

	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-s.ctx.Done():
			log.Println("Stopping database sync...")
			return
		case <-ticker.C:
			if err := s.FillTablesFromCache(); err != nil {
				log.Printf("Error syncing cache to DB: %v", err)
			}
		}
	}
}

// FillTablesFromCache заполняет таблицы базы данных данными из кэша.
func (s *DBService) FillTablesFromCache() error {
	log.Println("Starting to fill tables from cache")

	orderIDs := s.cache.GetAllOrderIDs()
	for _, id := range orderIDs {
		order, found := s.cache.GetOrder(id)
		if !found {
			log.Printf("Order with ID %s not found in cache", id)
			continue
		}

		orderData, err := json.Marshal(order)
		if err != nil {
			log.Printf("Error serializing order with ID %s: %v", id, err)
			continue
		}

		query := `INSERT INTO orders (order_uid, order_data) VALUES ($1, $2) ON CONFLICT (order_uid) DO NOTHING`
		if _, err := s.db.ExecContext(s.ctx, query, order.OrderUID, orderData); err != nil {
			log.Printf("Error inserting order with ID %s into database: %v", id, err)
			continue
		}

		log.Printf("Order with ID %s successfully added to database from cache", id)
	}

	log.Println("Finished filling tables from cache")
	return nil
}
