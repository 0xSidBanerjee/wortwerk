package main

import "testing"

func TestUsageConstant(t *testing.T) {
	if usage == "" {
		t.Fatal("usage text must be non-empty")
	}
}
