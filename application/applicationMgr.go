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
	Name      string     `json:"name"`
	DeviceId  string     `json:"devId"`
	DeviceEUI string     `json:"devEUI"`
	Service   ServiceCfg `json:"serviceProfile"`
}

type ServiceCfg struct {
	Name         string  `json:"name"`
	DataType     float32 `json:"dataType"`
	ChannelType  float32 `json:"channelType"`
	ChannelParam float32 `json:"channelParam"`
	Acquisition  float32 `json:"acquisition"`
	Period       float32 `json:"period"`
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
