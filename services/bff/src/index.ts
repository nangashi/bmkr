import Fastify from "fastify";
import { toJson } from "@bufbuild/protobuf";
import { createClient } from "@connectrpc/connect";
import { createConnectTransport } from "@connectrpc/connect-node";
import { ProductService, GetProductResponseSchema } from "../gen/product/v1/product_pb.js";

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

app.listen({ port, host: "0.0.0.0" }, (err) => {
  if (err) {
    app.log.error(err);
    process.exit(1);
  }
});
