package kafka

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"time"

	"github.com/segmentio/kafka-go"
)

type EventBroadcaster interface {
	BroadcastToMovie(movieID int64, payload interface{})
}

type Consumer struct {
	reader      *kafka.Reader
	broadcaster EventBroadcaster
	log         *slog.Logger
}

func NewConsumer(brokers []string, groupID, topic string, broadcaster EventBroadcaster, log *slog.Logger) *Consumer {
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:     brokers,
		GroupID:     groupID,
		Topic:       topic,
		MinBytes:    1,
		MaxBytes:    10e6,
		MaxWait:     100 * time.Millisecond,
		StartOffset: kafka.FirstOffset,
	})

	return &Consumer{
		reader:      reader,
		broadcaster: broadcaster,
		log:         log,
	}
}

func (c *Consumer) Start(ctx context.Context) {
	c.log.Info("kafka consumer started listening for comment events")

	go func() {
		for {
			select {
			case <-ctx.Done():
				c.log.Info("kafka consumer stopping...")
				return
			default:
				m, err := c.reader.ReadMessage(ctx)
				if err != nil {
					if errors.Is(err, io.EOF) || errors.Is(err, context.Canceled) {
						return
					}
					c.log.Error("error reading kafka message", slog.String("error", err.Error()))
					continue
				}

				var event struct {
					CommentID int64 `json:"comment_id"`
					MovieID   int64 `json:"movie_id"`
					UserID    int64 `json:"user_id"`
					CreatedAt int64 `json:"created_at"`
				}

				if err := json.Unmarshal(m.Value, &event); err != nil {
					c.log.Error("failed to unmarshal comment event", slog.String("error", err.Error()))
					continue
				}

				c.log.Info("kafka event consumed",
					slog.Int64("movie_id", event.MovieID),
					slog.Int64("comment_id", event.CommentID),
				)

				// Рассылаем событие всем WebSocket-клиентам, смотрящим этот фильм
				c.broadcaster.BroadcastToMovie(event.MovieID, map[string]interface{}{
					"event":      "comment_created",
					"comment_id": event.CommentID,
					"movie_id":   event.MovieID,
					"user_id":    event.UserID,
					"created_at": event.CreatedAt,
				})
			}
		}
	}()
}

func (c *Consumer) Close() error {
	return c.reader.Close()
}
