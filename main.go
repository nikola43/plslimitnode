package main

import (
	"context"
	"crypto/ecdsa"
	"encoding/hex"
	"fmt"
	"math/big"
	"os"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/fatih/color"
	"github.com/joho/godotenv"
	nineInchSpotLimitPLS "github.com/nikola43/plslimitnode/NineInchSpotLimitPLS"
)

func main() {
	// load .env file
	err := godotenv.Load()
	if err != nil {
		fmt.Println("Error loading .env file")
	}
	wsRPC := os.Getenv("WS_RPC")
	limitOrderAddress := os.Getenv("LIMIT_ADDRESS")
	privateKey := os.Getenv("PRIVATE_KEY")

	var client *ethclient.Client
	do := client == nil
	for do {
		// init eth client
		color.Cyan("Connecting to %s", wsRPC)
		client, err = ethclient.Dial(wsRPC)
		if err != nil {
			fmt.Println(color.RedString("Error connecting to %s", wsRPC))
		}
		do = client == nil
	}

	nineInchLimit, err := nineInchSpotLimitPLS.NewNineInchSpotLimitPLS(common.HexToAddress(limitOrderAddress), client)
	if err != nil {
		fmt.Println(err)
	}

	headers := make(chan *types.Header)
	sub, err := client.SubscribeNewHead(context.Background(), headers)
	if err != nil {
		fmt.Println(err)
	}

	color.Green("Connected to %s", wsRPC)
	color.Yellow("Listening for new headers...")

	for {
		select {
		case err := <-sub.Err():
			fmt.Println(err)
		case header := <-headers:

			block, err := client.BlockByHash(context.Background(), header.Hash())
			if err != nil {
				fmt.Println(err)
			}

			shouldBeExecuted, orderId, err := nineInchLimit.CheckUpkeep(nil, []byte("0x"))
			if err != nil {
				fmt.Println(err)
			}

			fmt.Println(color.YellowString("Block Number"), color.CyanString("%d", block.Number().Uint64()))
			fmt.Println(color.YellowString("OrderId"), color.CyanString(hex.EncodeToString(orderId)))
			fmt.Println(color.YellowString("Should be executed"), color.CyanString(fmt.Sprintf("%t", shouldBeExecuted)))
			fmt.Println("")

			if shouldBeExecuted {
				fmt.Println(color.YellowString("Performing upkeep..."))
				txHash, err := performUpkeep(client, nineInchLimit, orderId, privateKey)
				if err != nil {
					fmt.Println(err)
				}
				fmt.Println(color.YellowString("Tx Hash"), color.CyanString(txHash))
				fmt.Println("")
			}
		}
	}
}

func performUpkeep(client *ethclient.Client, nineInchLimit *nineInchSpotLimitPLS.NineInchSpotLimitPLS, performData []byte, pk string) (string, error) {
	// calculate gas and gas limit
	gasLimit := uint64(5300000) // in units
	gasPrice, gasPriceErr := client.SuggestGasPrice(context.Background())
	if gasPriceErr != nil {
		fmt.Println(gasPriceErr)
	}

	privateKey, err := crypto.HexToECDSA(pk)
	if err != nil {
		fmt.Println(err)
	}

	publicKey := privateKey.Public()
	publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
	if !ok {
		fmt.Println("cannot assert type: publicKey is not of type *ecdsa.PublicKey")
	}

	fromAddress := crypto.PubkeyToAddress(*publicKeyECDSA)
	transactor := buildTransactor(client, privateKey, fromAddress, big.NewInt(0), gasPrice, gasLimit)

	tx, err := nineInchLimit.PerformUpkeep(transactor, performData)
	if err != nil {
		fmt.Println(err)
	}

	return tx.Hash().Hex(), nil
}

func buildTransactor(client *ethclient.Client, privateKey *ecdsa.PrivateKey, fromAddress common.Address, value *big.Int, gasPrice *big.Int, gasLimit uint64) *bind.TransactOpts {
	chainId, err := client.NetworkID(context.Background())
	if err != nil {
		fmt.Println(err)
	}

	transactor, err := bind.NewKeyedTransactorWithChainID(privateKey, chainId)
	if err != nil {
		fmt.Println(err)
	}

	transactor.Value = big.NewInt(0)
	if value.Uint64() > 0 {
		transactor.Value = value
	}

	nonce, err := client.PendingNonceAt(context.Background(), fromAddress)
	if err != nil {
		fmt.Println(err)
	}

	transactor.GasPrice = gasPrice
	transactor.GasLimit = gasLimit
	transactor.Nonce = big.NewInt(int64(nonce))
	transactor.Context = context.Background()
	return transactor
}
