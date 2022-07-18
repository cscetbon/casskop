package cassandrarestore

import (
	"context"
	"fmt"
	"time"

	"emperror.dev/errors"
	v2 "github.com/cscetbon/casskop/api/v2"
	"github.com/cscetbon/casskop/controllers/common"
	"github.com/cscetbon/casskop/pkg/backrest"
	"github.com/cscetbon/casskop/pkg/errorfactory"
	"github.com/cscetbon/casskop/pkg/k8s"
	"github.com/cscetbon/casskop/pkg/util"
	"github.com/go-logr/logr"
	"github.com/sirupsen/logrus"
	v1 "k8s.io/api/core/v1"
	k8sErrors "k8s.io/apimachinery/pkg/api/errors"
	v12 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/record"
	controllerruntime "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

// CassandraClusterReconciler reconciles a CassandraCluster object
type CassandraRestoreReconciler struct {
	// This client, initialized using mgr.Client() above, is a split client
	// that reads objects from the cache and writes to the apiserver
	Recorder record.EventRecorder
	Client   client.Client
	Scheme   *runtime.Scheme
	Log      logr.Logger
}

// +kubebuilder:rbac:groups=db.orange.com,resources=cassandrarestores,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=db.orange.com,resources=cassandrarestores/status,verbs=get;update;patch

func (r CassandraRestoreReconciler) Reconcile(ctx context.Context, request reconcile.Request) (reconcile.Result, error) {

	reqLogger := logrus.WithFields(logrus.Fields{"Request.Namespace": request.Namespace, "Request.Name": request.Name})
	reqLogger.Info("Reconciling CassandraRestore")

	// Fetch the CassandraRestore cassandraRestore
	cassandraRestore := &v2.CassandraRestore{}

	err := r.Client.Get(ctx, request.NamespacedName, cassandraRestore)

	if err != nil {
		if k8sErrors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile request.
			// Owned objects are automatically garbage collected. For additional cleanup logic use finalizers.
			// Return and don't requeue
			return common.Reconciled()
		}
		// Error reading the object - requeue the request.
		return common.RequeueWithError(reqLogger, err.Error(), err)
	}

	// Check the referenced Cluster exists.
	var cassandraCluster *v2.CassandraCluster
	if cassandraCluster, err = k8s.LookupCassandraCluster(ctx, r.Client, cassandraRestore.Spec.CassandraCluster,
		cassandraRestore.Namespace); err != nil {
		// This shouldn't trigger anymore, but leaving it here as a safety belt
		if k8s.IsMarkedForDeletion(cassandraRestore.ObjectMeta) {
			reqLogger.Info("Cluster is gone already, there is nothing we can do")
			return common.Reconciled()
		}
		r.Recorder.Event(
			cassandraRestore,
			v1.EventTypeWarning,
			"CassandraClusterNotFound",
			fmt.Sprintf("Cassandra Cluster %s to restore not found", cassandraRestore.Spec.CassandraCluster))
		return common.RequeueWithError(reqLogger, "failed to lookup referenced cluster", err)
	}

	// Check the referenced Backup exists.
	var cassandraBackup *v2.CassandraBackup
	if cassandraBackup, err = k8s.LookupCassandraBackup(ctx, r.Client, cassandraRestore.Spec.CassandraBackup,
		cassandraRestore.Namespace); err != nil {
		r.Recorder.Event(
			cassandraRestore,
			v1.EventTypeWarning,
			"BackupNotFound",
			fmt.Sprintf("Backup %s to restore not found", cassandraRestore.Spec.CassandraBackup))
		return common.RequeueWithError(reqLogger, "failed to lookup referenced cassandraBackup", err)
	}

	// Require restore
	if len(cassandraRestore.Status.CoordinatorMember) == 0 {
		err = r.requiredRestore(ctx, cassandraRestore, cassandraCluster, cassandraBackup, reqLogger)
		if err != nil {
			switch errors.Cause(err).(type) {
			case errorfactory.ResourceNotReady:
				return controllerruntime.Result{
					RequeueAfter: time.Duration(15) * time.Second,
				}, nil
			default:
				return common.RequeueWithError(reqLogger, err.Error(), err)
			}
		}
		r.Recorder.Event(cassandraRestore,
			v1.EventTypeNormal,
			"RestoreRequired",
			r.restoreEventMessage(cassandraBackup, cassandraRestore.Spec.Datacenter, ""))
		return common.Reconciled()
	}

	restoreConditionType := v2.RestoreConditionType(cassandraRestore.Status.Condition.Type)

	if restoreConditionType.IsRequired() {
		err = r.handleRequiredRestore(ctx, cassandraRestore, cassandraCluster, cassandraBackup, reqLogger)
		if err != nil {
			switch errors.Cause(err).(type) {
			case errorfactory.CassandraBackupSidecarNotReady, errorfactory.ResourceNotReady:
				r.Recorder.Event(
					cassandraRestore,
					v1.EventTypeWarning,
					"PerformRestoreOperationFailed",
					r.restoreEventMessage(cassandraBackup, cassandraRestore.Spec.Datacenter, " failed to run, will retry"))
				return controllerruntime.Result{
					RequeueAfter: time.Duration(15) * time.Second,
				}, nil
			default:
				return common.RequeueWithError(reqLogger, err.Error(), err)
			}
		}
		r.Recorder.Event(cassandraRestore,
			v1.EventTypeNormal,
			"RestoreInitiated",
			r.restoreEventMessage(cassandraBackup, cassandraRestore.Spec.Datacenter, ""))

		return common.Reconciled()
	}

	if restoreConditionType.IsInProgress() {
		err = r.checkRestoreOperationState(ctx, cassandraRestore, cassandraCluster, cassandraBackup, reqLogger)
		if err != nil {
			switch errors.Cause(err).(type) {
			case errorfactory.CassandraBackupSidecarNotReady, errorfactory.ResourceNotReady:
				return controllerruntime.Result{
					RequeueAfter: time.Duration(15) * time.Second,
				}, nil
			case errorfactory.CassandraBackupOperationRunning:
				return controllerruntime.Result{
					RequeueAfter: time.Duration(20) * time.Second,
				}, nil
			case errorfactory.CassandraBackupOperationFailure:
				r.Recorder.Event(cassandraRestore,
					v1.EventTypeNormal,
					"RestoreFailed",
					r.restoreEventMessage(cassandraBackup, cassandraRestore.Spec.Datacenter, err.Error()))
				return common.Reconciled()
			default:
				return common.RequeueWithError(reqLogger, err.Error(), err)
			}
		}
		r.Recorder.Event(cassandraRestore,
			v1.EventTypeNormal,
			"RestoreCompleted",
			r.restoreEventMessage(cassandraBackup, cassandraRestore.Spec.Datacenter, ""))
	}
	return common.Reconciled()
}

