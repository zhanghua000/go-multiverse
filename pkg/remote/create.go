package remote

import (
	"context"
	"errors"

	"github.com/multiverse-vcs/go-multiverse/pkg/object"
	"github.com/multiverse-vcs/go-multiverse/pkg/p2p"
)

// CreateArgs contains the args.
type CreateArgs struct {
	// Name is the repository name.
	Name string
}

// CreateReply contains the reply
type CreateReply struct {
	// Remote is the repository path.
	Remote Path
}

// Create creates a new repository.
func (s *Server) Create(args *CreateArgs, reply *CreateReply) error {
	ctx := context.Background()

	key, err := p2p.DecodeKey(s.Config.PrivateKey)
	if err != nil {
		return err
	}

	peerID := s.Peer.Host.ID()
	author := s.Config.Author

	if _, ok := author.Repositories[args.Name]; ok {
		return errors.New("repository already exists")
	}

	repo := object.NewRepository()
	repoID, err := object.AddRepository(ctx, s.Peer.DAG, repo)
	if err != nil {
		return err
	}

	author.Repositories[args.Name] = repoID
	if err := s.Config.Write(); err != nil {
		return err
	}

	authorID, err := object.AddAuthor(ctx, s.Peer.DAG, author)
	if err != nil {
		return err
	}

	reply.Remote = NewPath(peerID, args.Name)
	return s.Namesys.Publish(ctx, key, authorID)
}
