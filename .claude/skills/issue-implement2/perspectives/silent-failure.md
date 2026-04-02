# Review: サイレントフェイル

## Overview
エラー握りつぶし、log して return nil、recover で panic 吸収など、障害を隠蔽するパターンを検出する。

## Step 1: Analysis

- error を返す関数呼び出し
- recover() の使用
- ログ出力後の制御フロー
- エラー無視の箇所

## Step 2: Findings

### 指摘とする基準

- エラーを無視して後続処理が続く
- ログだけ出して成功扱いする
- panic を吸収して正常系として継続する

### 指摘としない基準

- 慣習的な Close エラー無視
- テストコード内の軽微な無視
- 契約で明示されたフォールバック

### perspective ラベル
サイレントフェイル
