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