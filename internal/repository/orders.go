package repository

import (
	"context"
	"github.com/jackc/pgx/v5"
	"time"
	"wbl0/internal/model"
	"wbl0/internal/queries"
)

func (r *Repository) UpsertOrder(ctx context.Context, order model.Order) error {
	tctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	tx, err := r.db.Begin(tctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(tctx)

	_, err = tx.Exec(tctx, queries.UpsertOrder, order.OrderUID, order.TrackNumber, order.Entry, order.Locale, order.InternalSignature, order.CustomerID,
		order.DeliveryService, order.ShardKey, order.SmID, order.DateCreated, order.OofShard)
	if err != nil {
		return err
	}

	_, err = tx.Exec(tctx, queries.UpsertDeliveries, order.OrderUID, order.Delivery.Name, order.Delivery.Phone, order.Delivery.Zip, order.Delivery.City,
		order.Delivery.Address, order.Delivery.Region, order.Delivery.Email)
	if err != nil {
		return err
	}

	_, err = tx.Exec(tctx, queries.UpsertPayments, order.OrderUID, order.Payment.Transaction, order.Payment.RequestID, order.Payment.Currency, order.Payment.Provider,
		order.Payment.Amount, order.Payment.PaymentDT, order.Payment.Bank, order.Payment.DeliveryCost,
		order.Payment.GoodsTotal, order.Payment.CustomFee)
	if err != nil {
		return err
	}

	_, err = tx.Exec(tctx, queries.DeleteOrderItem, order.OrderUID)
	if err != nil {
		return err
	}

	var b pgx.Batch
	for _, item := range order.Items {
		b.Queue(queries.InsertOrderItem, order.OrderUID, item.ChrtID, item.TrackNumber, item.Price, item.RID, item.Name, item.Sale, item.Size, item.TotalPrice, item.NmID, item.Brand, item.Status)
	}
	res := tx.SendBatch(ctx, &b)
	if err = res.Close(); err != nil {
		return err
	}
	return tx.Commit(ctx)
}
