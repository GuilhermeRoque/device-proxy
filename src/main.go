package main

import (
	"github.com/joho/godotenv"
	"log"
	wsClient "lorawanMgnt/src/client"
)

func main() {
	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)
	log.Println("Loading ENV...")
	err := godotenv.Load(".env")
	if err != nil {
		log.Fatalln(err)
	}
	log.Println("Running...")
	//service := application.ServiceCfg{
	//	Name:         "Default",
	//	DataType:     3,
	//	ChannelType:  3,
	//	ChannelParam: 11,
	//	Acquisition:  2,
	//	Period:       50,
	//}
	//defaultPort := uint8(2)
	//defaultAck := uint8(1)
	//payloadService := []byte{
	//	defaultPort,
	//	(service.DataType << 5) | (service.ChannelType << 1) | (service.ChannelParam >> 3 & 0x1),
	//	(service.ChannelParam << 5) | (service.Acquisition << 2) | (defaultAck << 1) | (uint8(service.Period >> 16)),
	//	uint8(service.Period >> 8),
	//	uint8(service.Period),
	//}
	//log.Printf("%X", payloadService)
	client := wsClient.WsClient{}
	client.Run()
}
