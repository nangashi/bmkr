# 開発環境のセットアップ
setup:
    mise install
    pnpm install

# Protobuf のリント
lint:
    buf lint

# コード生成
generate:
    buf generate
