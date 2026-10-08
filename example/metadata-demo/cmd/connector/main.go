package main

import (
	"context"
	"encoding/json"
	"log/slog"
	"os"
	"strconv"
	"time"

	cdc "github.com/Trendyol/go-pq-cdc-kafka"
	"github.com/Trendyol/go-pq-cdc-kafka/config"
	cdcconfig "github.com/Trendyol/go-pq-cdc/config"
	"github.com/Trendyol/go-pq-cdc/pq/publication"
	"github.com/Trendyol/go-pq-cdc/pq/slot"
	gokafka "github.com/segmentio/kafka-go"
)

const topic = "cdc.demo"

func main() {
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, nil)))
	ctx := context.Background()

	cfg := config.Connector{
		CDC: cdcconfig.Config{
			Host:     env("PG_HOST", "127.0.0.1"),
			Username: "cdc_user",
			Password: "cdc_pass",
			Database: "cdc_db",
			Publication: publication.Config{
				CreateIfNotExists: true,
				Name:              "cdc_publication_demo",
				Operations: publication.Operations{
					publication.OperationInsert,
					publication.OperationUpdate,
					publication.OperationDelete,
				},
				Tables: publication.Tables{
					{Name: "parents", ReplicaIdentity: publication.ReplicaIdentityFull},
					{Name: "children", ReplicaIdentity: publication.ReplicaIdentityFull},
				},
			},
			Slot: slot.Config{
				CreateIfNotExists:           true,
				Name:                        "cdc_slot_demo",
				SlotActivityCheckerInterval: 3000,
			},
			Metric: cdcconfig.MetricConfig{Port: 8081},
			Logger: cdcconfig.LoggerConfig{LogLevel: slog.LevelInfo},
		},
		Kafka: config.Kafka{
			TableTopicMapping: map[string]string{
				"public.parents":  topic,
				"public.children": topic,
			},
			Brokers:                     []string{env("KAFKA_BROKER", "localhost:19092")},
			AllowAutoTopicCreation:      true,
			ProducerBatchTickerDuration: 200 * time.Millisecond,
			ProducerBatchSize:           50,
		},
	}

	connector, err := cdc.NewConnector(ctx, cfg, handler)
	if err != nil {
		slog.Error("new connector", "error", err)
		os.Exit(1)
	}
	defer connector.Close()

	slog.Info("cdc connector started", "topic", topic)
	connector.Start(ctx)
}

func handler(msg *cdc.Message) []gokafka.Message {
	data := msg.NewData
	if msg.Type.IsDelete() {
		data = msg.OldData
	}
	if data == nil {
		return nil
	}

	body := map[string]any{
		"table":         msg.TableNamespace + "." + msg.TableName,
		"operation":     string(msg.Type),
		"transactionId": msg.TransactionID,
		"lsn":           uint64(msg.LSN),
		"lsnText":       msg.LSN.String(),
		"data":          data,
	}
	value, err := json.Marshal(body)
	if err != nil {
		slog.Error("marshal", "error", err)
		return nil
	}

	txid := strconv.FormatUint(uint64(msg.TransactionID), 10)
	lsn := strconv.FormatUint(uint64(msg.LSN), 10)

	slog.Info("publishing",
		"table", msg.TableName,
		"op", msg.Type,
		"txid", msg.TransactionID,
		"lsn", lsn,
	)

	return []gokafka.Message{{
		Topic: topic,
		Key:   []byte(txid),
		Value: value,
		Headers: []gokafka.Header{
			{Key: "txid", Value: []byte(txid)},
			{Key: "lsn", Value: []byte(lsn)},
			{Key: "table", Value: []byte(msg.TableNamespace + "." + msg.TableName)},
			{Key: "op", Value: []byte(string(msg.Type))},
		},
	}}
}

func env(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
