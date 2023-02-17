package main

import (
	"fmt"

	"github.com/mkideal/cli"
	"github.com/pkg/errors"
)

type CCT struct {
	cli.Helper
	Name    string `cli:"name,na" usage:"company name"`
	City    string `cli:"city,ci" usage:"city"`
	Country string `cli:"country,co" usage:"country"`
	Domain  string `cli:"domain,do" usage:"domain name"`
	Ref     string `cli:"reference,re" usage:"reference"`
}

type RegisterT struct {
	cli.Helper
	Name string `cli:"name,na" usage:"client name"`
}

type AddTrustedpartyT struct {
	cli.Helper
	CompanyID string `cli:"company-id,cid" usage:"company id"`
	AgentId   string `cli:"agent-id,aid" usage:"agent id"`
}

type AddWhitelistT struct {
	cli.Helper
	CompanyID string `cli:"company-id,cid" usage:"company id"`
	FundID    string `cli:"fund-id,fid" usage:"fund id"`
	Address   string `cli:"address,ads" usage:"address"`
}

type ReadFeedT struct {
	cli.Helper
	FeedURL string `cli:"feed-url,furl" usage:"feed url"`
}

type ApproveActionT struct {
	cli.Helper
	ActionID string `cli:"action-id,aid" usage:"action id"`
}

type WithdrawT struct {
	cli.Helper
	CompanyID string `cli:"company-id,cid" usage:"company id"`
	WalletID  string `cli:"wallet-id,wid" usage:"wallet id"`
	Address   string `cli:"address,addr" usage:"address"`
	Amount    int64  `cli:"amount,am" usage:"amount"`
}

type CreateFundT struct {
	cli.Helper
	CompanyID       string `cli:"company-id,cid" usage:"company id"`
	MemberID        string `cli:"member-id,mid" usage:"member id"`
	FundName        string `cli:"fund-name,fn" usage:"fund name"`
	FundDescription string `cli:"fund-description,fd" usage:"fund description"`
}

func NewDCli() *dCli {

	demo, _ := NewDemo("https://play-api.qredo.network/api/v1/p", APIKey, PrivateKey)

	dCli := &dCli{demo: demo}

	var register = &cli.Command{
		Name:    "register",
		Aliases: []string{"reg"},
		Desc:    "register a new agent",
		Argv:    func() interface{} { return new(RegisterT) },
		Fn:      dCli.register,
	}

	var cc = &cli.Command{
		Name:    "create-company",
		Aliases: []string{"cc"},
		Desc:    "create a new company",
		Argv:    func() interface{} { return new(CCT) },
		Fn:      dCli.createCompany,
	}

	var addWhitelist = &cli.Command{
		Name:    "add-whitelist",
		Aliases: []string{"awl"},
		Desc:    "whitelists the given address",
		Argv:    func() interface{} { return new(AddWhitelistT) },
		Fn:      dCli.addWhitelist,
	}

	var trustedparty = &cli.Command{
		Name:    "add-trustedparty",
		Aliases: []string{"add-tp"},
		Desc:    "add a new trustedparty",
		Argv:    func() interface{} { return new(AddTrustedpartyT) },
		Fn:      dCli.trustedparty,
	}

	var readAction = &cli.Command{
		Aliases: []string{"furl"},
		Name:    "read-action",
		Desc:    "connect to qredo web socket stream by given feed url",
		Argv:    func() interface{} { return new(ReadFeedT) },
		Fn:      dCli.readAction,
	}

	var approveAction = &cli.Command{
		Aliases: []string{"aa"},
		Name:    "approve-action",
		Desc:    "approve action by given action id",
		Argv:    func() interface{} { return new(ApproveActionT) },
		Fn:      dCli.approveAction,
	}

	var withdraw = &cli.Command{
		Aliases: []string{"wd"},
		Name:    "withdraw",
		Desc:    "create withdraw request by given arguments",
		Argv:    func() interface{} { return new(WithdrawT) },
		Fn:      dCli.withdraw,
	}

	var createFund = &cli.Command{
		Aliases: []string{"cf"},
		Name:    "create-fund",
		Desc:    "create fund and wallets by given arguments",
		Argv:    func() interface{} { return new(CreateFundT) },
		Fn:      dCli.createFund,
	}

	dCli.registerCmd = register
	dCli.ccCmd = cc
	dCli.addWhitelistCmd = addWhitelist
	dCli.trustedpartyCmd = trustedparty
	dCli.readActionCmd = readAction
	dCli.approveActionCmd = approveAction
	dCli.withdrawCmd = withdraw
	dCli.createFundCmd = createFund

	return dCli
}

