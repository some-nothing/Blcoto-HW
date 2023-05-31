package main

import (
	"context"
	"encoding/hex"
	"fmt"
	"log"

	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func main() {
	config := LoadConfig()

	db, err := gorm.Open(mysql.Open(config.DatabaseURL), &gorm.Config{
		SkipDefaultTransaction: true,
	})
	if err != nil {
		log.Fatal(err)
	}

	db.AutoMigrate(&Block{}, &Transaction{})

	client, err := ethclient.Dial(config.EndpointURL)
	if err != nil {
		log.Fatal(err)
	}

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
			db.Transaction(func(tx *gorm.DB) error {
				block, err := client.BlockByHash(context.Background(), header.Hash())
				if err != nil {
					return err
				}

				fmt.Println("block: ", block.Number().Uint64())

				db.Create(&Block{
					Hash:       block.Hash().Hex(),
					Number:     block.Number().Uint64(),
					Timestamp:  block.Time(),
					ParentHash: block.ParentHash().Hex(),
				})

				for _, tx := range block.Transactions() {
					fmt.Println(tx.Hash().Hex())
					sender, _ := ParseSender(tx)
					db.Create(&Transaction{
						Hash:      tx.Hash().Hex(),
						BlockHash: block.Hash().Hex(),
						Type:      int(tx.Type()),
						From:      sender,
						To:        tx.To().Hex(),
						Value:     tx.Value().String(),
						Input:     "0x" + hex.EncodeToString(tx.Data()),
					})
				}

				return nil
			})

		}
	}
}