func (r CassandraRestoreReconciler) restoreEventMessage(cassandraBackup *v2.CassandraBackup,
	datacenter string, message string) string {
	return fmt.Sprintf("Restore of backup %s of datacenter %s of cluster %s to %s "+
		"under snapshot %s. %s", cassandraBackup.Name,
		datacenter, cassandraBackup.Spec.CassandraCluster, cassandraBackup.Spec.StorageLocation,
		cassandraBackup.Spec.SnapshotTag, message)
}

// requiredRestore select restore coordinator on a specific member of a Cluster
func (r *CassandraRestoreReconciler) requiredRestore(ctx context.Context, restore *v2.CassandraRestore, cc *v2.CassandraCluster,
	backup *v2.CassandraBackup, reqLogger *logrus.Entry) error {
	ns := restore.Namespace

	pods, err := r.listPods(ctx, ns, k8s.LabelsForCassandraDC(cc, backup.Spec.Datacenter))
	if err != nil {
		return errorfactory.New(errorfactory.ResourceNotReady{}, err, "No pods founds for this dc")
	}

	numberOfPods := len(pods.Items)

	if numberOfPods > 0 {
		if err := UpdateRestoreStatus(r.Client, restore,
			v2.BackRestStatus{
				Condition: &v2.BackRestCondition{
					Type:               string(v2.RestoreRequired),
					LastTransitionTime: v12.Now().Format(util.TimeStampLayout),
				},
				CoordinatorMember: pods.Items[random.Intn(numberOfPods)].Name,
			}, reqLogger); err != nil {
			return errors.WrapIfWithDetails(err, "Could not update status for restore",
				"restore", restore)
		}
		return nil
	}

	return errors.New("No pods found.")
}

