package main

import (
	"fmt"
	"github.com/mkideal/cli"
	"os"
)

type rootT struct {
	cli.Helper
}

func main() {

	dCli := NewDCli()

	var help = cli.HelpCommand("display help information")

	var root = &cli.Command{
		Desc: "signing agent demo",
		Argv: func() interface{} { return new(rootT) },
		Fn: func(ctx *cli.Context) error {
			ctx.String("use -h to see the available options.\n")
			return nil
		},
	}

	if err := cli.Root(root,
		cli.Tree(help),
		cli.Tree(dCli.ccCmd),
		cli.Tree(dCli.registerCmd),
		cli.Tree(dCli.trustedpartyCmd),
		cli.Tree(dCli.addWhitelistCmd),
		cli.Tree(dCli.readActionCmd),
		cli.Tree(dCli.approveActionCmd),
		cli.Tree(dCli.withdrawCmd),
		cli.Tree(dCli.createFundCmd),
	).Run(os.Args[1:]); err != nil {
		_, _ = fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
