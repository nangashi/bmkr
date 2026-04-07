import { createClient } from "@connectrpc/connect";
import { createConnectTransport } from "@connectrpc/connect-web";
import { ProductService } from "@bmkr/bff/gen/product/v1/product_pb.js";
import { CartService } from "@bmkr/bff/gen/ec/v1/cart_pb.js";

const transport = createConnectTransport({ baseUrl: "/" });

export const productClient = createClient(ProductService, transport);
export const cartClient = createClient(CartService, transport);
