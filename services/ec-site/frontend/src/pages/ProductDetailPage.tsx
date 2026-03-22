import { useState, useEffect } from "react";
import { useParams, Link } from "react-router";
import { createClient } from "@connectrpc/connect";
import { createConnectTransport } from "@connectrpc/connect-web";
import { ProductService, type Product } from "@bmkr/bff/gen/product/v1/product_pb.js";

const transport = createConnectTransport({
  baseUrl: "/",
});
const _client = createClient(ProductService, transport);
void Link;

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
  const [_product, _setProduct] = useState<Product | null>(null);
  const [_loading, _setLoading] = useState(true);
  const [_error, _setError] = useState<string | null>(null);

  useEffect(() => {
    // TODO: implement - id が存在する場合、client.getProduct を呼び出し state を更新する
  }, [id]);

  // TODO: implement - loading / error / product の分岐レンダリング
  return <div>not implemented</div>;
}
