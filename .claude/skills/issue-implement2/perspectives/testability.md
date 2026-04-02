# Review: テスト駆動性

## Overview
契約・テスト戦略・Red から、テストケース導出が可能か、粒度が適切かをレビューする。

## Step 1: Analysis

- `contract.md`
- `test-strategy.md`
- `red-summary.md`
- 新規・変更されたテスト

## Step 2: Findings

### 指摘とする基準

- テスト戦略と Red が一致していない
- 実装依存の期待値になっている
- エッジケースや異常系が曖昧
- 主要な受け入れ条件が検証されていない

### 指摘としない基準

- テストフレームワークの好み
- 全ケース網羅の要求

### perspective ラベル
テスト駆動性
