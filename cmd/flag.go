package cmd

import (
	"strings"

	"github.com/urfave/cli"
)

const (
	DEFAULT_LOG_LEVEL           = 1
	DEFAULT_LOG_FILE_PATH       = "./Log/"
	DEFAULT_BLOCK_CHAIN_RPC_URL = "http://localhost:8545"
)

var (
	PasswordFlag = cli.StringFlag{
		Name:  "password",
		Usage: "password for wallet",
	}
	NetworkFlag = cli.StringFlag{
		Name:  "network",
		Usage: "mainnet | testnet ",
		Value: "mainnet",
	}
	OperationFlag = cli.StringFlag{
		Name:  "operation",
		Usage: "send | check-excel",
	}

	NewWalletFlag = cli.BoolFlag{
		Name:  "new-wallet",
		Usage: "new wallet",
	}
	BalanceFlag = cli.BoolFlag{
		Name:  "balance",
		Usage: "check balance",
	}
	ExcelFileFlag = cli.StringFlag{
		Name:  "excel-file",
		Usage: "excel file path",
	}
)

// GetFlagName deal with short flag, and return the flag name whether flag name have short name
func GetFlagName(flag cli.Flag) string {
	name := flag.GetName()
	if name == "" {
		return ""
	}
	return strings.TrimSpace(strings.Split(name, ",")[0])
}
