solc --optimize --optimize-runs 999999 --abi swap/NineInchSpotLimitPLS.sol -o build 
solc --optimize --optimize-runs 999999 --bin swap/NineInchSpotLimitPLS.sol -o build 

/Users/kasiopea/go/bin/abigen --bin=./build/NineInchSpotLimitPLS.bin --abi=./build/NineInchSpotLimitPLS.abi --pkg=NineInchSpotLimitPLS --out=NineInchSpotLimitPLS.go
