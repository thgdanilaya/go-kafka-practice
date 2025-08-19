package queries

const UpsertOrder = `
INSERT INTO orders (
  order_uid, track_number, entry, locale, internal_signature, customer_id,
  delivery_service, shardkey, sm_id, date_created, oof_shard
) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11)
ON CONFLICT (order_uid) DO UPDATE SET
  track_number=EXCLUDED.track_number,
  entry=EXCLUDED.entry,
  locale=EXCLUDED.locale,
  internal_signature=EXCLUDED.internal_signature,
  customer_id=EXCLUDED.customer_id,
  delivery_service=EXCLUDED.delivery_service,
  shardkey=EXCLUDED.shardkey,
  sm_id=EXCLUDED.sm_id,
  date_created=EXCLUDED.date_created,
  oof_shard=EXCLUDED.oof_shard
`
const UpsertDeliveries = `
INSERT INTO deliveries (
  order_uid, name, phone, zip, city, address, region, email
) VALUES ($1,$2,$3,$4,$5,$6,$7,$8)
ON CONFLICT (order_uid) DO UPDATE SET
  name=EXCLUDED.name, phone=EXCLUDED.phone, zip=EXCLUDED.zip,
  city=EXCLUDED.city, address=EXCLUDED.address,
  region=EXCLUDED.region, email=EXCLUDED.email
`

const UpsertPayments = `
INSERT INTO payments (
  order_uid, transaction, request_id, currency, provider, amount,
  payment_dt, bank, delivery_cost, goods_total, custom_fee
) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11)
ON CONFLICT (order_uid) DO UPDATE SET
  transaction=EXCLUDED.transaction, request_id=EXCLUDED.request_id,
  currency=EXCLUDED.currency, provider=EXCLUDED.provider, amount=EXCLUDED.amount,
  payment_dt=EXCLUDED.payment_dt, bank=EXCLUDED.bank,
  delivery_cost=EXCLUDED.delivery_cost, goods_total=EXCLUDED.goods_total, custom_fee=EXCLUDED.custom_fee
`
const DeleteOrderItem = `
DELETE FROM order_items WHERE order_uid=$1
`
const InsertOrderItem = `
INSERT INTO order_items (
  order_uid, chrt_id, track_number, price, rid, name, sale, size, total_price, nm_id, brand, status
) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12)
`
