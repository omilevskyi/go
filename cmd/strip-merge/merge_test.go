package main

import (
	"bytes"
	"testing"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
)

func parseSingleAttribute(t *testing.T, src string) (*hclsyntax.Attribute, []byte) {
	t.Helper()

	b := []byte(src)
	file, diags := hclsyntax.ParseConfig(b, nullHcl, hcl.Pos{Line: 1, Column: 1})
	if diags.HasErrors() {
		t.Fatalf("parse error: %s", diags.Error())
	}

	body := file.Body.(*hclsyntax.Body)

	if len(body.Attributes) != 1 {
		t.Fatalf("expected exactly one attribute")
	}

	for _, attr := range body.Attributes {
		return attr, b
	}

	panic("unreachable")
}

func TestLargestMergeLiteral_Single(t *testing.T) {
	attr, src := parseSingleAttribute(t, `
inputs = merge(a, {
  x = 1
})
`)

	obj, ok := largestMergeLiteral(src, attr)
	if !ok {
		t.Fatalf("expected object")
	}

	if !bytes.Contains(obj, []byte("x = 1")) {
		t.Fatalf("unexpected object content: %s", obj)
	}
}

func TestLargestMergeLiteral_Multiple(t *testing.T) {
	attr, src := parseSingleAttribute(t, `
inputs = merge(
  { a = 1 },
  { b = 2, c = 3 },
  { d = 4 }
)
`)

	obj, ok := largestMergeLiteral(src, attr)
	if !ok {
		t.Fatalf("expected object")
	}

	if !bytes.Contains(obj, []byte("b = 2")) ||
		!bytes.Contains(obj, []byte("c = 3")) {
		t.Fatalf("did not select largest object: %s", obj)
	}
}

func TestLargestMergeLiteral_None(t *testing.T) {
	attr, src := parseSingleAttribute(t, `
inputs = merge(a, b, c)
`)

	_, ok := largestMergeLiteral(src, attr)
	if ok {
		t.Fatalf("expected no object")
	}
}

func TestLargestMergeLiteral_NotMerge(t *testing.T) {
	attr, src := parseSingleAttribute(t, `
inputs = { a = 1 }
`)

	_, ok := largestMergeLiteral(src, attr)
	if ok {
		t.Fatalf("expected no object")
	}
}
