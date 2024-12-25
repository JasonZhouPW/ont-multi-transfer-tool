package main

import (
	"fmt"
	"math/big"
	"ont-multi-transfer/cmd"
	"os"
	"runtime"
	"strconv"
	"strings"

	"github.com/ethereum/go-ethereum/common/hexutil"
	sdkcom "github.com/ontio/ontology-go-sdk/common"
	"github.com/ontio/ontology/common"

	ontsdk "github.com/ontio/ontology-go-sdk"
	"github.com/urfave/cli"
	"github.com/xuri/excelize/v2"
)

const (
	BATCH_SIZE = 500
	gasPrice   = uint64(2500)
	gasLimit   = uint64(600000)
)

func main() {
	println("Starting ONT Multi Transfer Agent...")
	if err := setupAPP().Run(os.Args); err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}

}

func setupAPP() *cli.App {
	app := cli.NewApp()
	app.Usage = "ont multi transfer tool"
	app.Action = startAgent
	app.Flags = []cli.Flag{

		cmd.NetworkFlag,
		cmd.NewWalletFlag,
		cmd.PasswordFlag,
		cmd.BalanceFlag,
		cmd.OperationFlag,
		cmd.ExcelFileFlag,
	}
	app.Before = func(context *cli.Context) error {
		runtime.GOMAXPROCS(runtime.NumCPU())
		return nil
	}
	return app
}

func startAgent(ctx *cli.Context) {

	osdk := ontsdk.NewOntologySdk()
	network := ctx.GlobalString(cmd.GetFlagName(cmd.NetworkFlag))
	if strings.EqualFold(network, "mainnet") {
		osdk.NewRpcClient().SetAddress("http://dappnode1.ont.io:20336")
	} else {
		osdk.NewRpcClient().SetAddress("http://polaris1.ont.io:20336")
	}

	newWallet := ctx.GlobalBool(cmd.GetFlagName(cmd.NewWalletFlag))
	if newWallet {
		wallet, err := osdk.CreateWallet("./wallet.dat")
		if err != nil {
			fmt.Printf("create wallet error: %v\n", err)
			return
		}
		password := ctx.GlobalString(cmd.GetFlagName(cmd.PasswordFlag))
		if len(password) == 0 {
			fmt.Println("Please input password.")
			return
		}
		acct, err := wallet.NewDefaultSettingAccount([]byte(password))
		if err != nil {
			fmt.Printf("new account error: %v\n", err)
			return
		}
		err = wallet.Save()
		if err != nil {
			fmt.Printf("save wallet error: %v\n", err)
			return
		}
		fmt.Printf("New account address: %s\n", acct.Address.ToBase58())
		return
	} else {

		if ctx.GlobalBool(cmd.GetFlagName(cmd.BalanceFlag)) {
			wallet, err := osdk.OpenWallet("./wallet.dat")
			if err != nil {
				fmt.Printf("open wallet error: %v\n", err)
				return
			}
			password := ctx.GlobalString(cmd.GetFlagName(cmd.PasswordFlag))
			if len(password) == 0 {
				fmt.Println("Please input password.")
				return
			}
			defaultAcct, err := wallet.GetDefaultAccount([]byte(password))
			if err != nil {
				fmt.Printf("get default account error: %v\n", err)
				return
			}
			ontBalance, err := osdk.Native.Ont.BalanceOfV2(defaultAcct.Address)
			if err != nil {
				fmt.Printf("get balance error: %v\n", err)
				return
			}

			ongBalance, err := osdk.Native.Ong.BalanceOfV2(defaultAcct.Address)
			if err != nil {
				fmt.Printf("get balance error: %v\n", err)
				return
			}
			fmt.Printf("Account address: %s\n", defaultAcct.Address.ToBase58())
			fmt.Printf("Network: %s\n", network)
			fmt.Printf("ONT balance: %s\n", ToStringByPrecise(ontBalance, 9))
			fmt.Printf("ONG balance: %s\n", ToStringByPrecise(ongBalance, 18))
			return
		}
		operation := ctx.GlobalString(cmd.GetFlagName(cmd.OperationFlag))
		excelFilePath := ctx.GlobalString(cmd.GetFlagName(cmd.ExcelFileFlag))
		if strings.EqualFold(operation, "check-excel") {
			if excelFilePath == "" {
				fmt.Println("Please input excel file path.")
				return
			}
			rows, err := readExcel(excelFilePath)
			if err != nil {
				fmt.Printf("read excel error: %v\n", err)
				return
			}
			fmt.Printf("Total rows: %d\n", len(rows))

			ontAmt, ongAmt, err := extractTotalAmount(rows)
			if err != nil {
				fmt.Printf("extract total amount error: %v\n", err)
				return
			}
			fmt.Printf("Total ONG amount: %f\n", ongAmt)
			fmt.Printf("Total ONT amount: %f\n", ontAmt)
			return
		}
		if strings.EqualFold(operation, "send") {
			wallet, err := osdk.OpenWallet("./wallet.dat")
			if err != nil {
				fmt.Printf("open wallet error: %v\n", err)
				return
			}
			password := ctx.GlobalString(cmd.GetFlagName(cmd.PasswordFlag))
			if len(password) == 0 {
				fmt.Println("Please input password.")
				return
			}
			defaultAcct, err := wallet.GetDefaultAccount([]byte(password))
			if err != nil {
				fmt.Printf("get default account error: %v\n", err)
				return
			}
			fmt.Printf("Account address: %s\n", defaultAcct.Address.ToBase58())
			rows, err := readExcel(excelFilePath)
			if err != nil {
				fmt.Printf("read excel error: %v\n", err)
				return
			}
			ontAmt, ongAmt, err := extractTotalAmount(rows)
			if err != nil {
				fmt.Printf("extract total amount error: %v\n", err)
				return
			}
			ontBalance, err := osdk.Native.Ont.BalanceOfV2(defaultAcct.Address)
			if err != nil {
				fmt.Printf("get balance error: %v\n", err)
				return
			}

			ongBalance, err := osdk.Native.Ong.BalanceOfV2(defaultAcct.Address)
			if err != nil {
				fmt.Printf("get balance error: %v\n", err)
				return
			}
			requiredOntAmt := ToIntByPrecise(fmt.Sprintf("%f", ontAmt), 9)

			requiredOngAmt := ToIntByPrecise(fmt.Sprintf("%f", ongAmt), 18)
			if requiredOntAmt.Cmp(ontBalance) > 0 {
				fmt.Println("Insufficient ONT balance.")
				return
			}
			if requiredOngAmt.Cmp(ongBalance) > 0 {
				fmt.Println("Insufficient ONG balance.")
				return
			}

			ontRows, ongRows, err := splitByToken(rows)
			if err != nil {
				fmt.Printf("split by token error: %v\n", err)
				return
			}
			if len(ontRows) > 0 {
				err = sendByToken(ontRows, "ONT", defaultAcct, osdk)
				if err != nil {
					fmt.Printf("send ONT error: %v\n", err)
					return
				}
				fmt.Println("Send ONT success.")
			}
			if len(ongRows) > 0 {
				err = sendByToken(ongRows, "ONG", defaultAcct, osdk)
				if err != nil {
					fmt.Printf("send ONG error: %v\n", err)
					return
				}
				fmt.Println("Send ONG success.")
			}
		}

	}

}

