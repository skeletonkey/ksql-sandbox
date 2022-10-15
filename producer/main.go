package main

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"github.com/Shopify/sarama"
	"io"
	"log"
	"os"
	"strconv"
	"time"
)

// kafka-topics --create --topic explore-main --bootstrap-server localhost:9092 --partitions 3 --replication-factor 1
const (
	dataDir   = "data"
	kafkaConn = "0.0.0.0:9092"
	topic     = "explore-main"
	throttle  = 1 * time.Second
)

type myData struct {
	Station     string `json:"station"`
	Name        string `json:"name"`
	Lat         string `json:"latitude"`
	Long        string `json:"longitude"`
	Elevation   string `json:"elevation"`
	RecordDate  string `json:"record_date"`
	RecordValue int    `json:"record_value"`
}

func main() {
	// create producer
	producer, err := initProducer()
	if err != nil {
		fmt.Println("Error producer: ", err.Error())
		os.Exit(1)
	}

	files, err := os.ReadDir(dataDir)
	checkFatal(err)
	var f *os.File
	defer func() {
		if f != nil {
			err = f.Close()
			if err != nil {
				fmt.Printf("Error while trying to close file: %s", err)
			}
		}
	}()

	for _, file := range files {
		if !file.IsDir() {
			fmt.Println("Name: " + file.Name())
			f, err = os.Open(dataDir + string(os.PathSeparator) + file.Name())
			checkFatal(err)
			r := csv.NewReader(f)
			for {
				record, err := r.Read()
				if err == io.EOF {
					break
				}
				checkFatal(err)
				if record[0] == "STATION" {
					continue
				}
				fmt.Println(record)
				recordValue, err := strconv.Atoi(record[6])
				checkFatal(err)
				data := myData{record[0], record[1], record[2], record[3], record[4],
					record[5], recordValue}
				msg, err := json.Marshal(data)
				checkFatal(err)
				publish(data.Station, string(msg), producer)
				time.Sleep(throttle)
			}
		}
	}
}

func checkFatal(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

func initProducer() (sarama.SyncProducer, error) {
	// setup sarama log to stdout
	sarama.Logger = log.New(os.Stdout, "", log.Ltime)

	// producer config
	config := sarama.NewConfig()
	config.Producer.Retry.Max = 5
	config.Producer.RequiredAcks = sarama.WaitForAll
	config.Producer.Return.Successes = true

	// async producer
	//prd, err := sarama.NewAsyncProducer([]string{kafkaConn}, config)

	// sync producer
	prd, err := sarama.NewSyncProducer([]string{kafkaConn}, config)

	return prd, err
}

func publish(key string, message string, producer sarama.SyncProducer) {
	// publish sync
	msg := &sarama.ProducerMessage{
		Topic: topic,
		Key:   sarama.StringEncoder(key),
		Value: sarama.StringEncoder(message),
	}
	p, o, err := producer.SendMessage(msg)
	if err != nil {
		fmt.Println("Error publish: ", err.Error())
	}

	// publish async
	//producer.Input() <- &sarama.ProducerMessage{

	fmt.Println("Partition: ", p)
	fmt.Println("Offset: ", o)
}
