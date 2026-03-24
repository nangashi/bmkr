package main

import (
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// ---------------------------------------------------------------------------
// 受け入れ条件1: 全サービスから "log" パッケージの import がなくなっている
// ---------------------------------------------------------------------------
// go/parser を使って各サービスの Go ファイルをパースし、
// "log" パッケージ（"log/slog" ではない）の import がないことを確認する。
// テストファイル自身は検査対象外。

// TestNoStdLogImport_EcSite は ec-site サービス配下の全 .go ファイルを検査する。
func TestNoStdLogImport_EcSite(t *testing.T) {
	assertNoStdLogImport(t, ".")
}

// TestNoStdLogImport_ProductMgmt は product-mgmt サービス配下の全 .go ファイルを検査する。
func TestNoStdLogImport_ProductMgmt(t *testing.T) {
	assertNoStdLogImport(t, "../product-mgmt")
}

// TestNoStdLogImport_CustomerMgmt は customer-mgmt サービス配下の全 .go ファイルを検査する。
func TestNoStdLogImport_CustomerMgmt(t *testing.T) {
	assertNoStdLogImport(t, "../customer-mgmt")
}

// assertNoStdLogImport は指定ディレクトリ配下の全 .go ファイルをパースし、
// "log" パッケージの import がないことを確認する。
// "log/slog" は許可する。テストファイル (_test.go) は対象外とする。
// db/generated 配下の自動生成コードも対象外とする。
func assertNoStdLogImport(t *testing.T, dir string) {
	t.Helper()

	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// ディレクトリはスキップ（ただし再帰は継続）
		if info.IsDir() {
			// db/generated は自動生成コードなのでスキップ
			if info.Name() == "generated" && strings.Contains(path, "db") {
				return filepath.SkipDir
			}
			return nil
		}

		// .go ファイル以外はスキップ
		if !strings.HasSuffix(path, ".go") {
			return nil
		}

		// テストファイルはスキップ
		if strings.HasSuffix(path, "_test.go") {
			return nil
		}

		// Go ソースをパース
		fset := token.NewFileSet()
		f, parseErr := parser.ParseFile(fset, path, nil, parser.ImportsOnly)
		if parseErr != nil {
			t.Errorf("failed to parse %s: %v", path, parseErr)
			return nil
		}

		// import を検査
		for _, imp := range f.Imports {
			// imp.Path.Value は `"log"` のようにダブルクォート付き
			importPath := strings.Trim(imp.Path.Value, `"`)
			if importPath == "log" {
				t.Errorf("%s: found import of \"log\" package (should use \"log/slog\" instead)", path)
			}
		}

		return nil
	})
	if err != nil {
		t.Fatalf("failed to walk directory %s: %v", dir, err)
	}
}
