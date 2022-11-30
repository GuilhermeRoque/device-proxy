package main

import (
	"github.com/joho/godotenv"
	"log"
	wsClient "lorawanMgnt/client"
)

func main() {
	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)
	log.Println("Loading ENV...")
	err := godotenv.Load(".env")
	if err != nil {
		log.Fatalln(err)
	}
	log.Println("Running...")
	client := wsClient.WsClient{}
	client.Run()
}
