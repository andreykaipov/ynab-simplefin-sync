package simplefin

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/olekukonko/tablewriter"
)

type Accounts struct{}

func (o *Accounts) AfterApply() error { return nil }
func (o *Accounts) Run(simplefin Client, table *tablewriter.Table) error {
	data, err := simplefin.Get("/accounts")
	if err != nil {
		return err
	}

	table.SetHeader([]string{"ID", "Organization", "Account", "Balance", "Available Balance"})
	for _, account := range data.Accounts {
		table.Append([]string{account.ID, account.OrgSlug(), account.NameSlug(), account.Balance, account.AvailableBalance})
	}

	table.Render()

	return nil
}

type Transactions struct {
	Start string `help:"If given, transactions will be restricted to those on or after this Unix epoch timestamp. Default is 7d ago."`
	End   string `help:"If given, transactions will be restricted to those before (but not on) this Unix epoch timestamp."`
}

func (o *Transactions) AfterApply() error { return nil }
func (o *Transactions) Run(simplefin Client, table *tablewriter.Table) error {
	data, err := simplefin.Get("/accounts", o.Start, o.End)
	if err != nil {
		return err
	}

	table.SetHeader([]string{"Account", "Date", "Payee", "Memo", "Amount"})

	for _, account := range data.Accounts {
		fqAccount := account.OrgSlug() + "_" + account.NameSlug()

		for _, tx := range account.Transactions {
			date := time.Unix(int64(tx.Posted), 0).Format("01-02-2006")

			memo := tx.Memo
			if memo == "" {
				memo = tx.Description
			}

			table.Append([]string{fqAccount, date, tx.Payee, memo, tx.Amount})
		}
	}

	table.Render()

	return nil
}

type Client struct{ AccessURL string }

type Response struct {
	Errors   []string  `json:"errors"`
	Accounts []Account `json:"accounts"`
}
type Org struct {
	Domain  string `json:"domain"`
	Name    string `json:"name"`
	SfinURL string `json:"sfin-url"`
	URL     string `json:"url"`
}
type Transaction struct {
	ID          string `json:"id"`
	Posted      int    `json:"posted"`
	Amount      string `json:"amount"`
	Description string `json:"description"`
	Payee       string `json:"payee"`
	Memo        string `json:"memo"`
}
type Account struct {
	Org              Org           `json:"org"`
	ID               string        `json:"id"`
	Name             string        `json:"name"`
	Currency         string        `json:"currency"`
	Balance          string        `json:"balance"`
	AvailableBalance string        `json:"available-balance"`
	BalanceDate      int           `json:"balance-date"`
	Transactions     []Transaction `json:"transactions"`
}

func (o *Client) Get(path string, dates ...string) (*Response, error) {
	var start string
	var end string

	if len(dates) > 2 {
		return nil, fmt.Errorf("Expected only start and end dates")
	}
	if len(dates) > 0 {
		start = dates[0]
	}
	if len(dates) > 1 {
		end = dates[1]
	}

	params := url.Values{}
	params.Add("start-date", start)
	params.Add("end-date", end)
	endpoint := fmt.Sprintf("%s%s?%s", o.AccessURL, path, params.Encode())

	resp, err := http.Get(endpoint)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	parsed := &Response{}
	if err := json.Unmarshal(body, parsed); err != nil {
		return nil, err
	}

	return parsed, nil
}

func (o *Account) OrgSlug() string  { return slug(o.Org.Name) }
func (o *Account) NameSlug() string { return slug(o.Name) }

func slug(s string) string {
	var onlyalpha strings.Builder
	for i := 0; i < len(s); i++ {
		b := s[i]
		if ('a' <= b && b <= 'z') ||
			('A' <= b && b <= 'Z') ||
			('0' <= b && b <= '9') ||
			b == ' ' {
			onlyalpha.WriteByte(b)
		}
	}

	lower := strings.ToLower(onlyalpha.String())
	nospace := strings.ReplaceAll(lower, " ", "-")

	return nospace
}
