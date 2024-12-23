# ont-multi-transfer

## Build

```
go build
```

## Template

Input data to template.xlsx

*** NOTE: ONLY 1 SHEET ALLOWED ***

address: ontology address start with 'A'

amount: amount without decimal . for example '0.1'

token: "ONT" or "ONG" for now


## Use case
create wallet
```
./ont-multi-transfer --new-wallet --password 123456
```

check balance
```
./ont-multi-transfer --balance --password 123456 --network testnet
```

check excel
```
./ont-multi-transfer --operation check-excel --excel-file ./template1.xlsx
```

send token
```
./ont-multi-transfer --operation send --excel-file ./template1.xlsx --network testnet --password 123456
```

```
GLOBAL OPTIONS:
   --network value     mainnet | testnet  (default: "mainnet")
   --new-wallet        new wallet
   --password value    password for wallet
   --balance           check balance
   --operation value   send | check-excel
   --excel-file value  excel file path
   --help, -h          show help
```   