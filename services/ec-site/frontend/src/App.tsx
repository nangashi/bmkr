import { useState } from "react";
import { createClient } from "@connectrpc/connect";
import { createConnectTransport } from "@connectrpc/connect-web";
import { ProductService, type Product } from "@bmkr/bff/gen/product/v1/product_pb.js";

const transport = createConnectTransport({
  baseUrl: "/",
});
const client = createClient(ProductService, transport);

function App() {
  const [productId, setProductId] = useState("1");
  const [product, setProduct] = useState<Product | null>(null);
  const [error, setError] = useState<string | null>(null);
  const [loading, setLoading] = useState(false);

  const fetchProduct = async () => {
    setError(null);
    setLoading(true);
    try {
      const resp = await client.getProduct({ id: BigInt(productId) });
      setProduct(resp.product ?? null);
    } catch (e) {
      setError(e instanceof Error ? e.message : String(e));
      setProduct(null);
    } finally {
      setLoading(false);
    }
  };

  return (
    <div style={{ padding: "2rem", fontFamily: "sans-serif" }}>
      <h1>bmkr</h1>
      <div style={{ marginBottom: "1rem" }}>
        <label>
          商品 ID:{" "}
          <input
            type="number"
            value={productId}
            onChange={(e) => setProductId(e.target.value)}
            min="1"
            style={{ marginRight: "0.5rem" }}
          />
        </label>
        <button onClick={fetchProduct} disabled={loading}>
          {loading ? "取得中..." : "取得"}
        </button>
      </div>
      {error && <p style={{ color: "red" }}>{error}</p>}
      {product && (
        <table style={{ borderCollapse: "collapse", border: "1px solid #ccc" }}>
          <tbody>
            <tr>
              <th style={thStyle}>ID</th>
              <td style={tdStyle}>{product.id.toString()}</td>
            </tr>
            <tr>
              <th style={thStyle}>名前</th>
              <td style={tdStyle}>{product.name}</td>
            </tr>
            <tr>
              <th style={thStyle}>説明</th>
              <td style={tdStyle}>{product.description}</td>
            </tr>
            <tr>
              <th style={thStyle}>価格</th>
              <td style={tdStyle}>{product.price.toString()} 円</td>
            </tr>
            <tr>
              <th style={thStyle}>在庫数</th>
              <td style={tdStyle}>{product.stockQuantity.toString()}</td>
            </tr>
          </tbody>
        </table>
      )}
    </div>
  );
}

const thStyle: React.CSSProperties = {
  border: "1px solid #ccc",
  padding: "0.5rem 1rem",
  textAlign: "left",
  background: "#f5f5f5",
};

const tdStyle: React.CSSProperties = {
  border: "1px solid #ccc",
  padding: "0.5rem 1rem",
};

export default App;
