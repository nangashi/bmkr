import { useState, useEffect } from "react";
import { useParams, Link } from "react-router";
import { createClient } from "@connectrpc/connect";
import { createConnectTransport } from "@connectrpc/connect-web";
import { OrderService, type Order } from "@bmkr/bff/gen/ec/v1/order_pb.js";

const transport = createConnectTransport({
  baseUrl: "/",
});
const orderClient = createClient(OrderService, transport);

// 固定の customer_id（認証スコープ外のため）
const CUSTOMER_ID = BigInt(1);

export function OrderDetailPage(): React.ReactElement {
  const { id } = useParams<{ id: string }>();
  const [order, setOrder] = useState<Order | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    if (id === undefined) {
      setError("注文 ID が指定されていません");
      setLoading(false);
      return;
    }

    const orderId = id;
    let cancelled = false;

    async function loadOrder(): Promise<void> {
      try {
        const response = await orderClient.getOrder({
          customerId: CUSTOMER_ID,
          orderId: BigInt(orderId),
        });
        if (cancelled) return;
        setOrder(response.order ?? null);
        setError(null);
      } catch (err) {
        if (cancelled) return;
        setError(err instanceof Error ? err.message : String(err));
      } finally {
        if (!cancelled) setLoading(false);
      }
    }

    void loadOrder();

    return () => {
      cancelled = true;
    };
  }, [id]);

  if (loading) {
    return <div>読み込み中...</div>;
  }

  if (error !== null) {
    return (
      <div className="grid gap-4">
        <p>{error}</p>
        <Link to="/orders">注文一覧に戻る</Link>
      </div>
    );
  }

  if (order === null) {
    return (
      <div className="grid gap-4">
        <p>注文が見つかりません</p>
        <Link to="/orders">注文一覧に戻る</Link>
      </div>
    );
  }

  return (
    <div className="grid gap-4">
      <Link to="/orders">注文一覧に戻る</Link>
      <h1>注文詳細</h1>
      <div className="border border-border rounded-lg p-4 grid gap-2">
        <div>注文ID: {order.id.toString()}</div>
        <div>
          日時:{" "}
          {order.createdAt != null
            ? new Date(Number(order.createdAt.seconds) * 1000).toLocaleString("ja-JP")
            : "-"}
        </div>
        <div>合計金額: {order.totalAmount.toString()}円</div>
        <div>ステータス: {order.status}</div>
      </div>
      <h2 className="text-xl font-bold">注文アイテム</h2>
      <ul className="list-none m-0 p-0 grid gap-3">
        {order.items.map((item) => (
          <li key={item.id.toString()} className="border border-border rounded-lg p-4 grid gap-1">
            <div>{item.productName}</div>
            <div>価格: {item.price.toString()}円</div>
            <div>数量: {item.quantity}</div>
          </li>
        ))}
      </ul>
    </div>
  );
}
