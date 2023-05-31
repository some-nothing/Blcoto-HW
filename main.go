package main

import (
	"blocto/models"
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/gin-gonic/gin"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func main() {
	db, client = initial()
	go sync(db, client)

	r := gin.Default()
	r.GET("/blocks", func(c *gin.Context) {
		// get param limit
		limit, err := strconv.Atoi(c.Query("limit"))
		fmt.Println(limit)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"message": "limit must be integer",
			})
			return
		}

		type Result struct {
			Hash       string `json:"hash"`
			Number     uint64 `json:"number"`
			ParentHash string `json:"parent_hash"`
			Timestamp  uint64 `json:"timestamp"`
		}
		var result []Result
		db.Model(&models.Block{}).Select("hash", "number", "parent_hash", "timestamp").Order("number desc").Limit(limit).Find(&result)

		c.JSON(http.StatusOK, gin.H{
			"blocks": result,
		})
	})
	r.Run() // listen and serve on 0.0.0.0:8080 (for windows "localhost:8080")
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
