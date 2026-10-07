package kafka

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/segmentio/kafka-go"
)

type Producer struct {
	writer *kafka.Writer
	log    *slog.Logger
}

func NewProducer(brokers []string, log *slog.Logger) *Producer {
	writer := &kafka.Writer{
		Addr:     kafka.TCP(brokers...),
		Balancer: &kafka.LeastBytes{},
	}

	return &Producer{
		writer: writer,
		log:    log,
	}
}

func (p *Producer) SendMessage(ctx context.Context, topic string, key, value []byte) error {
	const op = "kafka.SendMessage"

	msg := kafka.Message{
		Topic: topic,
		Key:   key,
		Value: value,
	}

	err := p.writer.WriteMessages(ctx, msg)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	p.log.Debug("kafka message sent",
		slog.String("topic", topic),
		slog.String("key", string(key)),
	)

	return nil
}

func (p *Producer) Close() error {
	if err := p.writer.Close(); err != nil {
		p.log.Error("failed to close kafka producer", slog.String("error", err.Error()))
		return err
	}
	p.log.Info("kafka producer closed gracefully")
	return nil
}
