package service

import (
	"context"
	"errors"
	"github.com/go-playground/validator/v10"
	"github.com/gojuno/minimock/v3"
	"github.com/sdvaanyaa/order-service/internal/models"
	"github.com/sdvaanyaa/order-service/internal/repository"
	rmocks "github.com/sdvaanyaa/order-service/internal/repository/mocks"
	tmocks "github.com/sdvaanyaa/order-service/pkg/pgdb/mocks"
	"github.com/stretchr/testify/assert"
	"log/slog"
	"testing"
	"time"
)

var (
	ErrDB = errors.New("db error")
	ErrTx = errors.New("tx error")
)

func Test_orderService_AddOrder(t *testing.T) {
	t.Parallel()

	now := time.Now().UTC()
	order := &models.Order{
		OrderUID:          "uid1",
		TrackNumber:       "track1",
		Entry:             "entry",
		Locale:            "en",
		InternalSignature: "",
		CustomerID:        "cust",
		DeliveryService:   "serv",
		Shardkey:          "9",
		SmID:              99,
		DateCreated:       now,
		OofShard:          "1",
		Delivery: models.Delivery{
			Name:    "name",
			Phone:   "+123",
			Zip:     "zip",
			City:    "city",
			Address: "addr",
			Region:  "region",
			Email:   "email@example.com",
		},
		Payment: models.Payment{
			Transaction:  "tx1",
			Currency:     "USD",
			Provider:     "prov",
			Amount:       100,
			PaymentDt:    now.Unix(),
			Bank:         "bank",
			DeliveryCost: 10,
			GoodsTotal:   90,
			CustomFee:    0,
		},
		Items: []models.Item{
			{
				ChrtID:      1,
				TrackNumber: "itemtrack",
				Price:       50,
				Name:        "item",
				Sale:        0,
				Size:        "M",
				TotalPrice:  50,
				NmID:        2,
				Brand:       "brand",
				Status:      200,
			},
		},
	}

	type fields struct {
		repoMock       *rmocks.OrderRepositoryMock
		transactorMock *tmocks.TransactorMock
	}
	type args struct {
		ctx   context.Context
		order *models.Order
	}
	tests := []struct {
		name       string
		prepare    func(a args, f *fields)
		args       args
		wantErr    error
		wantCached bool
	}{
		{
			name: "Success",
			args: args{
				ctx:   context.Background(),
				order: order,
			},
			prepare: func(a args, f *fields) {
				f.repoMock.GetOrderByUIDMock.Expect(a.ctx, order.OrderUID).Return(nil, repository.ErrOrderNotFound)
				f.transactorMock.WithinTransactionMock.Set(func(_ context.Context, fn func(context.Context) error) error {
					return fn(a.ctx)
				})
				f.repoMock.SaveOrderMock.Set(func(_ context.Context, o *models.Order) error {
					assert.Equal(t, order, o)
					return nil
				})
			},
			wantErr:    nil,
			wantCached: true,
		},
		{
			name: "Invalid Input",
			args: args{
				ctx:   context.Background(),
				order: &models.Order{},
			},
			prepare: func(a args, f *fields) {},
			wantErr: ErrInvalidInput,
		},
		{
			name: "Order Already Exists",
			args: args{
				ctx:   context.Background(),
				order: order,
			},
			prepare: func(a args, f *fields) {
				f.repoMock.GetOrderByUIDMock.Expect(a.ctx, order.OrderUID).Return(&models.Order{}, nil)
			},
			wantErr: ErrOrderAlreadyExists,
		},
		{
			name: "Repo Check Error",
			args: args{
				ctx:   context.Background(),
				order: order,
			},
			prepare: func(a args, f *fields) {
				f.repoMock.GetOrderByUIDMock.Expect(a.ctx, order.OrderUID).Return(nil, ErrDB)
			},
			wantErr: ErrDB,
		},
		{
			name: "Transaction Error",
			args: args{
				ctx:   context.Background(),
				order: order,
			},
			prepare: func(a args, f *fields) {
				f.repoMock.GetOrderByUIDMock.Expect(a.ctx, order.OrderUID).Return(nil, repository.ErrOrderNotFound)
				f.transactorMock.WithinTransactionMock.Return(ErrTx)
			},
			wantErr: ErrTx,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctrl := minimock.NewController(t)
			repoMock := rmocks.NewOrderRepositoryMock(ctrl)
			transactorMock := tmocks.NewTransactorMock(ctrl)

			s := &orderService{
				repo:       repoMock,
				transactor: transactorMock,
				log:        slog.Default(),
				cache:      make(map[string]*models.Order),
				val:        validator.New(),
			}

			tt.prepare(tt.args, &fields{
				repoMock:       repoMock,
				transactorMock: transactorMock,
			})

			err := s.AddOrder(tt.args.ctx, tt.args.order)
			if tt.wantErr != nil {
				assert.ErrorIs(t, err, tt.wantErr)
			} else {
				assert.NoError(t, err)
			}

			if tt.wantCached {
				s.mu.RLock()
				_, ok := s.cache[tt.args.order.OrderUID]
				s.mu.RUnlock()
				assert.True(t, ok)
			}
		})
	}
}

