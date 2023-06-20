./solc-macosx-amd64-v0.6.6 --optimize --optimize-runs 999999 --abi swap/NineInchRouter.sol -o build 
./solc-macosx-amd64-v0.6.6 --optimize --optimize-runs 999999 --bin swap/NineInchRouter.sol -o build 
 
/Users/kasiopea/go/bin/abigen --bin=./build/NineInchRouter.bin --abi=./build/NineInchRouter.abi --pkg=NineInchRouter --out=NineInchRouter.go
