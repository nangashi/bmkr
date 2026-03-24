import { useState, useEffect, useCallback } from "react";
import { Link } from "react-router";
import { createClient } from "@connectrpc/connect";
import { createConnectTransport } from "@connectrpc/connect-web";
import { CartService, type Cart } from "@bmkr/bff/gen/ec/v1/cart_pb.js";

const transport = createConnectTransport({
  baseUrl: "/",
});
const cartClient = createClient(CartService, transport);

// wip: 固定の customer_id（認証スコープ外のため）
const CUSTOMER_ID = BigInt(1);

// CartPage はカートページを表示するコンポーネント。
//
// wip: 正常系フロー
//   - マウント時に cartClient.getCart({ customerId: CUSTOMER_ID }) を呼び出してカート内容を取得する
//   - 取得中は「読み込み中...」を表示する
//   - 取得成功時、カート内の各アイテム（商品ID、数量）を一覧表示する
//   - 各アイテムに数量変更フォーム（+/-ボタンまたは入力）と削除ボタンを表示する
//   - カートが空の場合、「カートは空です」メッセージと商品一覧へのリンクを表示する
//
// wip: 数量変更
//   - 数量変更時に cartClient.updateQuantity({ customerId, itemId, quantity }) を呼び出す
//   - 成功時、レスポンスのカート内容で state を更新する
//   - quantity < 1 の場合、バックエンドが INVALID_ARGUMENT を返す（フロント側でもバリデーション可）
//
// wip: アイテム削除
//   - 削除ボタン押下時に cartClient.removeItem({ customerId, itemId }) を呼び出す
//   - 成功時、レスポンスのカート内容で state を更新する
//   - 存在しない item_id の場合、バックエンドが NOT_FOUND を返す
//
// wip: エラー
//   - API 呼び出しエラー時はエラーメッセージを表示する
//   - Error インスタンスの場合は message を、それ以外は String() で文字列化する
export function CartPage(): React.ReactElement {
  const [cart, setCart] = useState<Cart | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  const loadCart = useCallback(async () => {
    try {
      const response = await cartClient.getCart({ customerId: CUSTOMER_ID });
      setCart(response.cart ?? null);
      setError(null);
    } catch (err) {
      setError(err instanceof Error ? err.message : String(err));
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => {
    void loadCart();
  }, [loadCart]);

  // wip: handleUpdateQuantity は数量変更ハンドラ
  //   - itemId と新しい quantity を受け取る
  //   - cartClient.updateQuantity を呼び出す
  //   - 成功時、レスポンスの cart で state を更新する
  //   - エラー時、error state を更新する
  const handleUpdateQuantity = useCallback(async (_itemId: bigint, _quantity: number) => {
    // wip: 実装予定
  }, []);

  // wip: handleRemoveItem はアイテム削除ハンドラ
  //   - itemId を受け取る
  //   - cartClient.removeItem を呼び出す
  //   - 成功時、レスポンスの cart で state を更新する
  //   - エラー時、error state を更新する
  const handleRemoveItem = useCallback(async (_itemId: bigint) => {
    // wip: 実装予定
  }, []);

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

  if (cart === null || cart.items.length === 0) {
    return (
      <div>
        <h1>カート</h1>
        <p>カートは空です</p>
        <Link to="/">商品一覧に戻る</Link>
      </div>
    );
  }

  return (
    <div style={{ display: "grid", gap: "16px" }}>
      <Link to="/">商品一覧に戻る</Link>
      <h1>カート</h1>
      {/* wip: カートアイテム一覧を表示する。各アイテムに商品ID、数量、数量変更UI、削除ボタンを含む */}
      <ul style={{ listStyle: "none", margin: 0, padding: 0, display: "grid", gap: "12px" }}>
        {cart.items.map((item) => (
          <li
            key={item.id.toString()}
            style={{ border: "1px solid #d1d5db", borderRadius: "8px", padding: "16px" }}
          >
            <div>商品ID: {item.productId.toString()}</div>
            <div>数量: {item.quantity}</div>
            {/* wip: 数量変更ボタン (+/-) と削除ボタンを配置する */}
            <button onClick={() => void handleUpdateQuantity(item.id, item.quantity + 1)}>+</button>
            <button
              onClick={() => void handleUpdateQuantity(item.id, item.quantity - 1)}
              disabled={item.quantity <= 1}
            >
              -
            </button>
            <button onClick={() => void handleRemoveItem(item.id)}>削除</button>
          </li>
        ))}
      </ul>
    </div>
  );
}
