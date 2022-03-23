package cassandrabackup

import (
	"context"
	"fmt"
	"net/http"

	"github.com/antihax/optional"
	icarus "github.com/instaclustr/instaclustr-icarus-go-client/pkg/instaclustr_icarus"
	"github.com/mitchellh/mapstructure"
)

func performResoreOperation(body interface{}, _ *http.Response, err error) (
	response *icarus.RestoreOperationResponse, errout error) {
	if err != nil {
		return nil, err
	}

	errout = mapstructure.Decode(body, &response)
	return
}

func (client *client) PerformRestoreOperation(restoreOperationReq icarus.RestoreOperationRequest) (
	*icarus.RestoreOperationResponse, error) {
	podClient := client.podClient
	if podClient == nil {
		return nil, ErrNoCassandraBackupClientAvailable
	}
	return performResoreOperation(podClient.OperationsApi.OperationsPost(context.Background(), &icarus.OperationsApiOperationsPostOpts{
		Body: optional.NewInterface(restoreOperationReq),
	}))
}

func (client *client) RestoreOperationByID(operationId string) (*icarus.RestoreOperationResponse, error) {
	if operationId == "" {
		return nil, fmt.Errorf("must get a non empty id")
	}

	podClient := client.podClient
	if podClient == nil {
		return nil, ErrNoCassandraBackupClientAvailable
	}

	return performResoreOperation(podClient.OperationsApi.OperationsOperationIdGet(context.Background(), operationId))
}

func performBackupOperation(body interface{}, _ *http.Response, err error) (
	response *icarus.BackupOperationResponse, errout error) {
	if err != nil {
		return nil, err
	}

	errout = mapstructure.Decode(body, &response)
	return
}

func (client *client) BackupOperationByID(operationId string) (response *icarus.BackupOperationResponse, err error) {

	if operationId == "" {
		return nil, fmt.Errorf("must get a non empty id")
	}

	podClient := client.podClient
	if podClient == nil {
		return nil, ErrNoCassandraBackupClientAvailable
	}
	return performBackupOperation(podClient.OperationsApi.OperationsOperationIdGet(context.Background(), operationId))
}

func (client *client) PerformBackupOperation(request icarus.BackupOperationRequest) (
	response *icarus.BackupOperationResponse, err error) {
	podClient := client.podClient
	if podClient == nil {
		return nil, ErrNoCassandraBackupClientAvailable
	}

	return performBackupOperation(podClient.OperationsApi.OperationsPost(context.Background(), &icarus.OperationsApiOperationsPostOpts{
		Body: optional.NewInterface(request),
	}))
}