func Test_orderService_GetOrder(t *testing.T) {
	t.Parallel()
	order := &models.Order{OrderUID: "uid1"}

	type fields struct {
		repoMock       *rmocks.OrderRepositoryMock
		transactorMock *tmocks.TransactorMock
		cache          map[string]*models.Order
	}
	type args struct {
		ctx context.Context
		uid string
	}
	tests := []struct {
		name    string
		prepare func(a args, f *fields)
		args    args
		want    *models.Order
		wantErr error
	}{
		{
			name: "Cache Hit",
			args: args{
				ctx: context.Background(),
				uid: "uid1",
			},
			prepare: func(a args, f *fields) {
				f.cache["uid1"] = order
			},
			want: order,
		},
		{
			name: "Cache Miss - Found",
			args: args{
				ctx: context.Background(),
				uid: "uid1",
			},
			prepare: func(a args, f *fields) {
				f.repoMock.GetOrderByUIDMock.Expect(a.ctx, a.uid).Return(order, nil)
			},
			want: order,
		},
		{
			name: "Cache Miss - Not Found",
			args: args{
				ctx: context.Background(),
				uid: "uid1",
			},
			prepare: func(a args, f *fields) {
				f.repoMock.GetOrderByUIDMock.Expect(a.ctx, a.uid).Return(nil, repository.ErrOrderNotFound)
			},
			wantErr: repository.ErrOrderNotFound,
		},
		{
			name: "Cache Miss - Error",
			args: args{
				ctx: context.Background(),
				uid: "uid1",
			},
			prepare: func(a args, f *fields) {
				f.repoMock.GetOrderByUIDMock.Expect(a.ctx, a.uid).Return(nil, ErrDB)
			},
			wantErr: ErrDB,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctrl := minimock.NewController(t)
			repoMock := rmocks.NewOrderRepositoryMock(ctrl)
			transactorMock := tmocks.NewTransactorMock(ctrl)
			cache := make(map[string]*models.Order)

			s := &orderService{
				repo:       repoMock,
				transactor: transactorMock,
				log:        slog.Default(),
				cache:      cache,
				val:        validator.New(),
			}

			tt.prepare(tt.args, &fields{
				repoMock:       repoMock,
				transactorMock: transactorMock,
				cache:          cache,
			})

			got, err := s.GetOrder(tt.args.ctx, tt.args.uid)
			if tt.wantErr != nil {
				assert.ErrorIs(t, err, tt.wantErr)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.want, got)
			}
		})
	}
}

