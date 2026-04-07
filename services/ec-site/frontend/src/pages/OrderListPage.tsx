import { useState, useEffect } from "react";
import { Link } from "react-router";
import { createClient } from "@connectrpc/connect";
import { createConnectTransport } from "@connectrpc/connect-web";
import { OrderService, type Order } from "@bmkr/bff/gen/ec/v1/order_pb.js";

const transport = createConnectTransport({
  baseUrl: "/",
});
const orderClient = createClient(OrderService, transport);

// 固定の customer_id（認証スコープ外のため）
const CUSTOMER_ID = BigInt(1);

export function OrderListPage(): React.ReactElement {
  const [orders, setOrders] = useState<Order[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    let cancelled = false;

    async function loadOrders(): Promise<void> {
      try {
        const response = await orderClient.listOrders({ customerId: CUSTOMER_ID });
        if (cancelled) return;
        setOrders(response.orders);
        setError(null);
      } catch (err) {
        if (cancelled) return;
        setError(err instanceof Error ? err.message : String(err));
      } finally {
        if (!cancelled) setLoading(false);
      }
    }

    void loadOrders();

    return () => {
      cancelled = true;
    };
  }, []);

  if (loading) {
    return <div>読み込み中...</div>;
  }

  if (error !== null) {
    return (
      <div className="grid gap-4">
        <p>{error}</p>
        <Link to="/">商品一覧に戻る</Link>
      </div>
    );
  }

  if (orders.length === 0) {
    return (
      <div className="grid gap-4">
        <h1>注文履歴</h1>
        <p>注文はありません</p>
        <Link to="/">商品一覧に戻る</Link>
      </div>
    );
  }

  return (
    <div className="grid gap-4">
      <Link to="/">商品一覧に戻る</Link>
      <h1>注文履歴</h1>
      <ul className="list-none m-0 p-0 grid gap-3">
        {orders.map((order) => (
          <li key={order.id.toString()} className="border border-border rounded-lg p-4">
            <Link to={`/orders/${order.id.toString()}`} className="grid gap-1 no-underline">
              <div>注文ID: {order.id.toString()}</div>
              <div>
                日時:{" "}
                {order.createdAt != null
                  ? new Date(Number(order.createdAt.seconds) * 1000).toLocaleString("ja-JP")
                  : "-"}
              </div>
              <div>合計金額: {order.totalAmount.toString()}円</div>
              <div>ステータス: {order.status}</div>
            </Link>
          </li>
        ))}
      </ul>
    </div>
  );
}
