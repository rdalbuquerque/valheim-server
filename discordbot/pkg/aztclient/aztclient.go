package aztclient

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"reflect"
	"time"

	"godin/pkg/godinerrors"
	"godin/pkg/statestorageinterface"
	"godin/pkg/utils"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/data/aztables"
	"github.com/iancoleman/strcase"
)

type WriteableEntity struct {
	Entity aztables.Entity
	Value  interface{}
}

type TableClientInterface interface {
	Read(...string) (map[string]interface{}, error)
	Write(statestorageinterface.StateAttributes) error
}

type TableClient struct {
	partitionKey string
	rowKey       string
	etag         *azcore.ETag
	client       *aztables.Client
}

// NewTableClient creates a new instance of TableClient for the specified queue
func NewTableClient(tableName, partitionKey, rowKey string) (TableClientInterface, error) {
	// Get the connection string from environment variables
	connString := os.Getenv("AzureWebJobsStorage")
	if connString == "" {
		return nil, fmt.Errorf("error fetchin 'AzureWebJobsStorage' environment variable")
	}

	// Create a ServiceClient
	serviceClient, err := aztables.NewServiceClientFromConnectionString(connString, nil)
	if err != nil {
		log.Printf("failed to create service client: %v", err)
		return nil, fmt.Errorf("failed to create service client: %v", err)
	}

	// Create a QueueClient for the specified queue
	tableClient := serviceClient.NewClient(tableName)

	return &TableClient{
		client:       tableClient,
		partitionKey: partitionKey,
		rowKey:       rowKey,
	}, nil
}

func (tc *TableClient) createIfResourceNotFound(err error) (aztables.GetEntityResponse, error) {
	var responseError *azcore.ResponseError
	if errors.As(err, &responseError) {
		if responseError.ErrorCode == string(aztables.ResourceNotFound) {
			return tc.create()
		}
		return aztables.GetEntityResponse{}, err
	}
	return aztables.GetEntityResponse{}, err
}

// Reads table into a struct
func (tc *TableClient) Read(columns ...string) (map[string]interface{}, error) {
	entity, err := tc.client.GetEntity(context.TODO(), tc.partitionKey, tc.rowKey, nil)
	if err != nil {
		entity, err = tc.createIfResourceNotFound(err)
		if err != nil {
			return nil, err
		}
	}
	etag := entity.ETag
	tc.etag = &etag
	log.Printf("Read entity: %s", string(entity.Value))

	state := make(map[string]interface{})
	err = json.Unmarshal(entity.Value, &state)
	if err != nil {
		return nil, err
	}
	if err := utils.ValidateColumns(state, columns); err != nil {
		return nil, godinerrors.ReadError{
			Code:    godinerrors.MissingColumnError,
			Message: err.Error(),
		}
	}
	return state, nil
}

func (tc *TableClient) genEntity(state statestorageinterface.StateAttributes) aztables.EDMEntity {
	timestamp := aztables.EDMDateTime(time.Now())
	entity := aztables.Entity{
		PartitionKey: tc.partitionKey,
		RowKey:       tc.rowKey,
		Timestamp:    timestamp,
	}

	properties := make(map[string]interface{})
	value := reflect.ValueOf(state)
	typeOfState := value.Type()
	for i := 0; i < value.NumField(); i++ {
		key := strcase.ToSnake(typeOfState.Field(i).Name)
		value := value.Field(i).Interface()
		properties[key] = value.(string)
	}
	return aztables.EDMEntity{
		Entity:     entity,
		Properties: properties,
	}
}

// Write to table
func (tc *TableClient) Write(state statestorageinterface.StateAttributes) error {
	entity := tc.genEntity(state)
	updateOpts := aztables.UpdateEntityOptions{
		IfMatch:    tc.etag,
		UpdateMode: aztables.UpdateModeReplace,
	}
	entityBytes, err := entity.MarshalJSON()
	if err != nil {
		return err
	}
	log.Printf("Updating entity with: %s", string(entityBytes))
	updatedEntity, err := tc.client.UpdateEntity(context.TODO(), entityBytes, &updateOpts)
	tc.etag = &updatedEntity.ETag
	if err != nil {
		return err
	}
	return nil
}

// Create entity
func (tc *TableClient) create() (aztables.GetEntityResponse, error) {
	timestamp := aztables.EDMDateTime(time.Now())
	entityRequired := aztables.Entity{
		PartitionKey: tc.partitionKey,
		RowKey:       tc.rowKey,
		Timestamp:    timestamp,
	}
	entity := aztables.EDMEntity{
		Entity: entityRequired,
	}
	entityBytes, err := entity.MarshalJSON()
	if err != nil {
		return aztables.GetEntityResponse{}, err
	}
	fmt.Printf("Creating entity with: %s", string(entityBytes))
	addedentity, _ := tc.client.AddEntity(context.TODO(), entityBytes, nil)
	return aztables.GetEntityResponse(addedentity), nil
}
