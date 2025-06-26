package learn

import (
	"github.com/IBM/sarama"
	"github.com/stretchr/testify/assert"
	"testing"
)

var addrs = []string{"localhost:9094"}

func TestSyncProducer(t *testing.T) {
	cfg := sarama.NewConfig()
	cfg.Producer.Return.Successes = true
	cfg.Producer.Partitioner = sarama.NewRandomPartitioner
	producer, err := sarama.NewSyncProducer(addrs, cfg)
	assert.NoError(t, err)
	message, i, err := producer.SendMessage(&sarama.ProducerMessage{
		Topic: "test_topic",
		Value: sarama.StringEncoder("testvalue1"),
		Key:   sarama.StringEncoder("testkey1"),
	})
	if err != nil {
		return
	}
	t.Log(message)
	t.Log(i)
}

//func TestAsyncProducer(t *testing.T) {
//	cfg := sarama.NewConfig()
//	cfg.Producer.Return.Errors = true
//	cfg.Producer.Return.Successes = true
//	producer, err := sarama.NewAsyncProducer(addrs, cfg)
//	require.NoError(t, err)
//	mgsCh := producer.Input()
//	go func() {
//		for {
//			msg := &sarama.ProducerMessage{
//				Topic: "test_lcoal",
//			}
//
//		}
//	}()
//}
