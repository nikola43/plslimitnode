package main

import (
	"context"
	"crypto/ecdsa"
	"encoding/hex"
	"fmt"
	"log"
	"math/big"
	"os"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/joho/godotenv"
	nineInchSpotLimitPLS "github.com/nikola43/plslimitnode/NineInchSpotLimitPLS"
)

func main() {
	// load .env file
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}
	wsRPC := os.Getenv("WS_RPC")
	limitOrderAddress := os.Getenv("LIMIT_ADDRESS")
	privateKey := os.Getenv("PRIVATE_KEY")

	/* init eth client and router contract */
	client := initEthClient(wsRPC)
	nineInchLimit := initNineInchSpotLimit(limitOrderAddress, client)

	headers := make(chan *types.Header)
	sub, err := client.SubscribeNewHead(context.Background(), headers)
	if err != nil {
		log.Fatal(err)
	}

	for {
		select {
		case err := <-sub.Err():
			log.Fatal(err)
		case header := <-headers:

			block, err := client.BlockByHash(context.Background(), header.Hash())
			if err != nil {
				log.Fatal(err)
			}

			fmt.Println("Block Number:", block.Number().Uint64()) // 3477413

			shouldBeExecuted, orderId, err := nineInchLimit.CheckUpkeep(nil, []byte("0x"))
			if err != nil {
				log.Fatal(err)
			}
			orderIdHex := hex.EncodeToString(orderId)
			fmt.Println("OrderId:", orderId)
			fmt.Println("OrderIdHex:", orderIdHex)
			fmt.Println("Should be executed:", shouldBeExecuted)
			fmt.Println("")

			if shouldBeExecuted {
				fmt.Println("Perform upkeep")
				txHash, err := performUpkeep(client, nineInchLimit, orderId, privateKey)
				if err != nil {
					log.Fatal(err)
				}
				fmt.Println("Tx Hash:", txHash)
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
		log.Fatal(err)
	}

	publicKey := privateKey.Public()
	publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
	if !ok {
		log.Fatal("cannot assert type: publicKey is not of type *ecdsa.PublicKey")
	}

	fromAddress := crypto.PubkeyToAddress(*publicKeyECDSA)
	transactor := BuildTransactor(client, privateKey, fromAddress, big.NewInt(0), gasPrice, gasLimit)

	tx, err := nineInchLimit.PerformUpkeep(transactor, performData)
	if err != nil {
		fmt.Println(err)
	}

	return tx.Hash().Hex(), nil
}

func BuildTransactor(client *ethclient.Client, privateKey *ecdsa.PrivateKey, fromAddress common.Address, value *big.Int, gasPrice *big.Int, gasLimit uint64) *bind.TransactOpts {
	chainId, err := client.NetworkID(context.Background())
	if err != nil {
		log.Fatal(err)
	}

	transactor, err := bind.NewKeyedTransactorWithChainID(privateKey, chainId)
	if err != nil {
		log.Fatal(err)
	}

	transactor.Value = big.NewInt(0)
	if value.Uint64() > 0 {
		transactor.Value = value
	}

	nonce, err := client.PendingNonceAt(context.Background(), fromAddress)
	if err != nil {
		log.Fatal(err)
	}

	transactor.GasPrice = gasPrice
	transactor.GasLimit = gasLimit
	transactor.Nonce = big.NewInt(int64(nonce))
	transactor.Context = context.Background()
	return transactor
}

func initEthClient(wsRPC string) *ethclient.Client {
	client, err := ethclient.Dial(wsRPC)
	if err != nil {
		log.Fatal(err)
	}
	return client
}

func initNineInchSpotLimit(address string, client *ethclient.Client) *nineInchSpotLimitPLS.NineInchSpotLimitPLS {
	instance, err := nineInchSpotLimitPLS.NewNineInchSpotLimitPLS(common.HexToAddress(address), client)
	if err != nil {
		log.Fatal(err)
	}
	return instance
}
