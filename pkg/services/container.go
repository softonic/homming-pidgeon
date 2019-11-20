package services

import (
	"github.com/softonic/homing-pigeon/pkg/messages"
	"github.com/softonic/homing-pigeon/pkg/readers"
	amqpAdapter "github.com/softonic/homing-pigeon/pkg/readers/adapters/amqp"
	"github.com/softonic/homing-pigeon/pkg/writers"
	"github.com/sarulabs/dingo"
	writeAdapters "github.com/softonic/homing-pigeon/pkg/writers/adapters"
	readAdapters "github.com/softonic/homing-pigeon/pkg/readers/adapters"
	"os"
	"strconv"
)

var Container = []dingo.Def{
	{
		Name: "Reader",
		Build: &readers.Reader{},
		Params: dingo.Params{
			"WriteChannel": dingo.Service("WriteChannel"),
			"AckChannel": dingo.Service("AckChannel"),
			"ReadAdapter": dingo.Service("AmqpAdapter"),
		},
	},
	{
		Name: "DummyAdapter",
		Build: func() (readAdapters.ReadAdapter, error) {
			return &readAdapters.Dummy{}, nil
		},
	},
	{
		Name: "AmqpAdapter",
		Build: func(config amqpAdapter.Config) (readAdapters.ReadAdapter, error) {
			return &readAdapters.Amqp{
				Config: config,
			}, nil
		},
		Params: dingo.Params{
			"0": dingo.Service("AmqpConfig"),
		},
	},
	{
		Name: "AmqpConfig",
		Build: func() (amqpAdapter.Config, error) {
			qosPrefetchCount, err := strconv.Atoi(os.Getenv("RABBITMQ_QOS_PREFETCH_COUNT"))
			if err != nil {
				qosPrefetchCount = 0
			}
			return amqpAdapter.Config{
				Url: os.Getenv("RABBITMQ_URL"),
				DeadLettersExchangeName: os.Getenv("RABBITMQ_DLX_NAME"),
				DeadLettersQueueName: os.Getenv("RABBITMQ_DLX_QUEUE_NAME"),
				ExchangeName: os.Getenv("RABBITMQ_EXCHANGE_NAME"),
				QueueName: os.Getenv("RABBITMQ_QUEUE_NAME"),
				QosPrefetchCount: qosPrefetchCount,
			}, nil
		},
	},
	{
		Name: "Writer",
		Build: &writers.Writer{},
		Params: dingo.Params{
			"WriteChannel": dingo.Service("WriteChannel"),
			"AckChannel": dingo.Service("AckChannel"),
			"WriteAdapter": dingo.Service("ElasticsearchAdapter"),
		},
	},

	{
		Name: "ElasticsearchAdapter",
		Build: func() (writeAdapters.WriteAdapter, error) {
			return &writeAdapters.Elasticsearch{}, nil
		},
	},
	{
		Name: "NopAdapter",
		Build: func() (writeAdapters.WriteAdapter, error) {
			return &writeAdapters.Nop{}, nil
		},
	},
	{
		Name: "WriteChannel",
		Build: func() (*chan messages.Message, error) {
			c := make(chan messages.Message)
			return &c, nil
		},
	},
	{
		Name: "AckChannel",
		Build: func() (*chan messages.Ack, error) {
			c := make(chan messages.Ack)
			return &c, nil
		},
	},
}