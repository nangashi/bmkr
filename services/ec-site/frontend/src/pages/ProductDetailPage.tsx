import { useState, useEffect, useCallback } from "react";
import { useParams, Link } from "react-router";
import { createClient } from "@connectrpc/connect";
import { createConnectTransport } from "@connectrpc/connect-web";
import { ProductService, type Product } from "@bmkr/bff/gen/product/v1/product_pb.js";
import { CartService } from "@bmkr/bff/gen/ec/v1/cart_pb.js";

const transport = createConnectTransport({
  baseUrl: "/",
});
const client = createClient(ProductService, transport);
const cartClient = createClient(CartService, transport);

// wip: 固定の customer_id（認証スコープ外のため）
const CUSTOMER_ID = BigInt(1);

// ProductDetailPage は商品詳細ページを表示するコンポーネント。
//
// 動作:
//   - URL パラメータ :id から商品 ID を取得する
//   - マウント時に client.getProduct({ id: BigInt(id) }) を呼び出して商品を取得する
//   - 取得中は「読み込み中...」を表示する
//   - 取得成功時、商品名・説明・価格・在庫数を表示する
//   - 商品画像は No Image プレースホルダーを表示する
//   - 商品一覧ページ (/) に戻るリンクを表示する
//
// エラー:
//   - id パラメータが未指定の場合、エラーメッセージを表示する
//   - API 呼び出しエラー時（商品が見つからない場合を含む）はエラーメッセージを表示する
//   - Error インスタンスの場合は message を、それ以外は String() で文字列化する
export function ProductDetailPage(): React.ReactElement {
  const { id } = useParams<{ id: string }>();
  const [product, setProduct] = useState<Product | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [addingToCart, setAddingToCart] = useState(false);
  const [cartMessage, setCartMessage] = useState<string | null>(null);

  // wip: handleAddToCart はカートに商品を追加するハンドラ
  //   - cartClient.addItem({ customerId: CUSTOMER_ID, productId: product.id, quantity: 1 }) を呼び出す
  //   - 成功時、「カートに追加しました」メッセージを表示する
  //   - エラー時、エラーメッセージを表示する
  //   - 呼び出し中はボタンを無効化する（addingToCart state）
  const handleAddToCart = useCallback(async () => {
    if (product === null) return;
    setAddingToCart(true);
    setCartMessage(null);
    try {
      await cartClient.addItem({
        customerId: CUSTOMER_ID,
        productId: product.id,
        quantity: 1,
      });
      setCartMessage("カートに追加しました");
    } catch (err) {
      setCartMessage(err instanceof Error ? err.message : String(err));
    } finally {
      setAddingToCart(false);
    }
  }, [product]);

  useEffect(() => {
    if (id === undefined) {
      setError("商品 ID が指定されていません");
      setLoading(false);
      return;
    }

    const productId = id;
    let cancelled = false;

    async function loadProduct(): Promise<void> {
      try {
        const response = await client.getProduct({ id: BigInt(productId) });
        if (cancelled) {
          return;
        }
        setProduct(response.product ?? null);
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

    void loadProduct();

    return () => {
      cancelled = true;
    };
  }, [id]);

  if (loading) {
    return <div>読み込み中...</div>;
  }

  if (error !== null) {
    return (
      <div>
        <p>{error}</p>
        <Link to="/">商品一覧に戻る</Link>
      </div>
    );
  }

  if (product === null) {
    return (
      <div>
        <p>商品が見つかりません</p>
        <Link to="/">商品一覧に戻る</Link>
      </div>
    );
  }

  return (
    <div style={{ display: "grid", gap: "16px" }}>
      <Link to="/">商品一覧に戻る</Link>
      <div
        style={{
          backgroundColor: "#e5e7eb",
          color: "#6b7280",
          minHeight: "240px",
          display: "flex",
          alignItems: "center",
          justifyContent: "center",
          borderRadius: "8px",
        }}
      >
        No Image
      </div>
      <div style={{ display: "grid", gap: "8px" }}>
        <h1>{product.name}</h1>
        <p>{product.description}</p>
        <p>価格: {product.price.toString()}円</p>
        <p>在庫数: {product.stockQuantity.toString()}</p>
        {/* wip: カートに追加ボタン。クリック時に handleAddToCart を呼び出す。追加中はボタンを無効化する。 */}
        <button onClick={() => void handleAddToCart()} disabled={addingToCart}>
          {addingToCart ? "追加中..." : "カートに追加"}
        </button>
        {cartMessage !== null && <p>{cartMessage}</p>}
        <Link to="/cart">カートを見る</Link>
      </div>
    </div>
  );
}
