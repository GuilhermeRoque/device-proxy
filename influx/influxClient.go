package influx

import (
	"context"
	"log"
	"time"

	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
)

type SensorData struct {
	Measurement string
	Unit        string
	Value       interface{}
	Timestamp   time.Time
}

type InfluxClient struct {
	client influxdb2.Client
	org    string
	bucket string
}

func NewInfluxClient(baseURL string, token string, organization string, bucket string) *InfluxClient {
	client := &InfluxClient{
		client: influxdb2.NewClient(
			baseURL,
			token,
		),
		org:    organization,
		bucket: bucket}

	return client
}

func (influxClient *InfluxClient) WriteData(data SensorData) {
	// Use blocking write client for writes to desired bucket
	writeAPI := influxClient.client.WriteAPIBlocking(influxClient.org, influxClient.bucket)
	// Create point using fluent style
	p := influxdb2.NewPointWithMeasurement(data.Measurement).
		AddTag("device", data.Unit).
		AddField("value", data.Value).
		SetTime(data.Timestamp)

	log.Printf("Writing point: %+v into server %s bucket %s org %s", data, influxClient.client.ServerURL(), influxClient.bucket, influxClient.org)
	err := writeAPI.WritePoint(context.Background(), p)
	if err != nil {
		log.Println(err)
	}
}

func (influxClient *InfluxClient) Close() {
	influxClient.client.Close()
}
