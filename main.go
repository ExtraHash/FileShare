package main

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"flag"
	"io/ioutil"
	"log"
	"time"

	"github.com/ExtraHash/p2p"
)

var dataFolder = "data"
var fileFolder = dataFolder + "/files"

func main() {
	port := flag.Int("port", 10187, "--port 10187")
	logLevel := flag.Int("log-level", 0, "--log-level 0")
	testChatter := flag.Bool("chatter", false, "--chatter")
	flag.Parse()

	seeds := []p2p.Peer{
		{Host: "10.0.0.148", Port: 10187, SignKey: "c2bc4d085b46c61bfabf7e0c2809d7aba7421ad9057148d9831c2463a2b61f80"},
	}

	config := p2p.NetworkConfig{
		Port:      *port,
		LogLevel:  *logLevel,
		NetworkID: "35c36251-96b7-4e2a-b0bb-de40223d3034",
		Seeds:     seeds,
	}

	db := db{}
	db.initialize()

	p2p := p2p.DP2P{}
	go p2p.Initialize(config)

	if *testChatter {
		go chatter(&p2p)
	}

	api := api{}
	api.initialize(&p2p, &db)
	go listen(&p2p, &db, &api)
	api.run()
}

func chatter(p2p *p2p.DP2P) {
	for {
		time.Sleep(5 * time.Second)
		p2p.Broadcast(randomData())
	}
}

func listen(p2p *p2p.DP2P, db *db, api *api) {
	for {
		message := p2p.ReadMessage()
		api.emit(message)

		file := File{}
		json.Unmarshal(message, &file)

		checkFile := File{}
		db.db.Find(&checkFile, "id = ?", file.ID)

		if checkFile.ID == file.ID {
			continue
		}

		db.db.Create(&file)
		err := ioutil.WriteFile(fileFolder+"/"+file.ID, file.Data, 0666)
		if err != nil {
			log.Print(err)
		}
	}
}

func randomData() []byte {
	token := make([]byte, 32)
	rand.Read(token)
	// return token
	return []byte(hex.EncodeToString(token))
}
