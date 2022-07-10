package influx

import (
	"context"
	"fmt"
	"log"
	"time"

	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
)

type SensorData struct {
	Measurement string
	Unit        string
	Value       interface{}
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
		SetTime(time.Now())
	err := writeAPI.WritePoint(context.Background(), p)
	if err != nil {
		log.Println(err)
	}
}

func (influxClient *InfluxClient) queryData(measurement string) {
	// Get query client
	queryAPI := influxClient.client.QueryAPI(influxClient.org)

	// Get parser flux query result
	query := fmt.Sprintf(
		`from(bucket:"%s")|> range(start: -1h) |> filter(fn: (r) => r._measurement == "%s")`,
		influxClient.bucket,
		measurement)

	result, err := queryAPI.Query(context.Background(), query)
	if err == nil {
		// Use Next() to iterate over query result lines
		for result.Next() {
			// Observe when there is new grouping key producing new table
			if result.TableChanged() {
				fmt.Printf("table: %s\n", result.TableMetadata().String())
			}
			// read result
			values := result.Record().Values()
			fmt.Printf("%v\n", values)
		}
		if result.Err() != nil {
			fmt.Printf("Query error: %s\n", result.Err().Error())
		}
	} else {
		log.Println(err)
	}
}

func (influxClient *InfluxClient) Close() {
	influxClient.client.Close()
}
