package ynab

import (
	"fmt"
	"sort"

	"github.com/olekukonko/tablewriter"
	ynabgo "go.bmvs.io/ynab"
)

type Accounts struct{}

func (o *Accounts) AfterApply() error { return nil }
func (o *Accounts) Run(budgetID string, client ynabgo.ClientServicer, table *tablewriter.Table) error {
	accounts, err := client.Account().GetAccounts(budgetID)
	if err != nil {
		return err
	}

	table.SetHeader([]string{
		"ID",
		"Name",
		"Type",
		"On Budget",
		"Closed",
		"Working Balance",
		"Cleared Balance",
		"Uncleared Balance",
	})
	table.SetColumnAlignment([]int{
		tablewriter.ALIGN_LEFT,
		tablewriter.ALIGN_LEFT,
		tablewriter.ALIGN_LEFT,
		tablewriter.ALIGN_LEFT,
		tablewriter.ALIGN_LEFT,
		tablewriter.ALIGN_RIGHT,
		tablewriter.ALIGN_RIGHT,
		tablewriter.ALIGN_RIGHT,
	})

	rows := [][]string{}
	for _, account := range accounts {
		rows = append(rows, []string{
			account.ID,
			account.Name,
			string(account.Type),
			fmt.Sprintf("%t", account.OnBudget),
			fmt.Sprintf("%t", account.Closed),
			milliunitsToUSD(account.Balance),
			milliunitsToUSD(account.ClearedBalance),
			milliunitsToUSD(account.UnclearedBalance),
		})
	}

	sort.Slice(rows, func(i, j int) bool {
		a := rows[i]
		b := rows[j]

		// on budget accounts first
		if a[3] < b[3] {
			return false
		} else if a[3] > b[3] {
			return true
		}

		// on budget is equal, check type next
		if a[2] < b[2] {
			return true
		} else if a[2] > b[2] {
			return false
		}

		// on budget equal, type is equal, just alphabetize now
		return a[1] < b[1]
	})

	table.AppendBulk(rows)
	table.Render()

	return nil
}

func milliunitsToUSD(a int64) string {
	prefix := ""
	if a < 0 {
		prefix = "-"
		a = -a
	}

	cents := a % 1000 / 10
	dollars := a / 1000
	return fmt.Sprintf("%s$%d.%d", prefix, dollars, cents)
}
