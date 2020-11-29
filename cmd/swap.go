package cmd

import (
	"fmt"
	"os"
	"os/signal"

	"github.com/multiverse-vcs/go-multiverse/p2p"
	"github.com/urfave/cli/v2"
)

// NewSwapCommand returns a new serve command.
func NewSwapCommand() *cli.Command {
	return &cli.Command{
		Name:  "swap",
		Usage: "Exchange data with peers",
		Action: func(c *cli.Context) error {
			store, err := Store()
			if err != nil {
				return cli.Exit(err.Error(), 1)
			}

			if err := store.Online(c.Context); err != nil {
				return cli.Exit(err.Error(), 1)
			}

			fmt.Printf("bootstrapping network...\n")
			p2p.Bootstrap(c.Context, store.Host)

			if err := p2p.Discovery(c.Context, store.Host); err != nil {
				return cli.Exit(err.Error(), 1)
			}

			if err := store.Router.Bootstrap(c.Context); err != nil {
				return cli.Exit(err.Error(), 1)
			}

			fmt.Println("listening on:")
			id := store.Host.ID().Pretty()

			for _, a := range store.Host.Addrs() {
				fmt.Printf("* %s/p2p/%s\n", a, id)
			}

			interrupt := make(chan os.Signal, 1)
			signal.Notify(interrupt, os.Interrupt)

			<-interrupt
			return nil
		},
	}
}
