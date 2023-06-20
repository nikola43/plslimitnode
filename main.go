package main

import (
	"context"
	"fmt"
	"log"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"

	nineInchRouter "./NineInchRouter"
)

func main() {
	// https://sepolia.infura.io/v3/65a6236f83cd4a618a4c0d01a8c3bc11
	// wss://sepolia.infura.io/ws/v3/65a6236f83cd4a618a4c0d01a8c3bc11
	client, err := ethclient.Dial("wss://sepolia.infura.io/ws/v3/65a6236f83cd4a618a4c0d01a8c3bc11")
	if err != nil {
		log.Fatal(err)
	}

	address := common.HexToAddress("0x147B8eb97fD247D06C4006D269c90C1908Fb5D54")
	instance, err := nineInchRouter.NewNineInchRouter(address, client)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("contract is loaded")
	_ = instance

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

		}
	}
}
