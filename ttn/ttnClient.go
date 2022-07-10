package ttn

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	mqtt "github.com/eclipse/paho.mqtt.golang"
	"log"
	"lorawanMgnt/influx"
	"os"
)

var baseUrlInflux = os.Getenv("BASE_URL_INFLUX")
var tokenInflux = os.Getenv("TOKEN_INFLUX")
var bucketInflux = os.Getenv("BUCKET_INFLUX")
var organizationInflux = os.Getenv("ORGANIZATION_INFLUX")

type SensorUplink struct {
	End_device_ids struct {
		Device_id       string
		Application_ids struct {
			Application_id string
		}
	}
	Uplink_message struct {
		Frm_payload string
	}
}

var influxClient = influx.NewInfluxClient(baseUrlInflux, tokenInflux, organizationInflux, bucketInflux)

var messagePubHandler mqtt.MessageHandler = func(client mqtt.Client, msg mqtt.Message) {
	fmt.Printf("Received message: %s from topic: %s\n", msg.Payload(), msg.Topic())
	var data SensorUplink
	_ = json.Unmarshal(msg.Payload(), &data)
	endDeviceId := data.End_device_ids.Device_id
	appId := data.End_device_ids.Application_ids.Application_id
	value := data.Uplink_message.Frm_payload
	decodedString, _ := base64.StdEncoding.DecodeString(value)
	sensorData := influx.SensorData{
		Measurement: appId,
		Unit:        endDeviceId,
		Value:       decodedString[0],
	}

	log.Println(fmt.Sprintf("deviceId %s appId %s value %d", endDeviceId, appId, value))
	influxClient.WriteData(sensorData)
}

var connectHandler mqtt.OnConnectHandler = func(client mqtt.Client) {
	fmt.Println("Connected")
}

var connectLostHandler mqtt.ConnectionLostHandler = func(client mqtt.Client, err error) {
	fmt.Printf("Connect lost: %v", err)
}

type TtnClient struct {
	client          mqtt.Client
	applicationName string
}

func (ttnClient *TtnClient) Connect() {
	if token := ttnClient.client.Connect(); token.Wait() && token.Error() != nil {
		panic(token.Error())
	}
	ttnClient.subscribe()
}

func (ttnClient *TtnClient) subscribe() {
	username := fmt.Sprintf("%s@ttn", ttnClient.applicationName)
	topicDevicesUp := fmt.Sprintf("v3/%s/devices/#", username)
	token := ttnClient.client.Subscribe(topicDevicesUp, 1, nil)
	token.Wait()
	fmt.Printf("Subscribed to topic: %s", topicDevicesUp)
}

func (ttnClient *TtnClient) Close() {
	ttnClient.client.Disconnect(250)
}

func NewTtnClient(broker string, port uint32, applicationName string, password string) *TtnClient {
	opts := mqtt.NewClientOptions()
	opts.AddBroker(fmt.Sprintf("tcp://%s:%d", broker, port))
	username := fmt.Sprintf("%s@ttn", applicationName)
	opts.SetUsername(username)
	opts.SetPassword(password)
	opts.SetDefaultPublishHandler(messagePubHandler)
	opts.OnConnect = connectHandler
	opts.OnConnectionLost = connectLostHandler
	client := mqtt.NewClient(opts)
	ttnClient := &TtnClient{
		client:          client,
		applicationName: applicationName,
	}
	return ttnClient
}