func Test_orderService_loadCache(t *testing.T) {
	t.Parallel()

	now := time.Now().UTC()
	orders := map[string]*models.Order{
		"uid1": {
			OrderUID:          "uid1",
			TrackNumber:       "track1",
			Entry:             "entry",
			Locale:            "en",
			InternalSignature: "",
			CustomerID:        "cust",
			DeliveryService:   "serv",
			Shardkey:          "9",
			SmID:              99,
			DateCreated:       now,
			OofShard:          "1",
			Delivery: models.Delivery{
				Name:    "name",
				Phone:   "+123",
				Zip:     "zip",
				City:    "city",
				Address: "addr",
				Region:  "region",
				Email:   "email@example.com",
			},
			Payment: models.Payment{
				Transaction:  "tx1",
				Currency:     "USD",
				Provider:     "prov",
				Amount:       100,
				PaymentDt:    now.Unix(),
				Bank:         "bank",
				DeliveryCost: 10,
				GoodsTotal:   90,
				CustomFee:    0,
			},
			Items: []models.Item{
				{
					ChrtID:      1,
					TrackNumber: "itemtrack",
					Price:       50,
					Name:        "item",
					Sale:        0,
					Size:        "M",
					TotalPrice:  50,
					NmID:        2,
					Brand:       "brand",
					Status:      200,
				},
			},
		},
		"uid2": {
			OrderUID:          "uid2",
			TrackNumber:       "track2",
			Entry:             "entry",
			Locale:            "en",
			InternalSignature: "",
			CustomerID:        "cust",
			DeliveryService:   "serv",
			Shardkey:          "9",
			SmID:              99,
			DateCreated:       now,
			OofShard:          "1",
			Delivery: models.Delivery{
				Name:    "name",
				Phone:   "+123",
				Zip:     "zip",
				City:    "city",
				Address: "addr",
				Region:  "region",
				Email:   "email@example.com",
			},
			Payment: models.Payment{
				Transaction:  "tx2",
				Currency:     "USD",
				Provider:     "prov",
				Amount:       200,
				PaymentDt:    now.Unix(),
				Bank:         "bank",
				DeliveryCost: 20,
				GoodsTotal:   180,
				CustomFee:    0,
			},
			Items: []models.Item{
				{
					ChrtID:      2,
					TrackNumber: "itemtrack2",
					Price:       100,
					Name:        "item2",
					Sale:        0,
					Size:        "L",
					TotalPrice:  100,
					NmID:        3,
					Brand:       "brand2",
					Status:      200,
				},
			},
		},
	}

	type fields struct {
		repoMock *rmocks.OrderRepositoryMock
	}
	type args struct {
		ctx context.Context
	}
	tests := []struct {
		name         string
		prepare      func(a args, f *fields)
		args         args
		wantCacheLen int
		wantErr      bool
	}{
		{
			name: "Success - Load Orders",
			args: args{
				ctx: context.Background(),
			},
			prepare: func(a args, f *fields) {
				f.repoMock.LoadAllOrdersMock.Expect(a.ctx).Return(orders, nil)
			},
			wantCacheLen: 2,
			wantErr:      false,
		},
		{
			name: "Error - DB Fail",
			args: args{
				ctx: context.Background(),
			},
			prepare: func(a args, f *fields) {
				f.repoMock.LoadAllOrdersMock.Expect(a.ctx).Return(nil, ErrDB)
			},
			wantCacheLen: 0,
			wantErr:      true,
		},
		{
			name: "Empty DB",
			args: args{
				ctx: context.Background(),
			},
			prepare: func(a args, f *fields) {
				f.repoMock.LoadAllOrdersMock.Expect(a.ctx).Return(make(map[string]*models.Order), nil)
			},
			wantCacheLen: 0,
			wantErr:      false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctrl := minimock.NewController(t)
			repoMock := rmocks.NewOrderRepositoryMock(ctrl)

			s := &orderService{
				repo:  repoMock,
				cache: make(map[string]*models.Order),
				log:   slog.Default(),
			}

			tt.prepare(tt.args, &fields{
				repoMock: repoMock,
			})

			s.loadCache(tt.args.ctx)

			s.mu.RLock()
			assert.Len(t, s.cache, tt.wantCacheLen)
			s.mu.RUnlock()
		})
	}
}
