import { useState, useEffect } from "react";
import { Link } from "react-router";
import { createClient } from "@connectrpc/connect";
import { createConnectTransport } from "@connectrpc/connect-web";
import { ProductService, type Product } from "@bmkr/bff/gen/product/v1/product_pb.js";

const transport = createConnectTransport({
  baseUrl: "/",
});
const _client = createClient(ProductService, transport);
void Link;

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
  const [_products, _setProducts] = useState<Product[]>([]);
  const [_loading, _setLoading] = useState(true);
  const [_error, _setError] = useState<string | null>(null);

  useEffect(() => {
    // TODO: implement - client.listProducts({}) を呼び出し、state を更新する
  }, []);

  // TODO: implement - loading / error / empty / list の分岐レンダリング
  return <div>not implemented</div>;
}
