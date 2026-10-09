package kafka

import "github.com/segmentio/kafka-go"

func EnsureTopicExists(broker, topic string) error {
	conn, err := kafka.Dial("tcp", broker)
	if err != nil {
		return err
	}
	defer conn.Close()

	topicConfigs := []kafka.TopicConfig{
		{
			Topic:             topic,
			NumPartitions:     1,
			ReplicationFactor: 1,
		},
	}

	return conn.CreateTopics(topicConfigs...)
}
