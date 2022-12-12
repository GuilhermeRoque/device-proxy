package ttn

import (
	"encoding/base64"
	"encoding/binary"
	"encoding/json"
	"fmt"
	mqtt "github.com/eclipse/paho.mqtt.golang"
	"log"
	"lorawanMgnt/influx"
	"math"
	"os"
	"time"
)

type DeviceUplink struct {
	End_device_ids struct {
		Device_id       string
		Application_ids struct {
			Application_id string
		}
	}
	Received_at    string
	Uplink_message struct {
		Frm_payload string
	}
}

var messagePubHandler mqtt.MessageHandler = func(client mqtt.Client, msg mqtt.Message) {
	log.Printf("Received message: %s from topic: %s\n", msg.Payload(), msg.Topic())
	var data DeviceUplink
	_ = json.Unmarshal(msg.Payload(), &data)
	endDeviceId := data.End_device_ids.Device_id
	appId := data.End_device_ids.Application_ids.Application_id
	receivedAt := data.Received_at
	value := data.Uplink_message.Frm_payload
	decodedString, _ := base64.StdEncoding.DecodeString(value)
	bits := binary.BigEndian.Uint32(decodedString)
	valueFloat := math.Float32frombits(bits)
	timestamp, _ := time.Parse(time.RFC3339, receivedAt)
	log.Printf("deviceId %s appId %s raw value %s decoded string %X bits %X float %f timestamp %s", endDeviceId, appId, value, decodedString, bits, valueFloat, timestamp)
	var sensorData = influx.SensorData{
		Measurement: appId,
		Unit:        endDeviceId,
		Value:       valueFloat,
		Timestamp:   timestamp,
	}
	var baseUrlInflux = os.Getenv("BASE_URL_INFLUX")
	var tokenInflux = os.Getenv("TOKEN_INFLUX")
	var bucketInflux = os.Getenv("BUCKET_INFLUX")
	var organizationInflux = os.Getenv("ORGANIZATION_INFLUX")
	var influxClient = influx.NewInfluxClient(baseUrlInflux, tokenInflux, organizationInflux, bucketInflux)
	influxClient.WriteData(sensorData)
}

var connectHandler mqtt.OnConnectHandler = func(client mqtt.Client) {
	log.Println("Connected")
}

var connectLostHandler mqtt.ConnectionLostHandler = func(client mqtt.Client, err error) {
	log.Printf("Connect lost: %v", err)
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

type DownlinkPayload struct {
	Frm_payload string `json:"frm_payload"`
	F_port      uint8  `json:"f_port"`
	Priority    string `json:"priority"`
}

type DownlinkMsg struct {
	Downlinks []DownlinkPayload `json:"downlinks"`
}

func (ttnClient *TtnClient) Publish(deviceId string, payload []byte) {
	username := fmt.Sprintf("%s@ttn", ttnClient.applicationName)
	topic := fmt.Sprintf("v3/%s/devices/%s/down/push", username, deviceId)
	downlinkPayload := DownlinkPayload{
		Frm_payload: base64.StdEncoding.EncodeToString(payload),
		F_port:      1,
		Priority:    "NORMAL",
	}
	payloads := []DownlinkPayload{downlinkPayload}
	msg := DownlinkMsg{Downlinks: payloads}
	jsonMsg, _ := json.Marshal(msg)
	log.Printf("Sending %s to %s\n", jsonMsg, topic)
	ttnClient.client.Publish(
		topic,
		1,
		false,
		jsonMsg,
	)
}

func (ttnClient *TtnClient) subscribe() {
	username := fmt.Sprintf("%s@ttn", ttnClient.applicationName)
	topicDevicesUp := fmt.Sprintf("v3/%s/devices/up", username)
	token := ttnClient.client.Subscribe(topicDevicesUp, 1, nil)
	token.Wait()
	log.Printf("Subscribed to topic: %s", topicDevicesUp)
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
