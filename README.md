# ont-multi-transfer

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