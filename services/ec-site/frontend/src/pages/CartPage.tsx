import { useState, useEffect, useCallback } from "react";
import { Link } from "react-router";
import { type Cart } from "@bmkr/bff/gen/ec/v1/cart_pb.js";
import { cartClient } from "../api/client.js";
import { CUSTOMER_ID } from "../constants.js";

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
    <div className="grid gap-4">
      <Link to="/">商品一覧に戻る</Link>
      <h1>カート</h1>
      <ul className="list-none m-0 p-0 grid gap-3">
        {cart.items.map((item) => (
          <li key={item.id.toString()} className="border border-border rounded-lg p-4">
            <div>商品ID: {item.productId.toString()}</div>
            <div>数量: {item.quantity}</div>
            <button
              className="rounded border border-border bg-gray-100 px-4 py-1"
              onClick={() => void handleUpdateQuantity(item.id, item.quantity + 1)}
            >
              +
            </button>
            <button
              className="rounded border border-border bg-gray-100 px-4 py-1 disabled:opacity-50"
              onClick={() => void handleUpdateQuantity(item.id, item.quantity - 1)}
              disabled={item.quantity <= 1}
            >
              -
            </button>
            <button
              className="rounded border border-border bg-gray-100 px-4 py-1"
              onClick={() => void handleRemoveItem(item.id)}
            >
              削除
            </button>
          </li>
        ))}
      </ul>
    </div>
  );
}
