package main

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"cloud.google.com/go/pubsub"
)

type processingMode int

const (
	republish processingMode = iota
	acknowledgeOnly
	customProcess
)

type PubSubClient interface {
	Topic(id string) *pubsub.Topic
	Subscription(id string) *pubsub.Subscription
	Close() error
}

type TopicPublisher interface {
	Publish(ctx context.Context, msg *pubsub.Message) (string, error)
}

type MessageProcessor interface {
	ProcessMessage(ctx context.Context, msg *pubsub.Message, topic TopicPublisher) error
}

type config struct {
	sourceSubscription string
	destinationTopic   string
	mode               processingMode
	messageLimit       int
	previewOnly        bool
	processor          MessageProcessor
	parallelProcessing bool
}

type Message struct {
	Action     string `json:"action"`
	Path       string `json:"path"`
	Name       string `json:"name"`
	ImageTypes int    `json:"image_types"`
}

type topicWrapper struct {
	topic *pubsub.Topic
}

func (w *topicWrapper) Publish(ctx context.Context, msg *pubsub.Message) (string, error) {
	result := w.topic.Publish(ctx, msg)
	return result.Get(ctx)
}

func main() {
	cfg := getConfigFromUser()

	ctx := context.Background()
	client, err := createPubSubClient(ctx, cfg.sourceSubscription)
	if err != nil {
		log.Fatalf("Failed to create Pub/Sub client: %v", err)
	}
	defer client.Close()

	destinationClient, destinationTopic := setupDestination(ctx, client, cfg)
	if destinationClient != client {
		defer destinationClient.Close()
	}

	subscription := client.Subscription(extractSubscriptionID(cfg.sourceSubscription))

	if cfg.previewOnly {
		previewMessages(ctx, subscription, cfg.messageLimit)
		return
	}

	processMessages(ctx, subscription, destinationTopic, cfg)
}

func getConfigFromUser() config {
	reader := bufio.NewReader(os.Stdin)
	var cfg config

	fmt.Println("\n=== DLQ Message Processor ===")

	fmt.Print("\nEnter the full subscription path (e.g., projects/my-project/subscriptions/my-sub): ")
	cfg.sourceSubscription, _ = reader.ReadString('\n')
	cfg.sourceSubscription = strings.TrimSpace(cfg.sourceSubscription)

	fmt.Println("\nSelect processing mode:")
	fmt.Println("1. Preview messages only")
	fmt.Println("2. Republish messages to another topic")
	fmt.Println("3. Acknowledge messages without republishing")
	fmt.Println("4. Custom process messages")
	fmt.Print("Enter your choice (1-4): ")

	modeChoice, _ := reader.ReadString('\n')
	modeChoice = strings.TrimSpace(modeChoice)

	switch modeChoice {
	case "1":
		cfg.previewOnly = true
	case "2":
		cfg.mode = republish
	case "3":
		cfg.mode = acknowledgeOnly
	case "4":
		cfg.mode = customProcess
		cfg.processor = &DefaultMessageProcessor{}
	default:
		log.Fatal("Invalid mode selected")
	}

	if !cfg.previewOnly && (cfg.mode == republish || cfg.mode == customProcess) {
		fmt.Print("\nEnter the destination topic path (e.g., projects/my-project/topics/my-topic): ")
		cfg.destinationTopic, _ = reader.ReadString('\n')
		cfg.destinationTopic = strings.TrimSpace(cfg.destinationTopic)
	}

	fmt.Print("\nHow many messages to process? (0 for all): ")
	limitStr, _ := reader.ReadString('\n')
	limitStr = strings.TrimSpace(limitStr)
	cfg.messageLimit, _ = strconv.Atoi(limitStr)

	fmt.Print("\nEnable parallel processing for faster performance? (y/N): ")
	parallel, _ := reader.ReadString('\n')
	parallel = strings.TrimSpace(strings.ToLower(parallel))
	if parallel == "y" || parallel == "yes" {
		cfg.parallelProcessing = true
	}

	fmt.Println("\nConfiguration Summary:")
	fmt.Printf("Source Subscription: %s\n", cfg.sourceSubscription)
	if !cfg.previewOnly && cfg.destinationTopic != "" {
		fmt.Printf("Destination Topic: %s\n", cfg.destinationTopic)
	}
	fmt.Printf("Mode: %v\n", getModeString(cfg))
	fmt.Printf("Message Limit: %d\n", cfg.messageLimit)
	fmt.Printf("Parallel Processing: %v\n", cfg.parallelProcessing)
	fmt.Println("Using default credentials")

	fmt.Print("\nProceed with these settings? (y/N): ")
	confirm, _ := reader.ReadString('\n')
	confirm = strings.TrimSpace(strings.ToLower(confirm))
	if confirm != "y" && confirm != "yes" {
		log.Fatal("Operation cancelled by user")
	}

	return cfg
}

func getModeString(cfg config) string {
	if cfg.previewOnly {
		return "Preview Only"
	}
	switch cfg.mode {
	case republish:
		return "Republish"
	case acknowledgeOnly:
		return "Acknowledge Only"
	case customProcess:
		return "Custom Process"
	default:
		return "Unknown"
	}
}