func (r *CassandraRestoreReconciler) handleRequiredRestore(ctx context.Context, restore *v2.CassandraRestore,
	cc *v2.CassandraCluster, backup *v2.CassandraBackup, reqLogger *logrus.Entry) error {
	pods, err := r.listPods(ctx, restore.Namespace, k8s.LabelsForCassandraDC(cc, backup.Spec.Datacenter))
	if err != nil {
		return errorfactory.New(errorfactory.ResourceNotReady{}, err, "no pods founds for this dc")
	}

	sr, err := backrest.NewClient(r.Client, cc, k8s.PodByName(pods, restore.Status.CoordinatorMember))
	if err != nil {
		return sidecarError(reqLogger, err)
	}

	restoreStatus, err := sr.PerformRestore(restore, backup)
	if err != nil {
		return sidecarError(reqLogger, err)
	}

	restoreStatus.CoordinatorMember = restore.Status.CoordinatorMember
	if err := UpdateRestoreStatus(r.Client, restore, *restoreStatus, reqLogger); err != nil {
		return errors.WrapIfWithDetails(err, "Could not update status for restore", "restore", restore)
	}

	return nil
}

func sidecarError(reqLogger *logrus.Entry, err error) error {
	reqLogger.Info("Cassandra sidecar communication error checking running restore operation")
	return errorfactory.New(errorfactory.CassandraBackupSidecarNotReady{}, err,
		"cassandra sidecar communication error")
}

func (r *CassandraRestoreReconciler) checkRestoreOperationState(ctx context.Context, restore *v2.CassandraRestore,
	cc *v2.CassandraCluster, backup *v2.CassandraBackup, reqLogger *logrus.Entry) error {

	pods, err := r.listPods(ctx, restore.Namespace, k8s.LabelsForCassandraDC(cc, backup.Spec.Datacenter))
	if err != nil {
		return errorfactory.New(errorfactory.ResourceNotReady{}, err, "no pods founds for this dc")
	}

	restoreId := restore.Status.ID
	if restoreId == "" {
		return errors.New("no Restore operation id provided to be checked")
	}

	// Check Restore operation status
	sr, err := backrest.NewClient(r.Client, cc, k8s.PodByName(pods, restore.Status.CoordinatorMember))
	if err != nil {
		reqLogger.Info("cassandra backup sidecar communication error checking running Operation", "OperationId",
			restoreId)
		return errorfactory.New(errorfactory.CassandraBackupSidecarNotReady{}, err,
			"Icarus sidecar communication error")
	}

	status, err := sr.RestoreStatusByID(restoreId)
	status.CoordinatorMember = restore.Status.CoordinatorMember

	if err != nil {
		reqLogger.Info("cassandra backup sidecar communication error checking running Operation",
			"OperationId", restoreId)
		return errorfactory.New(errorfactory.CassandraBackupSidecarNotReady{}, err,
			"Icarus sidecar communication error")
	}

	if err := UpdateRestoreStatus(r.Client, restore, *status, reqLogger); err != nil {
		return errors.WrapIfWithDetails(err, "could not update status for restore",
			"restore", restore)
	}

	restoreConditionType := v2.RestoreConditionType(restore.Status.Condition.Type)

	// Restore operation failed or canceled,
	if restoreConditionType.IsInError() {
		errorMessage := ""
		if len(restore.Status.Condition.FailureCause) > 0 {
			errorMessage = restore.Status.Condition.FailureCause[0].Message
		}
		return errorfactory.New(errorfactory.CassandraBackupOperationFailure{}, errors.New(errorMessage),
			"Restore operation failed")
	}

	// Restore operation completed successfully
	if restoreConditionType.IsCompleted() {
		return nil
	}

	// restore operation still in progress
	reqLogger.Info("Cassandra backup sidecar operation is still running", "restoreId", restoreId)
	return errorfactory.New(errorfactory.CassandraBackupOperationRunning{},
		errors.New("cassandra backup sidecar restore operation still running"),
		fmt.Sprintf("restore operation id : %s", restoreId))
}

func (r *CassandraRestoreReconciler) listPods(ctx context.Context, namespace string, selector map[string]string) (*v1.PodList, error) {

	clientOpt := &client.ListOptions{
		Namespace:     namespace,
		LabelSelector: labels.SelectorFromSet(selector),
	}

	opt := []client.ListOption{
		clientOpt,
	}

	pl := &v1.PodList{}
	return pl, r.Client.List(ctx, pl, opt...)
}
