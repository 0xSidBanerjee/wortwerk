package parse

import "testing"

func TestParse_fixture(t *testing.T) {
	src := `---
title: Fixture
slug: fixture
level: A1
lang: de
---

Pinocchio steht früh auf. Er geht zum Markt.`
	doc, err := Parse(src)
	if err != nil {
		t.Fatal(err)
	}
	if doc.Meta.Slug != "fixture" {
		t.Fatalf("slug %q", doc.Meta.Slug)
	}
	if len(doc.Paragraphs) != 1 {
		t.Fatalf("paragraphs %d", len(doc.Paragraphs))
	}
}
