package main

import (
	"go/ast"
	"go/parser"
	"go/token"
	"testing"
)

// TestGetCart_NoGetProductLoop は handler.go の GetCart メソッド内に
// GetProduct の逐次呼び出しが存在しないことを go/ast で静的に検証する。
// GetCart は商品情報を返さない設計のため、GetProduct の呼び出しが
// GetCart 内に存在してはならない。
func TestGetCart_NoGetProductLoop(t *testing.T) {
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, "handler.go", nil, 0)
	if err != nil {
		t.Fatalf("failed to parse handler.go: %v", err)
	}

	// GetCart メソッドの関数本体を探す
	var getCartBody *ast.BlockStmt
	ast.Inspect(f, func(n ast.Node) bool {
		fn, ok := n.(*ast.FuncDecl)
		if !ok {
			return true
		}
		if fn.Name.Name == "GetCart" && fn.Body != nil {
			getCartBody = fn.Body
			return false
		}
		return true
	})

	if getCartBody == nil {
		t.Fatal("GetCart function not found in handler.go")
	}

	// GetCart 本体内に "GetProduct" のセレクタ呼び出しがないことを確認
	ast.Inspect(getCartBody, func(n ast.Node) bool {
		call, ok := n.(*ast.CallExpr)
		if !ok {
			return true
		}
		sel, ok := call.Fun.(*ast.SelectorExpr)
		if !ok {
			return true
		}
		if sel.Sel.Name == "GetProduct" {
			t.Errorf("GetCart contains a call to GetProduct at %s; GetCart should not call product service",
				fset.Position(call.Pos()))
		}
		return true
	})
}
