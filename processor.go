package main

import (
	"context"
	"encoding/json"
	"strings"
	"time"

	"cloud.google.com/go/pubsub"
)

type DefaultMessageProcessor struct{}

func (p *DefaultMessageProcessor) ProcessMessage(ctx context.Context, msg *pubsub.Message, destinationTopic TopicPublisher) error {
	publishCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	var messageData Message
	if err := json.Unmarshal(msg.Data, &messageData); err != nil {
		return err
	}

	// Create your custom logic here
    // messageData.Path = strings.Replace(messageData.Path, "profilecms.io", "profilecms.com", 1)

	newData, err := json.Marshal(messageData)
	if err != nil {
		return err
	}

	_, err = destinationTopic.Publish(publishCtx, &pubsub.Message{
		Data:       newData,
		Attributes: msg.Attributes,
	})
	return err
}
