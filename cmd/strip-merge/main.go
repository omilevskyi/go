package main

import (
	"fmt"
	"io"
	"os"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/hashicorp/hcl/v2/hclwrite"

	ut "github.com/omilevskyi/go/pkg/utils"
)

const (
	appName = "strip-merge"
	nullHcl = "null.hcl"
)

func largestMergeLiteral(src []byte, attr *hclsyntax.Attribute) ([]byte, bool) {
	call, ok := attr.Expr.(*hclsyntax.FunctionCallExpr)
	if ok && call.Name == "merge" {
		var bestObj *hclsyntax.ObjectConsExpr
		var bestSize int

		for _, arg := range call.Args {
			if obj, ok := arg.(*hclsyntax.ObjectConsExpr); ok {
				objRange := obj.SrcRange // original text of argument
				objSize := objRange.End.Byte - objRange.Start.Byte

				if bestObj == nil || objSize > bestSize {
					bestObj, bestSize = obj, objSize
				}
			}
		}

		if bestObj != nil {
			r := bestObj.SrcRange
			return src[r.Start.Byte:r.End.Byte], true
		}
	}
	return nil, false
}

func checkDiags(diags hcl.Diagnostics, w io.Writer, msg string, rc int) {
	if diags.HasErrors() {
		length := len(diags)
		_, _ = fmt.Fprintf(w, "%s: %d diagnostic(s):\n\n", msg, length)
		for i := 0; i < length; i++ {
			_, _ = fmt.Fprintf(w, "%s: %s\n", diags[i].Summary, diags[i].Detail)
		}
		if rc > 0 {
			os.Exit(rc)
		}
	}
}

func main() {
	src, err := io.ReadAll(os.Stdin)
	ut.IsErr(err, 201, appName)

	// syntax AST - read only
	syntaxFile, diags := hclsyntax.ParseConfig(src, nullHcl, hcl.Pos{Line: 1, Column: 1})
	checkDiags(diags, os.Stderr, "Failed to parse for reading", 202)

	// write AST - write only
	writeFile, diags := hclwrite.ParseConfig(src, nullHcl, hcl.Pos{Line: 1, Column: 1})
	checkDiags(diags, os.Stderr, "Failed to parse for writing", 203)

	writeBody := writeFile.Body()
	syntaxBody := syntaxFile.Body.(*hclsyntax.Body)

	for name, attr := range syntaxBody.Attributes {
		if objBytes, ok := largestMergeLiteral(src, attr); ok {
			// replace entire attribute with raw expression
			writeBody.SetAttributeRaw(name, hclwrite.Tokens{
				&hclwrite.Token{
					Type:  hclsyntax.TokenIdent,
					Bytes: objBytes,
				},
			})
		}
	}

	formatted := hclwrite.Format(writeFile.Bytes())
	_, _ = os.Stdout.Write(formatted)
}
