package main

import (
	"fmt"
	"io"
	"log"
	"os"
	"strings"

	"github.com/alecthomas/kong"
	"github.com/andreykaipov/ynab-simplefin-sync/simplefin"
	"github.com/andreykaipov/ynab-simplefin-sync/sync"
	"github.com/andreykaipov/ynab-simplefin-sync/ynab"
	"github.com/olekukonko/tablewriter"
	ynabgo "go.bmvs.io/ynab"
	"gopkg.in/yaml.v3"
)

var start string
var end string
var out string

type CLI struct {
	Config kong.ConfigFlag `type:"path" help:"Path to a YAML file with defaults"`

	SimpleFINAccessURL string `required:"" env:"SIMPLEFIN_ACCESS_URL" name:"simplefin_access_url" help:"SimpleFIN access URL (claimed from token already)"`
	YnabAccessToken    string `required:"" help:"Your YNAB access token to its API"`
	YnabBudgetID       string `required:"" help:"Your YNAB budget UUID"`

	Sync sync.Cmd `cmd:"" help:"Sync SimpleFIN transactions with your YNAB budget"`

	Ynab struct {
		Accounts ynab.Accounts `cmd:"" help:""`
	} `cmd:"" help:""`

	Simplefin struct {
		Accounts     simplefin.Accounts     `cmd:"" help:""`
		Transactions simplefin.Transactions `cmd:"" help:""`
	} `cmd:"" help:""`
}

func (o *CLI) AfterApply(ctx *kong.Context) error {
	// setup clients
	ctx.Bind(simplefin.Client{AccessURL: o.SimpleFINAccessURL})
	ctx.Bind(o.YnabBudgetID)
	ctx.BindTo(ynabgo.NewClient(o.YnabAccessToken), (*ynabgo.ClientServicer)(nil))

	// setup common table settings
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

	name := "ynab-simplefin-sync"
	ctx := kong.Parse(
		cli,
		kong.Name(name),
		kong.Description("Sync your YNAB accounts with info from SimpleFIN via SimpleFIN Bridge"),
		kong.Configuration(
			yamlEnvResolver,
			fmt.Sprintf("%s.config.yml", name),
			fmt.Sprintf("%s/.%s.yml", home, name),
			fmt.Sprintf("%s/config/%s/config.yml", home, name),
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
