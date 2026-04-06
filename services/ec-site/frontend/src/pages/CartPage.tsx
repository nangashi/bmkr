import { useState, useEffect, useCallback } from "react";
import { Link, useNavigate } from "react-router";
import { createClient, ConnectError, Code } from "@connectrpc/connect";
import { createConnectTransport } from "@connectrpc/connect-web";
import { CartService, type Cart } from "@bmkr/bff/gen/ec/v1/cart_pb.js";
import { OrderService } from "@bmkr/bff/gen/ec/v1/order_pb.js";

const transport = createConnectTransport({
  baseUrl: "/",
});
const cartClient = createClient(CartService, transport);
const orderClient = createClient(OrderService, transport);

// 固定の customer_id（認証スコープ外のため）
const CUSTOMER_ID = BigInt(1);

// CartPage はカートページを表示するコンポーネント。
export function CartPage(): React.ReactElement {
  const navigate = useNavigate();
  const [cart, setCart] = useState<Cart | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [placingOrder, setPlacingOrder] = useState(false);
  const [orderError, setOrderError] = useState<string | null>(null);

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

  const handlePlaceOrder = useCallback(async () => {
    setPlacingOrder(true);
    setOrderError(null);
    try {
      const response = await orderClient.placeOrder({ customerId: CUSTOMER_ID });
      const orderId = response.order?.id;
      if (orderId !== undefined) {
        void navigate(`/orders/${orderId.toString()}`);
      }
    } catch (err) {
      if (err instanceof ConnectError && err.code === Code.ResourceExhausted) {
        setOrderError("在庫不足のため注文できませんでした");
      } else {
        setOrderError(err instanceof Error ? err.message : String(err));
      }
    } finally {
      setPlacingOrder(false);
    }
  }, [navigate]);

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
      <div className="grid gap-4">
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
          <li key={item.id.toString()} className="border border-border rounded-lg p-4 grid gap-2">
            <div>商品ID: {item.productId.toString()}</div>
            <div>数量: {item.quantity}</div>
            <div className="flex gap-2">
              <button
                className="rounded border border-border bg-gray-100 px-3 py-1"
                onClick={() => void handleUpdateQuantity(item.id, item.quantity + 1)}
              >
                +
              </button>
              <button
                className="rounded border border-border bg-gray-100 px-3 py-1 disabled:opacity-50"
                onClick={() => void handleUpdateQuantity(item.id, item.quantity - 1)}
                disabled={item.quantity <= 1}
              >
                -
              </button>
              <button
                className="rounded border border-border bg-gray-100 px-3 py-1"
                onClick={() => void handleRemoveItem(item.id)}
              >
                削除
              </button>
            </div>
          </li>
        ))}
      </ul>
      {orderError !== null && <p className="text-red-600">{orderError}</p>}
      <button
        className="rounded border border-border bg-blue-600 text-white px-4 py-2 disabled:opacity-50"
        onClick={() => void handlePlaceOrder()}
        disabled={placingOrder}
      >
        {placingOrder ? "処理中..." : "注文を確定する"}
      </button>
    </div>
  );
}