func createPubSubClient(ctx context.Context, resourcePath string) (*pubsub.Client, error) {
	projectID := extractProjectID(resourcePath)
	return pubsub.NewClient(ctx, projectID)
}

func setupDestination(ctx context.Context, sourceClient *pubsub.Client, cfg config) (*pubsub.Client, TopicPublisher) {
	if cfg.mode == acknowledgeOnly || cfg.previewOnly {
		return sourceClient, nil
	}

	sourceProject := extractProjectID(cfg.sourceSubscription)
	destProject := extractProjectID(cfg.destinationTopic)

	var topic *pubsub.Topic
	if sourceProject == destProject {
		topic = sourceClient.Topic(extractTopicID(cfg.destinationTopic))
		return sourceClient, &topicWrapper{topic: topic}
	}

	destClient, err := createPubSubClient(ctx, cfg.destinationTopic)
	if err != nil {
		log.Fatalf("Failed to create destination Pub/Sub client: %v", err)
	}
	topic = destClient.Topic(extractTopicID(cfg.destinationTopic))
	return destClient, &topicWrapper{topic: topic}
}

func previewMessages(ctx context.Context, subscription *pubsub.Subscription, limit int) {
	if limit <= 0 {
		limit = int(^uint(0) >> 1) // Max int value
	}

	var (
		mu       sync.Mutex
		messages []*pubsub.Message
	)

	cctx, cancel := context.WithCancel(ctx)
	defer cancel()

	err := subscription.Receive(cctx, func(ctx context.Context, msg *pubsub.Message) {
		mu.Lock()
		if len(messages) < limit {
			messages = append(messages, msg)
		}
		currentCount := len(messages)
		mu.Unlock()

		msg.Nack() // Don't acknowledge preview messages

		if currentCount >= limit {
			cancel()
		}
	})

	if err != nil && err != context.Canceled {
		log.Printf("Error previewing messages: %v", err)
		return
	}

	// Print messages in order
	for i, msg := range messages {
		fmt.Printf("\nMessage %d:\nID: %s\nData: %s\nAttributes: %v\n\n",
			i+1, msg.ID, string(msg.Data), msg.Attributes)
	}
}

func processMessages(ctx context.Context, subscription *pubsub.Subscription, destinationTopic TopicPublisher, cfg config) {
	if cfg.messageLimit <= 0 {
		cfg.messageLimit = int(^uint(0) >> 1) // Max int value
	}

	var (
		mu             sync.Mutex
		processedCount int
	)

	cctx, cancel := context.WithCancel(ctx)
	defer cancel()

	// Set max outstanding messages to 1 to process sequentially
	if !cfg.parallelProcessing {
		subscription.ReceiveSettings.MaxOutstandingMessages = 1
		subscription.ReceiveSettings.NumGoroutines = 1
	}

	err := subscription.Receive(cctx, func(ctx context.Context, msg *pubsub.Message) {
		mu.Lock()
		currentCount := processedCount + 1
		if currentCount <= cfg.messageLimit {
			processedCount = currentCount
			mu.Unlock()

			log.Printf("Processing message %d (ID: %s)", currentCount, msg.ID)

			switch cfg.mode {
			case republish:
				republishMessage(ctx, msg, destinationTopic)
			case acknowledgeOnly:
				log.Printf("Acknowledging message without republishing")
				msg.Ack()
			case customProcess:
				err := cfg.processor.ProcessMessage(ctx, msg, destinationTopic)
				if err != nil {
					log.Printf("Failed to process message %s: %v", msg.ID, err)
					msg.Nack()
				} else {
					log.Printf("Successfully processed and published message %s", msg.ID)
					msg.Ack()
				}
			}

			if currentCount >= cfg.messageLimit {
				cancel()
			}
		} else {
			mu.Unlock()
			msg.Nack()
			return
		}
	})

	if err != nil && err != context.Canceled {
		log.Fatalf("Error receiving messages: %v", err)
	}

	log.Printf("Completed processing %d messages", processedCount)
}

func republishMessage(ctx context.Context, msg *pubsub.Message, destinationTopic TopicPublisher) {
	publishCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	id, err := destinationTopic.Publish(publishCtx, &pubsub.Message{
		Data:       msg.Data,
		Attributes: msg.Attributes,
	})
	if err != nil {
		log.Printf("Failed to republish message %s: %v", msg.ID, err)
		msg.Nack()
		return
	}

	log.Printf("Republished message %s to new ID: %s", msg.ID, id)
	msg.Ack()
}

// Helper functions
func extractProjectID(path string) string {
	parts := strings.Split(path, "/")
	if len(parts) >= 2 && parts[0] == "projects" {
		return parts[1]
	}
	return ""
}

func extractSubscriptionID(path string) string {
	parts := strings.Split(path, "/")
	if len(parts) >= 4 {
		return parts[3]
	}
	return ""
}

func extractTopicID(path string) string {
	parts := strings.Split(path, "/")
	if len(parts) >= 4 {
		return parts[3]
	}
	return ""
}
