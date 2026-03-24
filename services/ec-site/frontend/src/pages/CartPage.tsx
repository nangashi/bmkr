import { useState, useEffect, useCallback } from "react";
import { Link } from "react-router";
import { createClient } from "@connectrpc/connect";
import { createConnectTransport } from "@connectrpc/connect-web";
import { CartService, type Cart } from "@bmkr/bff/gen/ec/v1/cart_pb.js";

const transport = createConnectTransport({
  baseUrl: "/",
});
const cartClient = createClient(CartService, transport);

// 固定の customer_id（認証スコープ外のため）
const CUSTOMER_ID = BigInt(1);

// CartPage はカートページを表示するコンポーネント。
export function CartPage(): React.ReactElement {
  const [cart, setCart] = useState<Cart | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    let cancelled = false;

    async function loadCart(): Promise<void> {
      try {
        const response = await cartClient.getCart({ customerId: CUSTOMER_ID });
        if (cancelled) return;
        setCart(response.cart ?? null);
        setError(null);
      } catch (err) {
        if (cancelled) return;
        setError(err instanceof Error ? err.message : String(err));
      } finally {
        if (!cancelled) setLoading(false);
      }
    }

    void loadCart();

    return () => {
      cancelled = true;
    };
  }, []);

  const handleUpdateQuantity = useCallback(async (itemId: bigint, quantity: number) => {
    if (quantity < 1) {
      setError("quantity must be >= 1");
      return;
    }

    try {
      const response = await cartClient.updateQuantity({
        customerId: CUSTOMER_ID,
        itemId,
        quantity,
      });
      setCart(response.cart ?? null);
      setError(null);
    } catch (err) {
      setError(err instanceof Error ? err.message : String(err));
    }
  }, []);

  const handleRemoveItem = useCallback(async (itemId: bigint) => {
    try {
      const response = await cartClient.removeItem({ customerId: CUSTOMER_ID, itemId });
      setCart(response.cart ?? null);
      setError(null);
    } catch (err) {
      setError(err instanceof Error ? err.message : String(err));
    }
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
      <ul style={{ listStyle: "none", margin: 0, padding: 0, display: "grid", gap: "12px" }}>
        {cart.items.map((item) => (
          <li
            key={item.id.toString()}
            style={{ border: "1px solid #d1d5db", borderRadius: "8px", padding: "16px" }}
          >
            <div>商品ID: {item.productId.toString()}</div>
            <div>数量: {item.quantity}</div>
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
