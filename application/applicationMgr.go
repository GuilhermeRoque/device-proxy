package application

import (
	"log"
	"lorawanMgnt/ttn"
	"os"
	"strconv"
)

type Application struct {
	Name          string   `json:"name"`
	APIKey        string   `json:"apiKey"`
	ApplicationId string   `json:"applicationId"`
	Bucket        string   `json:"bucket"`
	Token         string   `json:"token"`
	Devices       []Device `json:"devices"`
}

type Device struct {
	Name       string     `json:"name"`
	DeviceId   string     `json:"devId"`
	DeviceEUI  string     `json:"devEUI"`
	Service    ServiceCfg `json:"serviceProfile"`
	Configured bool       `json:"configured"`
}

type ServiceCfg struct {
	Name         string `json:"name"`
	DataType     uint8  `json:"dataType"`
	ChannelType  uint8  `json:"channelType"`
	ChannelParam uint8  `json:"channelParam"`
	Acquisition  uint8  `json:"acquisition"`
	Period       uint32 `json:"period"`
}

type TtnApplication struct {
	app       *Application
	ttnClient *ttn.TtnClient
}

type ApplicationMgr struct {
	ttnApplications []*TtnApplication
}

func (appMgr *ApplicationMgr) AddApplication(app *Application) error {
	var brokerTTN = os.Getenv("BROKER_TTN")
	var portTTN = os.Getenv("PORT_TTN")
	portTTNInt, _ := strconv.Atoi(portTTN)
	ttnClient := ttn.NewTtnClient(brokerTTN, uint32(portTTNInt), app.Name, app.APIKey)
	log.Printf("Connecting to broker %s:%s to listen to app %s with API key %s\n", brokerTTN, portTTN, app.Name, app.APIKey)
	ttnClient.Connect()
	ttnApplication := &TtnApplication{
		app:       app,
		ttnClient: ttnClient,
	}
	appMgr.ttnApplications = append(appMgr.ttnApplications, ttnApplication)
	for _, device := range ttnApplication.app.Devices {
		if !device.Configured || (device.Service == ServiceCfg{}) {
			continue
		}
		service := device.Service
		defaultFport := uint8(2)
		defaultAck := uint8(0)
		defaultNApps := uint8(1)

		payloadService := []byte{
			defaultFport,
			(service.DataType << 5) | (service.ChannelType << 1) | (service.ChannelParam >> 3 & 0x1),
			(service.ChannelParam << 5) | (service.Acquisition << 2) | (defaultAck << 1) | (uint8(service.Period >> 16)),
			uint8(service.Period >> 8),
			uint8(service.Period),
		}

		controlMessage := append([]byte{defaultNApps}, payloadService...)
		log.Printf("Sending control message %X to device %s", controlMessage, device.DeviceId)
		ttnApplication.ttnClient.Publish(device.DeviceId, controlMessage)

	}
	return nil
}

func (appMgr *ApplicationMgr) IsApplicationNew(newApplication Application) bool {
	isNew := true
	for _, ttnApp := range appMgr.ttnApplications {
		if ttnApp.app.ApplicationId == newApplication.ApplicationId {
			isNew = false
			break
		}
	}
	return isNew
}
