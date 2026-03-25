import { defineConfig } from "vite";
import react from "@vitejs/plugin-react";
import tailwindcss from "@tailwindcss/vite";

export default defineConfig({
  plugins: [tailwindcss(), react()],
  server: {
    proxy: {
      "/product.v1.ProductService": {
        target: "http://localhost:3000",
      },
      "/ec.v1.CartService": {
        target: "http://localhost:3000",
      },
    },
  },
});
