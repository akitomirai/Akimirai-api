ALTER TABLE promo_codes
  ADD COLUMN IF NOT EXISTS discount_percent DECIMAL(5,2) NOT NULL DEFAULT 0;

ALTER TABLE payment_orders
  ADD COLUMN IF NOT EXISTS promo_code VARCHAR(32),
  ADD COLUMN IF NOT EXISTS discount_percent DECIMAL(5,2) NOT NULL DEFAULT 0,
  ADD COLUMN IF NOT EXISTS discount_amount DECIMAL(20,2) NOT NULL DEFAULT 0;

CREATE INDEX IF NOT EXISTS idx_payment_orders_promo_code ON payment_orders(promo_code);

COMMENT ON COLUMN promo_codes.discount_percent IS 'Payment discount percent; 0 means registration bonus only.';
COMMENT ON COLUMN payment_orders.promo_code IS 'Promo code applied to this payment order.';
COMMENT ON COLUMN payment_orders.discount_percent IS 'Promo discount percent captured at order creation.';
COMMENT ON COLUMN payment_orders.discount_amount IS 'Discount amount captured at order creation.';
