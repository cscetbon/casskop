package cassandrabackup

import (
	"cmp"
	"fmt"
	"net/http"
	"path"

	api "github.com/cscetbon/casskop/api/v2"
	icarus "github.com/instaclustr/instaclustr-icarus-go-client/pkg/instaclustr_icarus"
	corev1 "k8s.io/api/core/v1"
	controllerclient "sigs.k8s.io/controller-runtime/pkg/client"
)

// DefaultCassandraDirectory is the directory holding the Cassandra data folder in the pods
const DefaultCassandraDirectory = "/var/lib/cassandra"

// BackupOperationRequest is the request of the generated Icarus client plus the fields added by
// Icarus 2.0 that it does not know about
type BackupOperationRequest struct {
	icarus.BackupOperationRequest
	// DataDirs lists the Cassandra data directories, mandatory since Icarus 2.0
	DataDirs []string `json:"dataDirs,omitempty"`
}

// RestoreOperationRequest is the request of the generated Icarus client plus the fields added by
// Icarus 2.0 that it does not know about
type RestoreOperationRequest struct {
	icarus.RestoreOperationRequest
	// DataDirs lists the Cassandra data directories, mandatory since Icarus 2.0
	DataDirs []string `json:"dataDirs,omitempty"`
}

// DataDirs returns the data directories of a node from the directory holding its data folder,
// DefaultCassandraDirectory when empty
func DataDirs(cassandraDirectory string) []string {
	return []string{path.Join(cmp.Or(cassandraDirectory, DefaultCassandraDirectory), "data")}
}

type Client interface {
	PerformRestoreOperation(restoreOperation RestoreOperationRequest) (*icarus.RestoreOperationResponse, error)
	RestoreOperationByID(operationId string) (*icarus.RestoreOperationResponse, error)
	PerformBackupOperation(request BackupOperationRequest) (*icarus.BackupOperationResponse, error)
	BackupOperationByID(id string) (response *icarus.BackupOperationResponse, err error)
	Build() error
}

type client struct {
	Client
	config    *Config
	podClient *icarus.APIClient

	newClient func(*icarus.Configuration) *icarus.APIClient
}

func New(config *Config) Client {
	return &client{config: config, newClient: icarus.NewAPIClient}
}

func (cs *client) Build() error {
	cs.podClient = cs.newClient( cs.cassandraBackupSidecarConfig())
	return nil
}

func ClientFromCluster(k8sClient controllerclient.Client, cluster *api.CassandraCluster,
	pod *corev1.Pod) (Client, error) {
	config := ClusterConfig(k8sClient, cluster, pod)

	client := New(config)
	err := client.Build()
	if err != nil {
		return nil, err
	}

	return client, nil
}

func (cs *client) cassandraBackupSidecarConfig() (config *icarus.Configuration) {
	config = icarus.NewConfiguration()

	protocol := "http"

	if cs.config.UseSSL {
		config.Scheme = "HTTPS"
		transport := &http.Transport{TLSClientConfig: cs.config.TLSConfig}
		config.HTTPClient = &http.Client{Transport: transport, Timeout: cs.config.Timeout}
		protocol = "https"
	}

	config.BasePath = fmt.Sprintf("%s://%s:%d", protocol,cs.config.Host, cs.config.Port)
	//config.Host = cs.config.Host

	return config
}