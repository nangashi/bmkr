import Fastify from "fastify";
import { toJson } from "@bufbuild/protobuf";
import { createClient } from "@connectrpc/connect";
import { createConnectTransport } from "@connectrpc/connect-node";
import {
  ProductService,
  GetProductResponseSchema,
  ListProductsResponseSchema,
  BatchGetProductsResponseSchema,
} from "../gen/product/v1/product_pb.js";
import {
  CartService,
  GetCartResponseSchema,
  AddItemResponseSchema,
  RemoveItemResponseSchema,
  UpdateQuantityResponseSchema,
} from "../gen/ec/v1/cart_pb.js";

const port = Number(process.env.PORT) || 3000;
const productServiceURL = process.env.PRODUCT_SERVICE_URL || "http://localhost:8081";
const ecSiteServiceURL = process.env.EC_SITE_SERVICE_URL || "http://localhost:8080";

const productTransport = createConnectTransport({
  baseUrl: productServiceURL,
  httpVersion: "2",
});
const productClient = createClient(ProductService, productTransport);

const ecSiteTransport = createConnectTransport({
  baseUrl: ecSiteServiceURL,
  httpVersion: "2",
});
const cartClient = createClient(CartService, ecSiteTransport);

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

// Connect RPC proxy: BatchGetProducts
// wip: リクエストボディの ids フィールド（BigInt 配列）を ProductService.BatchGetProducts に転送する
//   - ids が空または未指定の場合、ProductService 側で INVALID_ARGUMENT が返る
//   - ids の要素が整数パース不能な場合、BigInt() が例外を投げ Fastify が 500 を返す
//   - ProductService への接続エラーは Fastify のデフォルトエラーハンドリングに任せる
app.post("/product.v1.ProductService/BatchGetProducts", async (req, reply) => {
  const body = req.body as { ids?: (string | number)[] };
  const ids = (body.ids ?? []).map((id) => BigInt(id));
  const resp = await productClient.batchGetProducts({ ids });
  reply.header("Content-Type", "application/json");
  return toJson(BatchGetProductsResponseSchema, resp);
});

// Connect RPC proxy: GetCart
// wip: リクエストボディの customer_id を CartService.GetCart に転送する
//   - customer_id が未指定の場合、デフォルト 0n が使われる
//   - customerId が整数パース不能な場合、BigInt() が例外を投げ Fastify が 500 を返す
//   - ec-site サービスへの接続エラーは Fastify のデフォルトエラーハンドリングに任せる
app.post("/ec.v1.CartService/GetCart", async (req, reply) => {
  const body = req.body as { customerId?: string | number };
  const customerId = BigInt(body.customerId ?? 0);
  const resp = await cartClient.getCart({ customerId });
  reply.header("Content-Type", "application/json");
  return toJson(GetCartResponseSchema, resp);
});

// Connect RPC proxy: AddItem
// wip: リクエストボディの customer_id, product_id, quantity を CartService.AddItem に転送する
//   - quantity が未指定の場合、デフォルト 0 が使われ、バックエンドで INVALID_ARGUMENT が返る
//   - product_id が存在しない場合、バックエンドで NOT_FOUND が返る
//   - customerId / productId が整数パース不能な場合、BigInt() が例外を投げ Fastify が 500 を返す
//   - ec-site サービスへの接続エラーは Fastify のデフォルトエラーハンドリングに任せる
app.post("/ec.v1.CartService/AddItem", async (req, reply) => {
  const body = req.body as {
    customerId?: string | number;
    productId?: string | number;
    quantity?: number;
  };
  const customerId = BigInt(body.customerId ?? 0);
  const productId = BigInt(body.productId ?? 0);
  const quantity = Number(body.quantity ?? 0);
  const resp = await cartClient.addItem({ customerId, productId, quantity });
  reply.header("Content-Type", "application/json");
  return toJson(AddItemResponseSchema, resp);
});

// Connect RPC proxy: RemoveItem
// wip: リクエストボディの customer_id, item_id を CartService.RemoveItem に転送する
//   - item_id が存在しない場合、バックエンドで NOT_FOUND が返る
//   - customerId / itemId が整数パース不能な場合、BigInt() が例外を投げ Fastify が 500 を返す
//   - ec-site サービスへの接続エラーは Fastify のデフォルトエラーハンドリングに任せる
app.post("/ec.v1.CartService/RemoveItem", async (req, reply) => {
  const body = req.body as { customerId?: string | number; itemId?: string | number };
  const customerId = BigInt(body.customerId ?? 0);
  const itemId = BigInt(body.itemId ?? 0);
  const resp = await cartClient.removeItem({ customerId, itemId });
  reply.header("Content-Type", "application/json");
  return toJson(RemoveItemResponseSchema, resp);
});

// Connect RPC proxy: UpdateQuantity
// wip: リクエストボディの customer_id, item_id, quantity を CartService.UpdateQuantity に転送する
//   - quantity が 1 未満の場合、バックエンドで INVALID_ARGUMENT が返る
//   - item_id が存在しない場合、バックエンドで NOT_FOUND が返る
//   - customerId / itemId が整数パース不能な場合、BigInt() が例外を投げ Fastify が 500 を返す
//   - ec-site サービスへの接続エラーは Fastify のデフォルトエラーハンドリングに任せる
app.post("/ec.v1.CartService/UpdateQuantity", async (req, reply) => {
  const body = req.body as {
    customerId?: string | number;
    itemId?: string | number;
    quantity?: number;
  };
  const customerId = BigInt(body.customerId ?? 0);
  const itemId = BigInt(body.itemId ?? 0);
  const quantity = Number(body.quantity ?? 0);
  const resp = await cartClient.updateQuantity({ customerId, itemId, quantity });
  reply.header("Content-Type", "application/json");
  return toJson(UpdateQuantityResponseSchema, resp);
});

app.listen({ port, host: "0.0.0.0" }, (err) => {
  if (err) {
    app.log.error(err);
    process.exit(1);
  }
});
