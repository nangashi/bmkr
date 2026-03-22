-- +goose Up
-- 商品シードデータ: 起動後すぐに商品一覧を確認できるようにする。
--
-- 動作:
--   - 5件の商品データを挿入する
--   - id は BIGSERIAL なので自動採番される
--   - description, price, stock_quantity はバリエーションを持たせる
--   - created_at, updated_at はデフォルトの now() を使用する
INSERT INTO products (name, description, price, stock_quantity) VALUES
  ('Tシャツ', 'シンプルな無地のTシャツです。', 2500, 100),
  ('デニムパンツ', 'ストレートフィットのデニムパンツです。', 6800, 50),
  ('スニーカー', '軽量で履き心地の良いスニーカーです。', 9800, 30),
  ('トートバッグ', 'A4サイズが入る大容量トートバッグです。', 3500, 80),
  ('キャップ', 'コットン素材のベースボールキャップです。', 1800, 120);

-- +goose Down
DELETE FROM products WHERE name IN ('Tシャツ', 'デニムパンツ', 'スニーカー', 'トートバッグ', 'キャップ');
