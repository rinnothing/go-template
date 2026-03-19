package queue

import (
	"bytes"
	"context"
	"encoding/json"

	"github.com/rinnothing/avito-pr/internal/model"
	"github.com/rinnothing/avito-pr/internal/usecase/async"
	"github.com/segmentio/kafka-go"
)

var _ async.Queue = &kafkaQueue{}

// ExtractTask implements [async.Queue].
func (k *kafkaQueue) ExtractTask(ctx context.Context) (*model.KeyVal, error) {
	msg, err := k.read.ReadMessage(ctx)
	if err != nil {
		return nil, err
	}

	dec := json.NewDecoder(bytes.NewReader(msg.Value))
	dec.DisallowUnknownFields()

	var keyval model.KeyVal
	err = dec.Decode(&keyval)
	if err != nil {
		return nil, err
	}

	return &keyval, nil
}

// PutTask implements [async.Queue].
func (k *kafkaQueue) PutTask(ctx context.Context, keyval *model.KeyVal) error {
	body, err := json.Marshal(keyval)
	if err != nil {
		return err
	}

	err = k.write.WriteMessages(ctx, kafka.Message{Value: body})
	if err != nil {
		return err
	}

	return nil
}
