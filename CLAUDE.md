# CLAUDE.md

ECサイト・商品管理・顧客管理のマイクロサービスの学習用実装

## Documents

- システム概要は docs/c4_overview.md を参照
- 技術的判断は docs/adr/ を参照
- 実装パターン・アンチパターンは docs/guides/ を参照
- ファイル編集時の制約は .claude/rules/ を参照
- プロジェクトで実行するコマンドは `just --list` を参照

## Rules

- 仕様調査には context7 スキルを利用する
- LLMの一時的な書き込みは.output/を利用する
