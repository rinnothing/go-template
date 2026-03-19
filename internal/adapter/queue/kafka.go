package queue

import "github.com/segmentio/kafka-go"

type kafkaQueue struct {
	read  *kafka.Reader
	write *kafka.Writer
}

func New(read *kafka.Reader, write *kafka.Writer) *kafkaQueue {
	return &kafkaQueue{
		read:  read,
		write: write,
	}
}
