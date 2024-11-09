package azqclient

import (
	"context"
	"encoding/base64"
	"fmt"
	"os"

	"github.com/Azure/azure-sdk-for-go/sdk/storage/azqueue"
)

type QueueClient struct {
	client *azqueue.QueueClient
}

// NewQueueClient creates a new instance of QueueClient for the specified queue
func NewQueueClient(queueName string) (*QueueClient, error) {
	// Get the connection string from environment variables
	connString := os.Getenv("AzureWebJobsStorage")
	if connString == "" {
		return nil, fmt.Errorf("AzureWebJobsStorage environment variable is not set")
	}

	// Create a ServiceClient
	serviceClient, err := azqueue.NewServiceClientFromConnectionString(connString, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create service client: %w", err)
	}

	// Create a QueueClient for the specified queue
	queueClient := serviceClient.NewQueueClient(queueName)

	return &QueueClient{
		client: queueClient,
	}, nil
}

// EnqueueMessage adds a message to the queue
func (qc *QueueClient) EnqueueMessage(message string) error {
	b64message := base64.StdEncoding.EncodeToString([]byte(message))
	_, err := qc.client.EnqueueMessage(context.Background(), b64message, nil)
	if err != nil {
		return fmt.Errorf("failed to enqueue message: %w", err)
	}
	return nil
}
