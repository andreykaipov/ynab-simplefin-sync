package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"strings"

	"github.com/alecthomas/kong"
	"github.com/andreykaipov/ynab-simplefin-sync/simplefin"
	"github.com/andreykaipov/ynab-simplefin-sync/ynab"
	"github.com/olekukonko/tablewriter"
	ynabgo "go.bmvs.io/ynab"
	"gopkg.in/yaml.v3"
)

var start string
var end string
var out string

//func init() {
//	flag.StringVar(&start, "start", "", "Start date of transactions range")
//	flag.StringVar(&end, "end", "", "End date of transactions range")
//	flag.StringVar(&out, "out", ".", "Output directory")
//	flag.Parse()
//
//	if accessURL = os.Getenv("ACCESS_URL"); accessURL == "" {
//		log.Fatal("Provide the SimpleFIN ACCESS_URL")
//	}
//}

type CLI struct {
	SimpleFINAccessURL string `required:"" env:"SIMPLEFIN_ACCESS_URL" name:"simplefin_access_url" help:"SimpleFIN access URL (claimed from token already)"`
	YnabAccessToken    string `required:"" help:"Your YNAB access token to its API"`
	YnabBudgetID       string `required:"" help:"Your YNAB budget UUID"`

	Sync SyncCmd `cmd:"" help:"Sync SimpleFIN transactions with your YNAB budget"`

	YnabCmd struct {
		Accounts ynab.Accounts `cmd:"" help:""`
	} `cmd:"" name:"ynab" help:""`

	SimpleFINCmd struct {
		Accounts     simplefin.Accounts     `cmd:"" help:""`
		Transactions simplefin.Transactions `cmd:"" help:""`
	} `cmd:"" name:"simplefin" help:""`
}

func (o *CLI) AfterApply(ctx *kong.Context) error {
	ctx.Bind(simplefin.Client{AccessURL: o.SimpleFINAccessURL})
	ctx.Bind(o.YnabBudgetID)
	ctx.BindTo(ynabgo.NewClient(o.YnabAccessToken), (*ynabgo.ClientServicer)(nil))

	table := tablewriter.NewWriter(os.Stdout)
	table.SetAutoWrapText(false)
	table.SetAutoFormatHeaders(true)
	table.SetHeaderAlignment(tablewriter.ALIGN_LEFT)
	table.SetAlignment(tablewriter.ALIGN_LEFT)
	table.SetCenterSeparator("")
	table.SetColumnSeparator("")
	table.SetRowSeparator("")
	table.SetHeaderLine(false)
	table.SetBorder(false)
	table.SetTablePadding("\t") // pad with tabs
	table.SetNoWhiteSpace(true)
	ctx.Bind(table)

	return nil
}

func main() {
	log.SetFlags(log.Lshortfile)
	home := os.Getenv("HOME")
	cli := &CLI{}

	ctx := kong.Parse(
		cli,
		kong.Name("ynab-simplefin-sync"),
		kong.Description("Sync your YNAB accounts with info from SimpleFIN via SimpleFIN Bridge"),
		kong.Configuration(
			yamlEnvResolver,
			"config.yml",
			fmt.Sprintf("%s/.ynab-simplefin-sync.yml", home),
			fmt.Sprintf("%s/config/ynab-simplefin-sync/.config.yml", home),
		),
		kong.ConfigureHelp(kong.HelpOptions{Compact: true}),
	)

	ctx.FatalIfErrorf(ctx.Run())
}

// resolves configs to init our commands' flags
func yamlEnvResolver(r io.Reader) (kong.Resolver, error) {
	values := map[string]interface{}{}

	if err := yaml.NewDecoder(r).Decode(&values); err != nil {
		return nil, err
	}

	var f kong.ResolverFunc = func(context *kong.Context, _ *kong.Path, flag *kong.Flag) (interface{}, error) {
		name := strings.ReplaceAll(flag.Name, "-", "_")
		val, ok := values[name]
		if !ok {
			return nil, nil
		}

		switch v := val.(type) {
		case string:
			k := "envexpand:"
			if strings.HasPrefix(v, k) {
				val = os.ExpandEnv(v[len(k):])
			}
		case map[string]interface{}:
		default:
			log.Fatalf("%s is of type %T instead of a string: %s", flag.Name, v, v)
		}

		return val, nil
	}

	return f, nil
}

func marshal(v interface{}) string {
	b, err := json.Marshal(v)
	if err != nil {
		panic(fmt.Errorf("marshalling %#v: %w", v, err))
	}

	return string(b)
}

/*
func f() {
	//_ = ynab.NewClient(o.YnabAccessToken)

	//budget, err := client.Transaction().CreateTransactions(o.BudgetID, nil)
	//if err != nil {
	//	return err
	//}

	//	for _, t := range budget.Budget.Transactions {
	//		fmt.Printf("%#v\n", t.Date)
	//	}

	for _, account := range data.Accounts {
		fname := slug(account.Org.Name) + "_" + slug(account.Name) + ".csv"
		f, err := os.Create(out + "/" + fname)
		if err != nil {
			log.Fatal(err)
		}
		defer f.Close()

		fmt.Printf("For %s, balances are... [Current: %v] [Available: %v]\n", fname, account.Balance, account.AvailableBalance)

		w := csv.NewWriter(f)
		defer w.Flush()

		w.Write([]string{"Date", "Payee", "Memo", "Amount"})

		for _, tx := range account.Transactions {
			date := time.Unix(int64(tx.Posted), 0).Format("01-02-2006")

			memo := tx.Memo
			if memo == "" {
				memo = tx.Description
			}

			w.Write([]string{date, tx.Payee, memo, tx.Amount})
		}
	}
}
*/
