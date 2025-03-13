package postgres

import (
	"context"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
	pb "lyceum/pkg/api"
	"lyceum/pkg/logger"
)

type PG struct {
	*pgxpool.Pool
}

func NewPG(p *pgxpool.Pool) *PG {
	return &PG{p}
}

func (p *PG) Create(ctx context.Context, id string, item string, quantity int32) {
	tx, err := p.Pool.Begin(ctx)
	if err != nil {
		logger.GetLogger(ctx).Error(ctx, "PG Begin Error", zap.Error(err))
		return
	}
	_, err = p.Exec(ctx, "INSERT INTO lyceum_schema.orders (id, item, quantity) VALUES ($1, $2, $3)", id, item, quantity)
	if err != nil {
		logger.GetLogger(ctx).Error(ctx, "Error create order", zap.Error(err))
		return
	}
	err = tx.Commit(ctx)
	if err != nil {
		logger.GetLogger(ctx).Error(ctx, "Error commit order", zap.Error(err))
		return
	}

}

func (p *PG) Get(ctx context.Context, id string) (*pb.Order, error) {
	order := &pb.Order{}
	row := p.Pool.QueryRow(ctx, "SELECT id, item, quantity FROM lyceum_schema.orders WHERE id = $1", id)
	if err := row.Scan(&order.Id, &order.Item, &order.Quantity); err != nil {
		logger.GetLogger(ctx).Error(ctx, "Error get order", zap.Error(err))
	}
	return order, nil
}

func (p *PG) Update(ctx context.Context, id string, item string, quantity int32) (*pb.Order, error) {
	_, err := p.Get(ctx, id)
	if err != nil {
		return nil, err
	}
	tx, err := p.Pool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	_, err = p.Pool.Exec(ctx, "UPDATE lyceum_schema.orders SET item = $1, quantity = $2 WHERE id = $3", item, quantity, id)
	if err != nil {
		return nil, err
	}
	err = tx.Commit(ctx)
	if err != nil {
		return nil, err
	}
	return p.Get(ctx, id)
}

func (p *PG) Delete(ctx context.Context, id string) error {
	_, err := p.Get(ctx, id)
	if err != nil {
		return err
	}
	tx, err := p.Pool.Begin(ctx)
	if err != nil {
		return err
	}
	_, err = p.Pool.Exec(ctx, "DELETE FROM lyceum_schema.orders WHERE id = $1", id)
	if err != nil {
		return err
	}
	err = tx.Commit(ctx)
	if err != nil {
		return err
	}
	return nil
}

func (p *PG) List(ctx context.Context) []*pb.Order {
	var orders []*pb.Order
	r, err := p.Pool.Query(ctx, "SELECT id, item, quantity FROM lyceum_schema.orders")
	if err != nil {
		logger.GetLogger(ctx).Error(ctx, "PG Query Error", zap.Error(err))
		return orders
	}
	defer r.Close()
	for r.Next() {
		order := &pb.Order{}
		err = r.Scan(&order.Id, &order.Item, &order.Quantity)
		if err != nil {
			logger.GetLogger(ctx).Error(ctx, "PG Scan Error", zap.Error(err))
		}
		orders = append(orders, order)
	}
	return orders
}
