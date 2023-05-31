package main

import "gorm.io/gorm"

type Block struct {
	gorm.Model
	Hash       string `gorm:"uniqueIndex,primaryKey"`
	Number     uint64 `gorm:"uniqueIndex"`
	Timestamp  uint64
	ParentHash string `gorm:"unique"`
	Finilized  bool
}

type Transaction struct {
	gorm.Model
	Hash      string `gorm:"uniqueIndex,primaryKey"`
	BlockHash string
	Type      int
	From      string
	To        string
	Value     string
	Input     string
}
