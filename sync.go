package main

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/andreykaipov/ynab-simplefin-sync/simplefin"
	"github.com/olekukonko/tablewriter"
	ynabgo "go.bmvs.io/ynab"
	"go.bmvs.io/ynab/api"
	"go.bmvs.io/ynab/api/transaction"
)

type SyncCmd struct {
	SimpleFINYnabAccountMapping map[string]string `mapsep:";" name:"simplefin_ynab_account_mapping" help:"SimpleFIN account IDs mapped to YNAB account IDs"`
	Start                       string            `help:"If given, transactions will be restricted to those on or after this Unix epoch timestamp. Default is 7d ago."`
	End                         string            `help:"If given, transactions will be restricted to those before (but not on) this Unix epoch timestamp."`
}

func (o *SyncCmd) Run(budgetID string, ynab ynabgo.ClientServicer, simplefin simplefin.Client, table *tablewriter.Table) error {
	data, err := simplefin.Get("/accounts", o.Start, o.End)
	if err != nil {
		return err
	}

	txs := []transaction.PayloadTransaction{}

	table.SetHeader([]string{"SimpleFIN Org", "SimpleFIN Account", "YNAB Account"})

	for _, account := range data.Accounts {
		simplefinID := strings.TrimPrefix(account.ID, "ACT-")

		ynabAccountID, ok := o.SimpleFINYnabAccountMapping[simplefinID]
		if !ok {
			return fmt.Errorf("didn't find %s in simplefin ynab mapping", simplefinID)
		}

		ynabAccount, err := ynab.Account().GetAccount(budgetID, ynabAccountID)
		if err != nil {
			return err
		}

		table.Append([]string{account.OrgSlug(), account.NameSlug(), ynabAccount.Name})

		for _, tx := range account.Transactions {
			memo := tx.Memo
			if memo == "" {
				memo = tx.Description
			}

			amount, err := strconv.ParseFloat(tx.Amount, 64)
			if err != nil {
				return fmt.Errorf("Failed parsing amount for tx %s in account %s", tx.Description, account.Name)
			}

			amount = amount * 1000
			date := api.Date{Time: time.Unix(int64(tx.Posted), 0)}

			txs = append(txs, transaction.PayloadTransaction{
				AccountID: ynabAccountID,
				Date:      date,
				Amount:    int64(amount),
				Cleared:   transaction.ClearingStatusCleared,
				Approved:  false,
				PayeeName: ptr(tx.Payee),
				Memo:      ptr(memo),
				FlagColor: ptr(transaction.FlagColorYellow),
				ImportID:  ptr(fmt.Sprintf("sync:%d:%d", int64(amount), tx.Posted)),
			})
		}
	}

	table.Render()

	if len(txs) == 0 {
		fmt.Println("No transactions to import for specified date range.")
		return nil
	}

	resp, err := ynab.Transaction().CreateTransactions(budgetID, txs)
	if err != nil {
		return err
	}

	fmt.Println()
	fmt.Printf("Imported %d transactions.\n", len(resp.TransactionIDs))
	fmt.Printf("There were %d duplicates.\n", len(resp.DuplicateImportIDs))

	return nil
}

func ptr[T any](v T) *T { return &v }
