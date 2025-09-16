package service

import (
	"context"
	"github.com/sdvaanyaa/order-service/internal/models"
	"github.com/sdvaanyaa/order-service/internal/repository"
	"github.com/sdvaanyaa/order-service/pkg/pgdb"
	"log/slog"
	"sync"
)

type OrderService interface {
	AddOrder(ctx context.Context, order *models.Order) error
	GetOrder(ctx context.Context, uid string) (*models.Order, error)
}

type orderService struct {
	repo       repository.OrderRepository
	transactor pgdb.Transactor
	log        *slog.Logger
	cache      map[string]*models.Order
	mu         sync.RWMutex
}

func New(repo repository.OrderRepository, transactor pgdb.Transactor, log *slog.Logger) OrderService {
	svc := &orderService{
		repo:       repo,
		transactor: transactor,
		log:        log,
		cache:      make(map[string]*models.Order),
	}

	svc.loadCache(context.Background())

	return svc
}

func (s *orderService) loadCache(ctx context.Context) {
	cached, err := s.repo.LoadAllOrders(ctx)
	if err != nil {
		s.log.Error("failed to load cache", "err", err)
		return
	}

	s.mu.Lock()
	s.cache = cached
	s.mu.Unlock()

	s.log.Info("cache loaded", "count", len(cached))
}

func (s *orderService) AddOrder(ctx context.Context, order *models.Order) error {
	return s.transactor.WithinTransaction(ctx, func(txCtx context.Context) error {
		err := s.repo.SaveOrder(txCtx, order)
		if err != nil {
			return err
		}

		s.mu.Lock()
		s.cache[order.OrderUID] = order
		s.mu.Unlock()

		return nil
	})
}

func (s *orderService) GetOrder(ctx context.Context, uid string) (*models.Order, error) {
	s.mu.RLock()
	order, ok := s.cache[uid]
	s.mu.RUnlock()
	if ok {
		return order, nil
	}

	order, err := s.repo.GetOrderByUID(ctx, uid)
	if err != nil {
		return nil, err
	}

	if order != nil {
		s.mu.Lock()
		s.cache[uid] = order
		s.mu.Unlock()
	}

	return order, nil
}
