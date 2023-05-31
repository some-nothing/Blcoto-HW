package main

import (
	"context"
	"encoding/hex"
	"fmt"
	"log"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"

	"blocto/models"
)

var (
	Config  ConfigStruct
	db      *gorm.DB
	client  *ethclient.Client
	headers chan *types.Header
	numbers chan uint64
)

func main() {
	db, client = initial()

	var (
		block               models.Block
		latestPositionLocal uint64
	)
	headers = make(chan *types.Header)
	numbers = make(chan uint64)
	// get latest block number from database or config
	var count int64
	if db.Model(&models.Block{}).Count(&count); count == 0 {
		latestPositionLocal = Config.StartPosition
	} else {
		db.Model(&models.Block{}).Select("number").Order("number desc").First(&block)
		latestPositionLocal = block.Number
	}

	latestPositionChain, err := client.BlockNumber(context.Background())
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(latestPositionChain, latestPositionLocal)

	go func() {
		for {
			select {
			case number := <-numbers:
				addBlockAndTXByNumber(number)
			}
		}
	}()

	for i := latestPositionLocal + 1; i <= latestPositionChain; i = i + 1 {
		go func(i uint64) {
			numbers <- i
		}(i)
	}

	// Subscribe to new headers
	client.SubscribeNewHead(context.Background(), headers)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("subscribe new head")

	for {
		select {
		case header := <-headers:
			addBlockAndTX(header.Hash())
		}
	}

}

func initial() (*gorm.DB, *ethclient.Client) {
	Config = LoadConfig()

	db, err := gorm.Open(mysql.Open(Config.DatabaseURL), &gorm.Config{
		SkipDefaultTransaction: true,
	})
	if err != nil {
		log.Fatal(err)
	}

	db.AutoMigrate(&models.Block{}, &models.Transaction{})

	client, err := ethclient.Dial(Config.EndpointURL)
	if err != nil {
		log.Fatal(err)
	}

	return db, client
}

func addBlockAndTXByNumber(number uint64) {
	block, err := client.BlockByNumber(context.Background(), big.NewInt(int64(number)))
	if err != nil {
		log.Fatal(err)
	}

	headers <- block.Header()

}

func addBlockAndTX(hash common.Hash) error {
	// db.Transaction(func(gtx *gorm.DB) error {
	fmt.Println("addBlockAndTX start")
	block, err := client.BlockByHash(context.Background(), hash)
	if err != nil {
		return err
	}

	db.Create(&models.Block{
		Hash:       block.Hash().Hex(),
		Number:     block.Number().Uint64(),
		Timestamp:  block.Time(),
		ParentHash: block.ParentHash().Hex(),
	})

	for _, tx := range block.Transactions() {
		fmt.Println(tx.Hash().Hex())
		sender, _ := ParseSender(tx)

		toAddress := ""
		if tx.To() != nil {
			toAddress = tx.To().Hex()
		}

		inputData := "0"
		if tx.Data() != nil {
			inputData = "0x" + hex.EncodeToString(tx.Data())
		}

		db.Create(&models.Transaction{
			Hash:      tx.Hash().Hex(),
			BlockHash: block.Hash().Hex(),
			Type:      int(tx.Type()),
			From:      sender,
			To:        toAddress,
			Value:     tx.Value().String(),
			Input:     inputData,
		})
	}

	return nil
	// })
}
