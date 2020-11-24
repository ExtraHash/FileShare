package main

import (
	"os"
	"time"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type db struct {
	db *gorm.DB
}

type apiModel struct {
	ID        uint           `gorm:"primarykey" json:"-"`
	CreatedAt time.Time      `json:"-"`
	UpdatedAt time.Time      `json:"-"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

type FileMessage struct {
	Type string    `json:"type"`
	ID   string    `json:"id"`
	File File      `json:"file"`
	Time time.Time `json:"time"`
}

// File is an uploaded file
type File struct {
	apiModel
	ID       string `json:"id" gorm:"unique"`
	FileName string `json:"fileName"`
	Data     []byte `json:"data" gorm:"-"`
}

func (d *db) initialize() {
	if !fileExists(dataFolder) {
		os.Mkdir(dataFolder, 0700)
	}
	if !fileExists(fileFolder) {
		os.Mkdir(fileFolder, 0700)
	}

	// initialize database
	db, err := gorm.Open(sqlite.Open(dataFolder+"/db.sqlite"), &gorm.Config{})
	if err != nil {
		panic(err)
	}

	db.AutoMigrate(&File{})

	d.db = db
}
