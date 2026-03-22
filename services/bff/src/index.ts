import Fastify from "fastify";
import { toJson } from "@bufbuild/protobuf";
import { createClient } from "@connectrpc/connect";
import { createConnectTransport } from "@connectrpc/connect-node";
import {
  ProductService,
  GetProductResponseSchema,
  ListProductsResponseSchema,
} from "../gen/product/v1/product_pb.js";

const port = Number(process.env.PORT) || 3000;
const productServiceURL = process.env.PRODUCT_SERVICE_URL || "http://localhost:8081";

const productTransport = createConnectTransport({
  baseUrl: productServiceURL,
  httpVersion: "2",
});
const productClient = createClient(ProductService, productTransport);

const app = Fastify({ logger: true });

app.get("/health", async () => {
  return { status: "ok" };
});

// Connect RPC proxy: accepts Connect protocol JSON requests
app.post("/product.v1.ProductService/GetProduct", async (req, reply) => {
  const body = req.body as { id?: string | number };
  const id = BigInt(body.id ?? 0);
  const resp = await productClient.getProduct({ id });
  reply.header("Content-Type", "application/json");
  return toJson(GetProductResponseSchema, resp);
});

// Connect RPC proxy: ListProducts
// リクエストボディは空オブジェクト {} を期待する。
//
// 動作:
//   - productClient.listProducts を呼び出して全商品を取得する
//   - ListProductsResponseSchema を使って JSON にシリアライズして返す
//   - 商品が0件の場合、空の products 配列を持つレスポンスを返す（エラーにしない）
//
// エラー:
//   - product-mgmt サービスへの接続エラーは Fastify のデフォルトエラーハンドリングに任せる
//     （Fastify が 500 を返す）
app.post("/product.v1.ProductService/ListProducts", async (_req, reply) => {
  const resp = await productClient.listProducts({});
  reply.header("Content-Type", "application/json");
  return toJson(ListProductsResponseSchema, resp);
});

app.listen({ port, host: "0.0.0.0" }, (err) => {
  if (err) {
    app.log.error(err);
    process.exit(1);
  }
});
