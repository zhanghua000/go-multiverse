package core

import (
	"context"
	"testing"

	"github.com/ipfs/go-merkledag/dagutils"
	"github.com/spf13/afero"
)

func TestCat(t *testing.T) {
	fs := afero.NewMemMapFs()
	dag := dagutils.NewMemoryDagService()

	if err := afero.WriteFile(fs, "test.txt", []byte("foo bar"), 0644); err != nil {
		t.Fatalf("failed to write file")
	}

	node, err := Add(context.TODO(), fs, dag, "test.txt", nil)
	if err != nil {
		t.Fatalf("failed to add file")
	}

	text, err := Cat(context.TODO(), dag, node.Cid())
	if err != nil {
		t.Fatal("failed to cat file")
	}

	if text != "foo bar" {
		t.Error("unexpected file contents")
	}
}
