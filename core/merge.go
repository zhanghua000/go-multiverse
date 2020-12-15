package core

import (
	"context"
	"errors"
	"strings"

	"github.com/ipfs/go-cid"
	ipld "github.com/ipfs/go-ipld-format"
	"github.com/ipfs/go-merkledag"
	"github.com/ipfs/go-merkledag/dagutils"
	"github.com/multiverse-vcs/go-multiverse/diff"
	"github.com/multiverse-vcs/go-multiverse/object"
	"github.com/spf13/afero"
)

// Merge combines the work trees of a and b into the base o.
func Merge(ctx context.Context, fs afero.Fs, dag ipld.DAGService, o, a, b cid.Cid) (ipld.Node, error) {
	changesA, err := Diff(ctx, dag, o, a)
	if err != nil {
		return nil, err
	}

	changesB, err := Diff(ctx, dag, o, b)
	if err != nil {
		return nil, err
	}

	changes, conflicts := dagutils.MergeDiffs(changesA, changesB)
	for _, c := range conflicts {
		change, err := mergeConflict(ctx, fs, dag, c.A, c.B)
		if err != nil {
			return nil, err
		}

		changes = append(changes, change)
	}

	node, err := dag.Get(ctx, o)
	if err != nil {
		return nil, err
	}

	commit, err := object.CommitFromCBOR(node.RawData())
	if err != nil {
		return nil, err
	}

	tree, err := dag.Get(ctx, commit.Tree)
	if err != nil {
		return nil, err
	}

	proto, ok := tree.(*merkledag.ProtoNode)
	if !ok {
		return nil, errors.New("invalid commit tree")
	}

	return dagutils.ApplyChange(ctx, dag, proto, changes)
}

// mergeConflict combines the contents of two conflicting dag changes.
func mergeConflict(ctx context.Context, fs afero.Fs, dag ipld.DAGService, a, b *dagutils.Change) (*dagutils.Change, error) {
	if a.Type == dagutils.Remove {
		return b, nil
	}

	if b.Type == dagutils.Remove {
		return a, nil
	}

	textO, err := Cat(ctx, dag, a.Before)
	if err != nil {
		return nil, err
	}

	textA, err := Cat(ctx, dag, a.After)
	if err != nil {
		return nil, err
	}

	textB, err := Cat(ctx, dag, b.After)
	if err != nil {
		return nil, err
	}

	merged := diff.Merge(textO, textA, textB)
	reader := strings.NewReader(merged)

	merge, err := add(ctx, dag, reader)
	if err != nil {
		return nil, err
	}

	change := dagutils.Change{
		Type:   dagutils.Mod,
		Path:   a.Path,
		Before: a.Before,
		After:  merge.Cid(),
	}

	if a.Type == dagutils.Add && b.Type == dagutils.Add {
		change.Type = dagutils.Add
	}

	return &change, nil
}
