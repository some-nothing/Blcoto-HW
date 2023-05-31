package main

import (
	"context"
	"encoding/hex"
	"log"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
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

func sync(db *gorm.DB, client *ethclient.Client) {
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

	for {
		select {
		case header := <-headers:
			addBlockAndTX(header.Hash())
		}
	}

}

func addBlockAndTXByNumber(number uint64) {
	block, err := client.BlockByNumber(context.Background(), big.NewInt(int64(number)))
	if err != nil {
		log.Fatal(err)
	}

	headers <- block.Header()

}

func addBlockAndTX(hash common.Hash) {
	db.Transaction(func(gtx *gorm.DB) error {
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
	})
}
