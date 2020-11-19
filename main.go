package main

import (
	"crypto/rand"
	"encoding/hex"
	"flag"
	"fmt"
	"time"

	"github.com/ExtraHash/p2p"
)

func main() {
	port := flag.Int("port", 10187, "--port 10187")
	logLevel := flag.Int("log-level", 0, "--log-level 0")
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

	p2p := p2p.DP2P{}
	go p2p.Initialize(config)
	go listen(&p2p)
	chatter(&p2p)
}

func chatter(p2p *p2p.DP2P) {
	for {
		time.Sleep(5 * time.Second)
		p2p.Broadcast(randomData())
	}
}

func listen(p2p *p2p.DP2P) {
	for {
		message := p2p.ReadMessage()
		fmt.Println("received", string(message))
	}
}

func randomData() []byte {
	token := make([]byte, 32)
	rand.Read(token)
	// return token
	return []byte(hex.EncodeToString(token))
}