func readExcel(filePath string) ([][]string, error) {
	excelFile, err := excelize.OpenFile(filePath)
	if err != nil {
		fmt.Printf("open excel file error: %v\n", err)
		return nil, err
	}
	sc := excelFile.SheetCount
	fmt.Printf("Sheet count: %d\n", sc)
	if sc > 1 {
		fmt.Printf("Only support one sheet.\n")
		return nil, fmt.Errorf("only support one sheet")
	}
	sn := excelFile.GetSheetName(0)
	excelFile.SetActiveSheet(0)
	rows, err := excelFile.GetRows(sn)
	if err != nil {
		fmt.Printf("get rows error: %v\n", err)
		return nil, err
	}
	// fmt.Printf("Total rows: %d\n", len(rows))
	return rows[1:], nil

}

func extractTotalAmount(rows [][]string) (float64, float64, error) {
	ongAmt, ontAmt := float64(0), float64(0)
	for i, row := range rows {
		if len(row) < 3 {
			fmt.Printf("Row %d has less than 3 columns.\n", i+1)
			return ongAmt, ontAmt, fmt.Errorf("invalid row format at row %d", i+1)
		}
		amt, err := strconv.ParseFloat(row[1], 64)
		if err != nil {
			fmt.Printf("Parse amount error at row %d: %v\n", i+1, err)
			return ongAmt, ontAmt, err

		}
		if strings.EqualFold(row[2], "ONG") {
			ongAmt += amt
		} else if strings.EqualFold(row[2], "ONT") {
			ontAmt += amt
		} else {
			fmt.Printf("Invalid token at row %d.\n", i+2)
			return ongAmt, ontAmt, fmt.Errorf("invalid token at row %d", i+2)
		}
	}
	return ontAmt, ongAmt, nil
}

func splitByToken(origin [][]string) ([][]string, [][]string, error) {

	ongRows := make([][]string, 0)
	ontRows := make([][]string, 0)
	for i, row := range origin {
		if strings.EqualFold(row[2], "ONG") {
			ongRows = append(ongRows, row)
		} else if strings.EqualFold(row[2], "ONT") {
			ontRows = append(ontRows, row)
		} else {
			fmt.Printf("Invalid token.\n")
			return nil, nil, fmt.Errorf("invalid token at row %d", i+2)
		}
	}
	return ontRows, ongRows, nil
}