type dCli struct {
	registerCmd      *cli.Command
	ccCmd            *cli.Command
	addWhitelistCmd  *cli.Command
	trustedpartyCmd  *cli.Command
	readActionCmd    *cli.Command
	approveActionCmd *cli.Command
	withdrawCmd      *cli.Command
	createFundCmd    *cli.Command
	demo             *SaCli
}

func (d *dCli) register(ctx *cli.Context) error {
	argv, ok := ctx.Argv().(*RegisterT)
	if !ok {
		return errors.New("cannot cast RegisterT")
	}

	register, err := d.demo.Register(argv.Name)
	if err != nil {
		return err
	}
	fmt.Printf("[ok] account-code: %s,  feed: %s\n", register.AccountCode, register.FeedURL)
	return nil
}

func (d *dCli) createCompany(ctx *cli.Context) error {
	argv, ok := ctx.Argv().(*CCT)
	if !ok {
		return errors.New("cannot cast CCT")
	}

	company, err := d.demo.CreateCompany(argv.Name, argv.City, argv.Country, argv.Domain, argv.Ref)
	if err != nil {
		return err
	}
	fmt.Printf("[ok] company-id: %s,  ref: %s\n", company.CompanyID, company.Ref)
	return nil
}

func (d *dCli) addWhitelist(ctx *cli.Context) error {
	argv, ok := ctx.Argv().(*AddWhitelistT)
	if !ok {
		return errors.New("cannot cast AddWhitelistT")
	}
	if err := d.demo.AddWhitelist(argv.CompanyID, argv.FundID, argv.Address); err != nil {
		return err
	}
	fmt.Println("[ok]")
	return nil
}

func (d *dCli) trustedparty(ctx *cli.Context) error {
	argv, ok := ctx.Argv().(*AddTrustedpartyT)
	if !ok {
		return errors.New("cannot cast AddTrustedpartyT")
	}

	if err := d.demo.AddTrustedparty(argv.CompanyID, argv.AgentId); err != nil {
		return err
	}
	fmt.Println("[ok]")
	return nil
}

func (d *dCli) readAction(ctx *cli.Context) error {
	argv, ok := ctx.Argv().(*ReadFeedT)
	if !ok {
		return errors.New("cannot cast ReadFeedT")
	}

	if err := d.demo.ReadAction(argv.FeedURL); err != nil {
		return err
	}
	fmt.Println("[ok]")
	return nil
}

func (d *dCli) approveAction(ctx *cli.Context) error {
	argv, ok := ctx.Argv().(*ApproveActionT)
	if !ok {
		return errors.New("cannot cast ApproveActionT")
	}

	if err := d.demo.Approve(argv.ActionID); err != nil {
		return err
	}
	fmt.Println("[ok]")
	return nil
}

func (d *dCli) withdraw(ctx *cli.Context) error {
	argv, ok := ctx.Argv().(*WithdrawT)
	if !ok {
		return errors.New("cannot cast WithdrawT")
	}

	tResp, err := d.demo.Withdraw(argv.CompanyID, argv.WalletID, argv.Address, argv.Amount)
	if err != nil {
		return err
	}
	fmt.Printf("[ok] status %v tx-id %s\n", tResp.Status, tResp.TxID)

	return nil
}

func (d *dCli) createFund(ctx *cli.Context) error {
	argv, ok := ctx.Argv().(*CreateFundT)
	if !ok {
		return errors.New("cannot cast CreateFundT")
	}

	fResp, err := d.demo.CreateFund(argv.CompanyID, argv.FundName, argv.FundDescription, argv.MemberID)
	if err != nil {
		return err
	}
	fmt.Printf("[ok] fund-id %s, custodygroup-withdraw %s, custodygroup-tx %s\n",
		fResp.FundID, fResp.CustodygroupWithdraw, fResp.CustodygroupTx)

	return nil
}
