package application

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

type TtnApplication struct {
	client mqtt.Client
	app    *Application
}

func (ttnApplication *TtnApplication) sendDeviceCfgById(deviceId string) {
	for _, device := range ttnApplication.app.Devices {
		if device.DeviceId == deviceId {
			ttnApplication.sendDeviceCfg(device)
		}
	}
}
func (ttnApplication *TtnApplication) sendDeviceCfg(device Device) {
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
	ttnApplication.Publish(device.DeviceId, controlMessage)
}

func (ttnApplication *TtnApplication) Connect() {
	if token := ttnApplication.client.Connect(); token.Wait() && token.Error() != nil {
		panic(token.Error())
	}
	ttnApplication.subscribe()
}

func (ttnApplication *TtnApplication) messageJoinHandler(client mqtt.Client, msg mqtt.Message) {
	log.Printf("Received message: %s from topic: %s\n", msg.Payload(), msg.Topic())
	var data DeviceJoin
	_ = json.Unmarshal(msg.Payload(), &data)
	endDeviceId := data.End_device_ids.Device_id
	ttnApplication.sendDeviceCfgById(endDeviceId)
}

func (ttnApplication *TtnApplication) messagePubHandler(client mqtt.Client, msg mqtt.Message) {
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
	var influxClient = influx.NewInfluxClient(baseUrlInflux, ttnApplication.app.Token, ttnApplication.app.Organization, ttnApplication.app.Bucket)
	influxClient.WriteData(sensorData)
}

func (ttnApplication *TtnApplication) Publish(deviceId string, payload []byte) {
	username := fmt.Sprintf("%s@ttn", ttnApplication.app.ApplicationId)
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
	ttnApplication.client.Publish(
		topic,
		1,
		false,
		jsonMsg,
	)
}

func (ttnApplication *TtnApplication) subscribe() {
	username := fmt.Sprintf("%s@ttn", ttnApplication.app.ApplicationId)
	topicDevicesUp := fmt.Sprintf("v3/%s/devices/+/up", username)
	token := ttnApplication.client.Subscribe(topicDevicesUp, 1, ttnApplication.messagePubHandler)
	token.Wait()

	topicDevicesJoin := fmt.Sprintf("v3/%s/devices/+/join", username)
	tokenJoin := ttnApplication.client.Subscribe(topicDevicesJoin, 1, ttnApplication.messageJoinHandler)
	tokenJoin.Wait()
	log.Printf("Subscribed to topic: %s", tokenJoin)
}

func (ttnApplication *TtnApplication) Close() {
	ttnApplication.client.Disconnect(250)
}

func NewTtnApp(broker string, port uint32, application *Application) *TtnApplication {
	opts := mqtt.NewClientOptions()
	opts.AddBroker(fmt.Sprintf("tcp://%s:%d", broker, port))
	username := fmt.Sprintf("%s@ttn", application.ApplicationId)
	opts.SetUsername(username)
	opts.SetPassword(application.APIKey)
	opts.OnConnect = connectHandler
	opts.OnConnectionLost = connectLostHandler
	client := mqtt.NewClient(opts)
	ttnApplication := &TtnApplication{
		client: client,
		app:    application,
	}
	return ttnApplication
}

var connectHandler mqtt.OnConnectHandler = func(client mqtt.Client) {
	log.Println("Connected")
}

var connectLostHandler mqtt.ConnectionLostHandler = func(client mqtt.Client, err error) {
	log.Printf("Connect lost: %v", err)
}