func ToIntByPrecise(str string, precise uint64) *big.Int {
	result := new(big.Int)
	splits := strings.Split(str, ".")
	if len(splits) == 1 { // doesn't contain "."
		var i uint64 = 0
		for ; i < precise; i++ {
			str += "0"
		}
		intValue, ok := new(big.Int).SetString(str, 10)
		if ok {
			result.Set(intValue)
		}
	} else if len(splits) == 2 {
		value := new(big.Int)
		ok := false
		floatLen := uint64(len(splits[1]))
		if floatLen <= precise { // add "0" at last of str
			parseString := strings.Replace(str, ".", "", 1)
			var i uint64 = 0
			for ; i < precise-floatLen; i++ {
				parseString += "0"
			}
			value, ok = value.SetString(parseString, 10)
		} else { // remove redundant digits after "."
			splits[1] = splits[1][:precise]
			parseString := splits[0] + splits[1]
			value, ok = value.SetString(parseString, 10)
		}
		if ok {
			result.Set(value)
		}
	}

	return result
}

func sendByToken(rows [][]string, token string, defaultAcct *ontsdk.Account, sdk *ontsdk.OntologySdk) error {

	cnt := len(rows) / BATCH_SIZE
	if len(rows)%BATCH_SIZE != 0 {
		cnt += 1
	}
	for i := 0; i < cnt; i++ {
		start, end := i*BATCH_SIZE, (i+1)*BATCH_SIZE
		if end > len(rows) {
			end = len(rows)
		}
		batchRows := rows[start:end]
		fmt.Printf("Batch %d: %d - %d\n", i+1, start+1, end)
		var states []*sdkcom.TransferStateV2
		for j, row := range batchRows {
			fmt.Printf("Row %d: %s %s %s\n", j+start+2, row[0], row[1], row[2])
			toAddr, err := parseToAddress(row[0])
			if err != nil {
				fmt.Printf("Parse to address error at row %d: %v\n", j+start+2, err)
				return err
			}
			decimals := uint64(0)
			if strings.EqualFold(row[2], "ONG") {
				decimals = 18
			} else if strings.EqualFold(row[2], "ONT") {
				decimals = 9
			} else {
				fmt.Printf("Invalid token at row %d.\n", j+start+2)
				return fmt.Errorf("invalid token at row %d.\n", j+start+2)
			}
			value := ToIntByPrecise(row[1], decimals)
			state := &sdkcom.TransferStateV2{
				From:  defaultAcct.Address,
				To:    toAddr,
				Value: value,
			}
			states = append(states, state)
		}
		if strings.EqualFold(token, "ONT") {
			txhash, err := sdk.Native.Ont.MultiTransferV2(gasPrice, gasLimit, defaultAcct, states, defaultAcct)
			if err != nil {
				fmt.Printf("ONT transfer error: %v\n", err)
				return err
			}
			fmt.Printf("sent txhash: %s\n", txhash.ToHexString())

		} else if strings.EqualFold(token, "ONG") {
			txhash, err := sdk.Native.Ong.MultiTransferV2(gasPrice, gasLimit, states, defaultAcct)
			if err != nil {
				fmt.Printf("ONG transfer error: %v\n", err)
				return err
			}
			fmt.Printf("sent txhash: %s\n", txhash.ToHexString())

		}
	}
	return nil
}

func parseToAddress(str string) (common.Address, error) {
	if strings.HasPrefix(str, "0x") {
		bts, err := hexutil.Decode(str)
		if err != nil {
			fmt.Printf("err: %s\n", err)
			return common.Address{}, err
		}
		addr, err := common.AddressParseFromBytes(bts)
		if err != nil {
			fmt.Printf("err: %s\n", err)
			return common.Address{}, err
		}
		fmt.Printf("change evm address: %s - > %s\n", str, addr.ToBase58())
		return addr, nil
	}
	addr, err := common.AddressFromBase58(str)
	if err != nil {
		fmt.Printf("Parse to address error: %v\n", err)
		return common.Address{}, err
	}
	return addr, nil
}

func ToStringByPrecise(bigNum *big.Int, precise uint64) string {
	result := ""
	destStr := bigNum.String()
	destLen := uint64(len(destStr))
	if precise >= destLen { // add "0.000..." at former of destStr
		var i uint64 = 0
		prefix := "0."
		for ; i < precise-destLen; i++ {
			prefix += "0"
		}
		result = prefix + destStr
	} else { // add "."
		pointIndex := destLen - precise
		result = destStr[0:pointIndex] + "." + destStr[pointIndex:]
	}
	result = removeZeroAtTail(result)
	return result
}
func removeZeroAtTail(str string) string {
	i := len(str) - 1
	for ; i >= 0; i-- {
		if str[i] != '0' {
			break
		}
	}
	str = str[:i+1]
	// delete "." at last of result
	if str[len(str)-1] == '.' {
		str = str[:len(str)-1]
	}
	return str
}
