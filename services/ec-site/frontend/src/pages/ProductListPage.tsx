import { useState, useEffect } from "react";
import { Link } from "react-router";
import { createClient } from "@connectrpc/connect";
import { createConnectTransport } from "@connectrpc/connect-web";
import { ProductService, type Product } from "@bmkr/bff/gen/product/v1/product_pb.js";

const transport = createConnectTransport({
  baseUrl: "/",
});
const client = createClient(ProductService, transport);

// ProductListPage は商品一覧ページを表示するコンポーネント。
//
// 動作:
//   - マウント時に client.listProducts を呼び出して全商品を取得する
//   - 取得中は「読み込み中...」を表示する
//   - 取得成功時、商品名と価格を一覧表示する
//   - 各商品は /products/:id へのリンクを持ち、クリックで詳細ページに遷移する
//   - 商品が0件の場合、「商品がありません」メッセージを表示する
//   - 商品画像は No Image プレースホルダーを表示する
//
// エラー:
//   - API 呼び出しエラー時はエラーメッセージを表示する
//   - Error インスタンスの場合は message を、それ以外は String() で文字列化する
export function ProductListPage(): React.ReactElement {
  const [products, setProducts] = useState<Product[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    let cancelled = false;

    async function loadProducts(): Promise<void> {
      try {
        const response = await client.listProducts({});
        if (cancelled) {
          return;
        }
        setProducts(response.products);
        setError(null);
      } catch (err) {
        if (cancelled) {
          return;
        }
        setError(err instanceof Error ? err.message : String(err));
      } finally {
        if (!cancelled) {
          setLoading(false);
        }
      }
    }

    void loadProducts();

    return () => {
      cancelled = true;
    };
  }, []);

  if (loading) {
    return <div>読み込み中...</div>;
  }

  if (error !== null) {
    return <div>{error}</div>;
  }

  if (products.length === 0) {
    return <div>商品がありません</div>;
  }

  return (
    <div>
      <h1>商品一覧</h1>
      <ul className="list-none m-0 p-0 grid gap-4">
        {products.map((product) => (
          <li key={product.id.toString()} className="border border-border rounded-lg p-4">
            <Link
              to={`/products/${product.id.toString()}`}
              className="text-inherit no-underline grid gap-3"
            >
              <div className="bg-surface text-muted min-h-40 flex items-center justify-center rounded-md">
                No Image
              </div>
              <div>
                <div>{product.name}</div>
                <div>{product.price.toString()}円</div>
              </div>
            </Link>
          </li>
        ))}
      </ul>
    </div>
  );
}
