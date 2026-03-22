package main

// ProductItem は商品一覧テーブルの1行分のプレゼンテーション用の型。
// DB 生成型 (db.Product) をテンプレートに直接渡さず、表示用に変換して使う。
//
// 動作:
//   - ID: 商品ID（int64）をそのまま保持する
//   - Name: 商品名をそのまま保持する
//   - Price: 円単位の整数値（int64）。テンプレート側で表示フォーマットする
//   - StockQuantity: 在庫数（int32）をそのまま保持する
//   - CreatedAt: "2006-01-02 15:04" 形式にフォーマット済みの文字列
type ProductItem struct {
	ID            int64
	Name          string
	Price         int64
	StockQuantity int32
	CreatedAt     string // フォーマット済み日時文字列
}

// ProductListData は商品一覧ページ全体のプレゼンテーション用の型。
// テンプレートに渡すデータをまとめる。
//
// 動作:
//   - Products: 現在のページに表示する商品一覧（0件の場合は空スライス）
//   - CurrentPage: 現在のページ番号（1始まり）
//   - TotalPages: 総ページ数（商品0件のとき1）
//   - TotalCount: 商品の総件数
//   - PerPage: 1ページあたりの表示件数
type ProductListData struct {
	Products    []ProductItem
	CurrentPage int
	TotalPages  int
	TotalCount  int64
	PerPage     int
}

// HasPrevPage は前のページが存在するかを返す。
// CurrentPage が 2 以上のとき true を返す。
func (d ProductListData) HasPrevPage() bool {
	return d.CurrentPage > 1
}

// HasNextPage は次のページが存在するかを返す。
// CurrentPage が TotalPages 未満のとき true を返す。
func (d ProductListData) HasNextPage() bool {
	return d.CurrentPage < d.TotalPages
}

// PrevPage は前のページ番号を返す。
// 最初のページのとき 1 を返す（下限クランプ）。
func (d ProductListData) PrevPage() int {
	if d.CurrentPage <= 1 {
		return 1
	}
	return d.CurrentPage - 1
}

// NextPage は次のページ番号を返す。
// 最後のページのとき TotalPages を返す（上限クランプ）。
func (d ProductListData) NextPage() int {
	if d.CurrentPage >= d.TotalPages {
		return d.TotalPages
	}
	return d.CurrentPage + 1
}

// defaultPerPage は商品一覧のデフォルト表示件数。
const defaultPerPage = 20
