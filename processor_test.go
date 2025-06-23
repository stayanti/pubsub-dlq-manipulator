package main

import (
	"context"
	"encoding/json"
	"testing"

	"cloud.google.com/go/pubsub"
)

type mockPublishResult struct{}

func (r *mockPublishResult) Get(ctx context.Context) (string, error) {
	return "mock-id", nil
}

type mockTopic struct {
	publishedMsg *pubsub.Message
}

func (t *mockTopic) Publish(ctx context.Context, msg *pubsub.Message) (string, error) {
	t.publishedMsg = msg
	return "mock-id", nil
}

func TestProcessMessage(t *testing.T) {
	// Arrange
	inputMsg := &pubsub.Message{
		Data:       []byte(`{"action":"test","path":"profilecms.io/test","name":"test","image_types":1}`),
		Attributes: map[string]string{"attr1": "value1"},
	}
	mockTopic := &mockTopic{}
	processor := &DefaultMessageProcessor{}

	// Act
	err := processor.ProcessMessage(context.Background(), inputMsg, mockTopic)

	// Assert
	if err != nil {
		t.Fatalf("ProcessMessage failed: %v", err)
	}

	// Verify the published message
	var processedData Message
	err = json.Unmarshal(mockTopic.publishedMsg.Data, &processedData)
	if err != nil {
		t.Fatalf("Failed to unmarshal processed message: %v", err)
	}

	// Assert message transformation
	if processedData.Path != "profilecms.com/test" {
		t.Errorf("Expected path to be 'profilecms.com/test', got '%s'", processedData.Path)
	}

	// Assert attributes preservation
	if mockTopic.publishedMsg.Attributes["attr1"] != "value1" {
		t.Error("Attributes were not preserved")
	}
}
