# bmkr C4 Overview

## Level 1: System Context

```mermaid
flowchart LR
    customer["EC Customer\nBrowses products and uses shopping features."]
    operator["Store Operator\nMaintains product information and checks system behavior."]
    bmkr["bmkr\nLearning-oriented EC platform that provides storefront, cart, and product management capabilities."]
    hydra[("Ory Hydra\nOAuth 2.0 authorization server used for service authentication.")]
    db[("PostgreSQL\nStores product, customer, and cart data.")]

    customer -->|HTTPS\nUses storefront and cart features| bmkr
    operator -->|HTTPS\nManages products through the admin UI| bmkr
    bmkr -->|HTTPS\nObtains and validates OAuth 2.0 access tokens| hydra
    bmkr -->|TCP\nReads and writes domain data| db
```

## Level 2: Container

```mermaid
flowchart TB
    customer["EC Customer"]
    operator["Store Operator"]

    subgraph bmkr ["bmkr"]
        spa["React SPA\n[TypeScript, Vite]\nStorefront UI for browsing products."]
        bff["BFF\n[TypeScript, Fastify]\nAPI gateway that aggregates backend calls.\nPort 3000"]
        ecsite["EC Site API\n[Go, Echo]\nCart management service.\nPort 8080"]
        productmgmt["Product Mgmt API\n[Go, Echo]\nProduct CRUD and admin UI (templ + HTMX).\nPort 8081"]
        customermgmt["Customer Mgmt API\n[Go, Echo]\nCustomer registration and lookup.\nPort 8082"]
    end

    hydra[("Ory Hydra\n[OAuth 2.0 Server]\nIssues and validates\nservice-to-service tokens.")]
    ecdb[("PostgreSQL\necsite database\nCarts, cart items.")]
    productdb[("PostgreSQL\nproduct database\nProducts.")]
    customerdb[("PostgreSQL\ncustomer database\nCustomers.")]

    customer -->|HTTPS| spa
    spa -->|"Connect protocol (JSON)\nvia BFF proxy"| bff
    operator -->|"HTTPS\ntempl + HTMX"| productmgmt

    bff -->|"Connect RPC (HTTP/2)"| productmgmt

    ecsite -->|"Connect RPC (HTTP/2)"| productmgmt

    ecsite --> ecdb
    productmgmt --> productdb
    customermgmt --> customerdb

    bff -.->|"OAuth 2.0\nClient Credentials"| hydra
    ecsite -.->|"OAuth 2.0\nClient Credentials"| hydra
```

> 実線 = 実装済み、点線 = ADR 決定済み・未実装

## File Mapping

| Container | Directory | Entry Point |
|-----------|-----------|-------------|
| React SPA | `services/ec-site/frontend/` | `src/App.tsx` |
| BFF | `services/bff/` | `src/index.ts` |
| EC Site API | `services/ec-site/` | `main.go` |
| Product Mgmt API | `services/product-mgmt/` | `main.go` |
| Customer Mgmt API | `services/customer-mgmt/` | `main.go` |

| Proto Definition | Path |
|------------------|------|
| ProductService | `proto/product/v1/product.proto` |
| CartService | `proto/ec/v1/cart.proto` |
| OrderService | `proto/ec/v1/order.proto` |
| CustomerService | `proto/customer/v1/customer.proto` |
