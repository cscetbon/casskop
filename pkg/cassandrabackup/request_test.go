package cassandrabackup

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDataDirs(t *testing.T) {
	assert.Equal(t, []string{"/var/lib/cassandra/data"}, DataDirs(""))
	assert.Equal(t, []string{"/mnt/cassandra/data"}, DataDirs("/mnt/cassandra/"))
}

// Icarus 2.0 and later refuse a request without dataDirs, the field must reach the JSON body next to
// the fields of the generated client
func TestRequestsCarryDataDirs(t *testing.T) {
	backup, err := json.Marshal(BackupOperationRequest{
		Type_: "backup", StorageLocation: "s3://bucket", SnapshotTag: "tag", DataDirs: DataDirs(""),
	})
	assert.NoError(t, err)
	assert.JSONEq(t, `{"type":"backup","storageLocation":"s3://bucket","snapshotTag":"tag","dataDirs":["/var/lib/cassandra/data"]}`, string(backup))

	restore, err := json.Marshal(RestoreOperationRequest{
		Type_: "restore", StorageLocation: "s3://bucket", SnapshotTag: "tag", RestorationStrategyType: "HARDLINKS",
		DataDirs: DataDirs("/mnt/cassandra"),
	})
	assert.NoError(t, err)
	assert.JSONEq(t, `{"type":"restore","storageLocation":"s3://bucket","snapshotTag":"tag","restorationStrategyType":"HARDLINKS","dataDirs":["/mnt/cassandra/data"]}`, string(restore))
}
