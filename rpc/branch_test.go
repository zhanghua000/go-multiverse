package rpc

import (
	"context"
	"io/ioutil"
	"testing"

	"github.com/ipfs/go-datastore"
	"github.com/multiverse-vcs/go-multiverse/data"
	"github.com/multiverse-vcs/go-multiverse/peer"
)

func TestListBranches(t *testing.T) {
	ctx := context.Background()

	dstore := datastore.NewMapDatastore()
	store := data.NewStore(dstore)

	mock, err := peer.Mock(ctx, dstore)
	if err != nil {
		t.Fatal("failed to create peer")
	}

	json, err := ioutil.ReadFile("testdata/repository.json")
	if err != nil {
		t.Fatal("failed to read json")
	}

	repo, err := data.RepositoryFromJSON(json)
	if err != nil {
		t.Fatal("failed to parse repository json")
	}

	id, err := data.AddRepository(ctx, mock, repo)
	if err != nil {
		t.Fatal("failed to create repository")
	}

	if err := store.PutCid(repo.Name, id); err != nil {
		t.Fatal("failed to put cid in store")
	}

	client, err := connect(mock, store)
	if err != nil {
		t.Fatal("failed to connect to rpc server")
	}

	args := BranchArgs{
		Name: repo.Name,
	}

	var reply BranchReply
	if err := client.Call("Service.ListBranches", &args, &reply); err != nil {
		t.Fatal("failed to call rpc")
	}

	if len(reply.Branches) != 1 {
		t.Error("unexpected branches")
	}

	if reply.Branches["default"] != repo.Branches["default"] {
		t.Error("unexpected branches")
	}
}
