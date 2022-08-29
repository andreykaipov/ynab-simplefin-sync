## ynab-simplefin-sync

Sync your YNAB budget with the accounts you've added to [SimpleFIN
Bridge](https://beta-bridge.simplefin.org/).

## why?

It's an alternative to the YNAB bank import feature for linked accounts.

## usage

It also doubles as a handy CLI for viewing your YNAB and SimpleFIN accounts:

```console
❯ ynab-simplefin-sync --help
Usage: ynab-simplefin-sync --simplefin_access_url=STRING --ynab-access-token=STRING --ynab-budget-id=STRING <command>

Sync your YNAB accounts with info from SimpleFIN via SimpleFIN Bridge

Flags:
  -h, --help                           Show context-sensitive help.
      --config=CONFIG-FLAG             Path to a YAML file with defaults
      --simplefin_access_url=STRING    SimpleFIN access URL (claimed from token already) ($SIMPLEFIN_ACCESS_URL)
      --ynab-access-token=STRING       Your YNAB access token to its API
      --ynab-budget-id=STRING          Your YNAB budget UUID

Commands:
  sync                      Sync SimpleFIN transactions with your YNAB budget
  ynab accounts
  simplefin accounts
  simplefin transactions

Run "ynab-simplefin-sync <command> --help" for more information on a command.
```

Assuming you already have a [SimpleFIN access
URL](https://beta-bridge.simplefin.org/info/developers) (considering you
stumbled across this repo), the easiest way to get started would be to first
rename the `example.config.yml` at the root of this repo to
`ynab-simplefin-sync.config.yml` and edit it with your bugdet and account IDs:

```yaml
---
simplefin_access_url: envexpand:$SIMPLEFIN_ACCESS_URL

ynab_budget_id: uuid
ynab_access_token: envexpand:$YNAB_ACCESS_TOKEN

# no need for the ACT- prefix when specifying the SimpleFIN account UUIDs
#
simplefin_ynab_account_mapping:
  simplefin-uuid-1: ynab-uuid-1
  simplefin-uuid-2: ynab-uuid-2
  simplefin-uuid-3: ynab-uuid-3
  simplefin-uuid-4: ynab-uuid-4
```

You can find the YNAB IDs from the web URL when you navigate your budget (e.g.
`https://app.youneedabudget.com/<budget-id>/accounts/<account-id>`). The
SimpleFIN account IDs can be found by running `simplefin accounts`. The `ACT-`
prefixes can be omitted if you'd like.

Now you're ready to access your financial info from the commandline like a fool:

```console
❯ ynab-simplefin-sync ynab accounts
...

❯ ynab-simplefin-sync simplefin accounts
...

❯ ynab-simplefin-sync simplefin transactions --start "$(date +%s -d-7day)"
...

❯ ynab-simplefin-sync sync --start "$(date +%s -d-7day)"
Imported 5 transactions.
There were 4 duplicates.
```
