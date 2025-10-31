// Copyright 2019 Orange
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// 	You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// 	See the License for the specific language governing permissions and
// limitations under the License.

package cassandracluster

import (
	"context"
	"fmt"
	"reflect"
	"strconv"
	"time"

	api "github.com/cscetbon/casskop/api/v2"
	"github.com/cscetbon/casskop/controllers/cassandracluster/cassandrapod"
	"github.com/cscetbon/casskop/controllers/cassandracluster/consts"
	"github.com/cscetbon/casskop/controllers/cassandracluster/sts"
	"github.com/cscetbon/casskop/pkg/k8s"
	"github.com/r3labs/diff"
	"github.com/sirupsen/logrus"
	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Global var used to know if we need to update the CRD
var needUpdate bool

// updateCassandraStatus updates the CRD if the status has changed
// if needUpdate is set that mean that we have updated some fields in the CRD
// This method also stored the annotation cassandraclusters.db.orange.com/last-applied-configuration with last-applied-configuration
func (rcc *CassandraClusterReconciler) updateCassandraStatus(ctx context.Context, cc *api.CassandraCluster,
	status *api.CassandraClusterStatus) error {
	// don't update the status if there aren't any changes.
	if cc.Annotations == nil {
		cc.Annotations = map[string]string{}
	}

	lastApplied, _ := cc.ComputeLastAppliedConfiguration()

	if !needUpdate &&
		reflect.DeepEqual(cc.Status, *status) && //Do We need to update Status ?
		reflect.DeepEqual(cc.Annotations[api.AnnotationLastApplied], string(lastApplied)) && //Do We need to update Annotation ?
		cc.Annotations[api.AnnotationLastApplied] != "" {
		return nil
	}
	needUpdate = false
	//make also deepcopy to avoid pointer conflict
	cc.Annotations[api.AnnotationLastApplied] = string(lastApplied)

	err := rcc.Client.Update(ctx, cc)
	if err != nil {
		logrus.WithFields(logrus.Fields{"cluster": cc.Name, "err": err}).Errorf("Issue when updating CassandraCluster")
	}
	cc.Status = *status.DeepCopy()
	err = rcc.Client.Status().Update(ctx, cc)
	if err != nil {
		logrus.WithFields(logrus.Fields{"cluster": cc.Name, "err": err}).Errorf("Issue when updating CassandraCluster Status")
	}
	return err
}

// getNextCassandraClusterStatus goal is to detect some changes in the status between cassandracluster and its statefulset
// We follow only one change at a Time : so this function will return on the first change found
func (rcc *CassandraClusterReconciler) getNextCassandraClusterStatus(ctx context.Context, cc *api.CassandraCluster, dc, rack int,
	completeDcRackName api.CompleteRackName, storedStatefulSet *appsv1.StatefulSet, status *api.CassandraClusterStatus) error {

	//UpdateStatusIfUpdateResources(cc, dcRackName, storedStatefulSet, status)

	if needToWaitDelayBeforeCheck(cc, completeDcRackName.DcRackName, status) {
		return nil
	}

	if rcc.UpdateStatusIfActionEnded(ctx, cc, completeDcRackName, storedStatefulSet, status) {
		return nil
	}

	//If we set up UnlockNextOperation in CRD we allow to see mode change even last operation didn't ended correctly
	unlockNextOperation := false
	if cc.Spec.UnlockNextOperation &&
		rcc.hasUnschedulablePod(ctx, completeDcRackName) {
		unlockNextOperation = true
	}
	//Do nothing in Initial phase except if we force it
	if status.GetCassandraRackStatus(completeDcRackName.DcRackName).IsInInitialPhase() {
		if !unlockNextOperation {
			ClusterPhaseMetric.set(api.ClusterPhaseInitial, cc.Name)
			return nil
		}
		status.GetCassandraRackStatus(completeDcRackName.DcRackName).SetPendingPhase()
		ClusterPhaseMetric.set(api.ClusterPhasePending, cc.Name)
	}

	lastAction := &status.GetCassandraRackStatus(completeDcRackName.DcRackName).CassandraLastAction

	// Do not check for new action if there is one ongoing or planed
	// Check to discover new changes are not done if action.status is Ongoing or ToDo/Finalizing
	// (a change is already performing)
	// action.status=Continue (which is set when decommission is successful) will be tested to see if we need to
	// decommission more
	// We don't want to check for new operation while there are already ongoing one in order not to break them (ie decommission..)
	// Meanwhile we allow to check for new changes if unlockNextOperation	 has been set (to recover from problems)

	if unlockNextOperation ||
		(rcc.hasNoPodDisruption() &&
			lastAction.Status != api.StatusOngoing &&
			lastAction.Status != api.StatusToDo &&
			lastAction.Status != api.StatusFinalizing) {

		// Update Status if ConfigMap Has Changed
		if UpdateStatusIfconfigMapHasChanged(cc, completeDcRackName.DcRackName, storedStatefulSet, status) {
			return nil
		}

		// Update Status if ConfigMap Has Changed
		if UpdateStatusIfDockerImageHasChanged(cc, completeDcRackName.DcRackName, storedStatefulSet, status) {
			return nil
		}

		rcc.storedStatefulSet = storedStatefulSet
		if rcc.UpdateStatusIfStorageUpsize(completeDcRackName, status) {
			return nil
		}

		// Update Status if There is a ScaleUp or ScaleDown
		if UpdateStatusIfScaling(cc, completeDcRackName.DcRackName, storedStatefulSet, status) {
			return nil
		}

		// Update Status if Topology for SeedList has changed
		if UpdateStatusIfSeedListHasChanged(cc, completeDcRackName.DcRackName, storedStatefulSet, status) {
			return nil
		}

		if UpdateStatusIfRollingRestart(cc, dc, rack, completeDcRackName.DcRackName, status) {
			return nil
		}

		if UpdateStatusIfStatefulSetChanged(completeDcRackName.DcRackName, storedStatefulSet, status) {
			return nil
		}
	} else {
		logrus.WithFields(logrus.Fields{"cluster": cc.Name,
			"dc-rack": completeDcRackName.DcRackName}).Info("We don't check for new action before the cluster become stable again")
	}

	if lastAction.Status == api.StatusToDo && lastAction.Name == api.ActionUpdateResources.Name {
		now := metav1.Now()
		lastAction.StartTime = &now
		lastAction.Status = api.StatusOngoing
	}

	return nil
}

// needToWaitDelayBeforeCheck will return if last action start time is < to api.DefaultDelayWait
// that means the last operation was started only a few seconds ago and checking now would not make any sense
// this is mostly to give cassandra and the operator enough time to correctly stage the action
// DefaultDelayWait is of 2 minutes
func needToWaitDelayBeforeCheck(cc *api.CassandraCluster, dcRackName api.DcRackName, status *api.CassandraClusterStatus) bool {
	lastAction := &status.GetCassandraRackStatus(dcRackName).CassandraLastAction

	if lastAction.StartTime != nil {
		t := *lastAction.StartTime
		now := metav1.Now()

		if t.Add(delayWait()).After(now.Time) {
			logrus.WithFields(logrus.Fields{"cluster": cc.Name,
				"rack": dcRackName}).Info(
				fmt.Sprintf("The Operator Waits %s seconds for the action to start correctly",
					strconv.Itoa(api.DefaultDelayWait)),
			)
			return true
		}
	}
	return false
}

// visible for tests
var delayWait = defaultDelayWait

func defaultDelayWait() time.Duration {
	return api.DefaultDelayWait * time.Second
}

// UpdateStatusIfconfigMapHasChanged updates CassandraCluster Action Status if it detect a changes :
// - a new configmapName in the CRD
// - or the add or remoove of the configmap in the CRD
func UpdateStatusIfconfigMapHasChanged(cc *api.CassandraCluster, dcRackName api.DcRackName, storedStatefulSet *appsv1.StatefulSet, status *api.CassandraClusterStatus) bool {

	var updateConfigMap bool = false

	if storedStatefulSet.Spec.Template.Spec.Volumes == nil && cc.Spec.ConfigMapName != "" {
		logrus.Infof("[%s][%s]: We ask to change ConfigMap New-CRD:%s -> Old-StatefulSet:%s", cc.Name, dcRackName,
			cc.Spec.ConfigMapName, "-")
		updateConfigMap = true
	}
	if storedStatefulSet.Spec.Template.Spec.Volumes != nil {
		var found bool = false
		for _, volume := range storedStatefulSet.Spec.Template.Spec.Volumes {
			if volume.Name == cassandraConfigMapName {
				found = true
				if volume.ConfigMap != nil && volume.ConfigMap.Name != cc.Spec.ConfigMapName {
					logrus.Infof("[%s][%s]: We ask to change ConfigMap New-CRD:%s -> Old-StatefulSet:%s", cc.Name, dcRackName,
						cc.Spec.ConfigMapName, volume.ConfigMap.Name)
					updateConfigMap = true
				}
				break // we have found the configmap
			}
		}
		//If volume for configmap don't exist and we ask for a configmap
		if !found && cc.Spec.ConfigMapName != "" {
			logrus.Infof("[%s][%s]: We ask to change ConfigMap New-CRD:%s -> Old-StatefulSet:%s", cc.Name, dcRackName,
				cc.Spec.ConfigMapName, "-")
			updateConfigMap = true
		}
	}

	if updateConfigMap {
		lastAction := &status.GetCassandraRackStatus(dcRackName).CassandraLastAction
		lastAction.Status = api.StatusToDo
		lastAction.Name = api.ActionUpdateConfigMap.Name
		ClusterActionMetric.set(api.ActionUpdateConfigMap, cc.Name)
		lastAction.StartTime = nil
		lastAction.EndTime = nil
		return true
	}
	return false
}

// UpdateStatusIfDockerImageHasChanged updates CassandraCluster Action Status if it detect a changes in the DockerImage:
func UpdateStatusIfDockerImageHasChanged(cc *api.CassandraCluster, dcRackName api.DcRackName, storedStatefulSet *appsv1.StatefulSet, status *api.CassandraClusterStatus) bool {

	desiredDockerImage := cc.Spec.CassandraImage

	//This needs to be refactor if we load more than 1 container
	if storedStatefulSet.Spec.Template.Spec.Containers != nil {
		for _, container := range storedStatefulSet.Spec.Template.Spec.Containers {
			if container.Name == consts.CassandraContainerName && desiredDockerImage != container.Image {
				{
					logrus.Infof("[%s][%s]: We ask to change DockerImage CRD:%s -> StatefulSet:%s", cc.Name, dcRackName, desiredDockerImage, storedStatefulSet.Spec.Template.Spec.Containers[0].Image)
					lastAction := &status.GetCassandraRackStatus(dcRackName).CassandraLastAction
					lastAction.Status = api.StatusToDo
					lastAction.Name = api.ActionUpdateDockerImage.Name
					ClusterActionMetric.set(api.ActionUpdateDockerImage, cc.Name)
					lastAction.StartTime = nil
					lastAction.EndTime = nil
					return true
				}
			}
		}
	}
	return false
}

func UpdateStatusIfRollingRestart(cc *api.CassandraCluster, dc,
	rack int, dcRackName api.DcRackName, status *api.CassandraClusterStatus) bool {

	if cc.Spec.Topology.DC[dc].Rack[rack].RollingRestart {
		logrus.WithFields(logrus.Fields{"cluster": cc.Name,
			"dc-rack": dcRackName}).Info("Scoping RollingRestart of the Rack")
		lastAction := &status.GetCassandraRackStatus(dcRackName).CassandraLastAction
		lastAction.Status = api.StatusToDo
		lastAction.Name = api.ActionRollingRestart.Name
		ClusterActionMetric.set(api.ActionRollingRestart, cc.Name)
		lastAction.StartTime = nil
		lastAction.EndTime = nil
		cc.Spec.Topology.DC[dc].Rack[rack].RollingRestart = false
		return true
	}
	return false
}

// UpdateStatusIfSeedListHasChanged updates CassandraCluster Action Status if it detects a change
func UpdateStatusIfSeedListHasChanged(cc *api.CassandraCluster, dcRackName api.DcRackName,
	storedStatefulSet *appsv1.StatefulSet, status *api.CassandraClusterStatus) bool {

	storedSeedList := getStoredSeedList(storedStatefulSet)

	//If Automatic Update of SeedList is enabled in the CRD
	if cc.Spec.AutoUpdateSeedList {
		seedList := cc.InitSeedList()
		if changes, err := diff.Diff(storedSeedList, seedList); err == nil && len(changes) != 0 {
			status.SeedList = k8s.MergeSlice(storedSeedList, seedList)
			logrus.Infof("[%s][%s]: We need to update the seed list: %v -> %v",
				cc.Name, dcRackName, storedSeedList, status.SeedList)
		}
	}

	// If seed list has changed in the CRD, we have a manual change on the SeedList.
	// We flag the rack with UpdateSeedList Operation Configuring
	// Once all racks will be enabled with UpdateSeedList=Configuring,
	// then we update to ongoing and start the rollUpgrade
	// This is to ensure that we won't do 2 different kind of operations in different racks at the same time (ex:scaling + updateseedlist)
	if !reflect.DeepEqual(status.SeedList, storedSeedList) {
		logrus.Infof("[%s][%s]: We ask to Change the Cassandra SeedList", cc.Name, dcRackName)
		lastAction := &status.GetCassandraRackStatus(dcRackName).CassandraLastAction
		lastAction.Status = api.StatusConfiguring
		lastAction.Name = api.ActionUpdateSeedList.Name
		ClusterActionMetric.set(api.ActionUpdateSeedList, cc.Name)
		lastAction.StartTime = nil
		lastAction.EndTime = nil
		return true
	}

	return false
}

// UpdateStatusIfScaling will detect any change of replicas
// To Scale Down the operator will need to first decommission the last node from Cassandra before removing it from kubernetes.
// To Scale Up some PodOperations may be scheduled if Auto-pilot is activeted.
func UpdateStatusIfScaling(cc *api.CassandraCluster, dcRackName api.DcRackName, storedStatefulSet *appsv1.StatefulSet,
	status *api.CassandraClusterStatus) bool {
	nodesPerRacks := cc.GetNodesPerRacksStrongType(dcRackName)
	if nodesPerRacks != *storedStatefulSet.Spec.Replicas {
		lastAction := &status.GetCassandraRackStatus(dcRackName).CassandraLastAction
		lastAction.Status = api.StatusToDo
		if nodesPerRacks > *storedStatefulSet.Spec.Replicas {
			lastAction.Name = api.ActionScaleUp.Name
			ClusterActionMetric.set(api.ActionScaleUp, cc.Name)
			logrus.Infof("[%s][%s]: Scaling Cluster : Ask %d and have %d --> ScaleUP", cc.Name, dcRackName, nodesPerRacks, *storedStatefulSet.Spec.Replicas)
		} else {
			logrus.Infof("[%s][%s]: Scaling Cluster : Ask %d and have %d --> ScaleDown", cc.Name, dcRackName, nodesPerRacks, *storedStatefulSet.Spec.Replicas)
			ClusterActionMetric.set(api.ActionScaleDown, cc.Name)
			setDecommissionStatus(status, dcRackName)
			ClusterPhaseMetric.set(api.ClusterPhasePending, cc.Name)
		}
		lastAction.StartTime = nil
		lastAction.EndTime = nil
		return true
	}
	return false
}

// UpdateStatusIfStatefulSetChanged detects if there is a change in the statefulset which was not already caught
// If we detect a Statefulset change with this method, then the operator won't catch it before the statefulset tells
// the operator that a change is ongoing. That means that all statefulsets may do their rolling upgrade in parallel, so
// there will be <nbRacks> node down in // in the cluster.
func UpdateStatusIfStatefulSetChanged(dcRackName api.DcRackName, storedStatefulSet *appsv1.StatefulSet,
	status *api.CassandraClusterStatus) bool {
	// We have not detected any change with out specific tests
	lastAction := &status.GetCassandraRackStatus(dcRackName).CassandraLastAction
	if storedStatefulSet.Status.CurrentRevision != storedStatefulSet.Status.UpdateRevision {
		lastAction.Name = api.ActionUpdateStatefulSet.Name
		lastAction.Status = api.StatusOngoing
		now := metav1.Now()
		lastAction.StartTime = &now
		lastAction.EndTime = nil
		return true
	}
	return false
}

// UpdateStatusIfActionEnded Implement Tests to detect End of Ongoing Actions
func (rcc *CassandraClusterReconciler) UpdateStatusIfActionEnded(ctx context.Context, cc *api.CassandraCluster,
	completeDcRackName api.CompleteRackName, storedStatefulSet *appsv1.StatefulSet, status *api.CassandraClusterStatus) bool {

	rackLastAction := &status.GetCassandraRackStatus(completeDcRackName.DcRackName).CassandraLastAction
	now := metav1.Now()

	if rackLastAction.Status == api.StatusOngoing ||
		rackLastAction.Status == api.StatusContinue {

		nodesPerRacks := cc.GetNodesPerRacksStrongType(completeDcRackName.DcRackName)
		switch rackLastAction.Name {

		case api.ActionScaleUp.Name:

			//Does the Scaling ended ?
			if nodesPerRacks == storedStatefulSet.Status.Replicas {

				labelsForList := k8s.LabelsForCassandraDCRackStrongTypes(cc, completeDcRackName.DcName, completeDcRackName.RackName)
				podsList, err := rcc.ListPodsOrderByNameAscending(ctx, cc.Namespace, labelsForList)
				nb := len(podsList.Items)
				if err != nil || nb < 1 {
					return false
				}
				if nb < int(nodesPerRacks) {
					logrus.WithFields(logrus.Fields{"cluster": cc.Name, "rack": completeDcRackName.DcRackName}).Warn(fmt.Sprintf(
						"Although statefulSet has %d replicas, only %d matching pods found", nodesPerRacks, nb))
					return false
				}
				pod := podsList.Items[nodesPerRacks-1]

				//We need lastPod to be running to consider ScaleUp ended
				if cassandrapod.IsReady(&pod) {
					if hasJoiningNodes, err := rcc.hasJoiningNodes(ctx, cc); err != nil {
						return false
					} else if hasJoiningNodes {
						logrus.WithFields(logrus.Fields{"cluster": cc.Name, "dc": completeDcRackName.DcName,
							"rack": completeDcRackName.RackName, "err": err}).
							Info("Cluster has joining nodes, ScaleUp not yet completed")
						return false
					}

					logrus.WithFields(logrus.Fields{"cluster": cc.Name, "rack": completeDcRackName.DcRackName}).Info("ScaleUp is Done")
					rackLastAction.Status = api.StatusDone
					rackLastAction.EndTime = &now

					labels := map[string]string{"operation-name": api.OperationCleanup}
					if cc.Spec.AutoPilot {
						labels["operation-status"] = api.StatusToDo
					} else {
						labels["operation-status"] = api.StatusManual
					}
					rcc.addPodOperationLabels(ctx, cc, completeDcRackName, labels)

					return true
				}
				return false
			}

		case api.ActionScaleDown.Name:

			if nodesPerRacks == storedStatefulSet.Status.Replicas {
				if cc.Status.GetCassandraRackStatus(completeDcRackName.DcRackName).PodLastOperation.Name == api.OperationDecommission &&
					cc.Status.GetCassandraRackStatus(completeDcRackName.DcRackName).PodLastOperation.Status == api.StatusDone {
					logrus.WithFields(logrus.Fields{"cluster": cc.Name, "rack": completeDcRackName.DcRackName}).Info("ScaleDown is Done")
					rackLastAction.Status = api.StatusDone
					rackLastAction.EndTime = &now
					return true
				}
				logrus.WithFields(logrus.Fields{"cluster": cc.Name, "rack": completeDcRackName.DcRackName}).Info("ScaleDown not yet Completed: Waiting for Pod operation to be Done")
			}

		case api.ClusterPhaseInitial.Name:
			ClusterPhaseMetric.set(api.ClusterPhaseInitial, cc.Name)
			//nothing particular here
			return false

		case api.ActionStorageUpsize.Name:
			//nothing particular here
			return false

		default:
			// Do the update has finished on all pods ?
			if storedStatefulSet.Status.CurrentRevision == storedStatefulSet.Status.UpdateRevision {
				logrus.Infof("[%s][%s]: Update %s is Done", cc.Name, completeDcRackName.DcRackName, rackLastAction.Name)
				rackLastAction.Status = api.StatusDone
				now := metav1.Now()
				rackLastAction.EndTime = &now
				return true
			}

		}

	}
	return false
}

// UpdateCassandraRackStatusPhase goal is to calculate the Cluster Phase according to StatefulSet Status.
// The Phase is: Initializing -> Running <--> Pending
// The Phase is a very high level view of the cluster, for a better view we need to see Actions and Pod Operations
func (rcc *CassandraClusterReconciler) UpdateCassandraRackStatusPhase(ctx context.Context, cc *api.CassandraCluster,
	completeDcRackName api.CompleteRackName, storedStatefulSet *appsv1.StatefulSet, status *api.CassandraClusterStatus) {
	dcRackName := completeDcRackName.DcRackName
	lastAction := &status.GetCassandraRackStatus(dcRackName).CassandraLastAction

	logrusFields := logrus.Fields{"cluster": cc.Name, "rack": dcRackName,
		"ReadyReplicas": storedStatefulSet.Status.ReadyReplicas, "RequestedReplicas": *storedStatefulSet.Spec.Replicas}

	if status.GetCassandraRackStatus(dcRackName).IsInInitialPhase() {
		nodesPerRacks := cc.GetNodesPerRacksStrongType(dcRackName)
		//If we are stuck in initializing state, we can rollback the add of dc which implies decommissioning nodes
		if nodesPerRacks <= 0 {
			logrus.WithFields(logrus.Fields{"cluster": cc.Name,
				"rack": dcRackName}).Warn("Aborting Initializing..., start ScaleDown")
			setDecommissionStatus(status, dcRackName)
			ClusterPhaseMetric.set(api.ClusterPhasePending, cc.Name)
			return
		}

		ClusterPhaseMetric.set(api.ClusterPhaseInitial, cc.Name)

		if sts.IsStatefulSetNotReady(storedStatefulSet) {
			logrus.WithFields(logrusFields).Infof("Initializing StatefulSet: Replicas count is not okay")
			return
		}
		//If yes, just check that lastPod is running
		labelsForList := k8s.LabelsForCassandraDCRackStrongTypes(cc, completeDcRackName.DcName, completeDcRackName.RackName)
		podsList, err := rcc.ListPodsOrderByNameAscending(ctx, cc.Namespace, labelsForList)
		if err != nil || len(podsList.Items) < 1 {
			return
		}
		if len(podsList.Items) < int(nodesPerRacks) {
			logrus.WithFields(logrusFields).Infof("StatefulSet is scaling up")
			return
		}
		pod := podsList.Items[nodesPerRacks-1]
		if cassandrapod.IsReady(&pod) {
			status.GetCassandraRackStatus(dcRackName).SetRunningPhase()
			ClusterPhaseMetric.set(api.ClusterPhaseRunning, cc.Name)
			now := metav1.Now()
			lastAction.EndTime = &now
			lastAction.Status = api.StatusDone
			logrus.WithFields(logrusFields).Infof("StatefulSet: Replicas count is okay")
		}
	}

	//No more in Initializing state
	if sts.IsStatefulSetNotReady(storedStatefulSet) {
		logrus.WithFields(logrusFields).Infof("StatefulSet: Replicas count is not okay")
		status.GetCassandraRackStatus(dcRackName).SetPendingPhase()
		ClusterPhaseMetric.set(api.ClusterPhasePending, cc.Name)
	} else if !status.GetCassandraRackStatus(dcRackName).IsInRunningPhase() {
		logrus.WithFields(logrusFields).Infof("StatefulSet: Rack Phase is not %s", api.ClusterPhaseRunning.Name)
		status.GetCassandraRackStatus(dcRackName).SetRunningPhase()
		ClusterPhaseMetric.set(api.ClusterPhaseRunning, cc.Name)
	}
}

// UpdateCassandraRackStatusFirstPodPerRackInitPhase goal is to calculate the Cluster Subphase of Phase Initializing according to StatefulSet Status.
// The Subphase goes one way: FirstPodPerRack -> NextPodPerRack
func (rcc *CassandraClusterReconciler) UpdateCassandraRackStatusFirstPodPerRackInitPhase(ctx context.Context,
	cc *api.CassandraCluster, completeDcRackName api.CompleteRackName,
	storedStatefulSet *appsv1.StatefulSet, status *api.CassandraClusterStatus) {

	logrusFields := logrus.Fields{"cluster": cc.Name, "rack": completeDcRackName.RackName,
		"phase":         status.GetCassandraRackStatus(completeDcRackName.DcRackName).CassandraPhase,
		"ReadyReplicas": storedStatefulSet.Status.ReadyReplicas, "RequestedReplicas": *storedStatefulSet.Spec.Replicas}

	ClusterPhaseMetric.set(api.ClusterPhaseInitial, cc.Name)

	if sts.IsStatefulSetNotReady(storedStatefulSet) {
		logrus.WithFields(logrusFields).Infof("Initializing StatefulSet: Replicas count is not okay")
		return
	}
	//If yes, just check that lastPod is running
	labels := k8s.LabelsForCassandraDCRackStrongTypes(cc, completeDcRackName.DcName, completeDcRackName.RackName)
	podsList, err := rcc.ListPodsOrderByNameAscending(ctx, cc.Namespace, labels)
	if err != nil || len(podsList.Items) < 1 {
		return
	}
	pod := podsList.Items[len(podsList.Items)-1]
	if cassandrapod.IsReady(&pod) {
		status.GetCassandraRackStatus(completeDcRackName.DcRackName).SetNextPodPerRackInitPhase()
		logrus.WithFields(logrusFields).Infof("StatefulSet: Replicas count is okay")
	}
}

func setDecommissionStatus(status *api.CassandraClusterStatus, dcRackName api.DcRackName) {
	rackStatus := status.GetCassandraRackStatus(dcRackName)
	rackStatus.SetPendingPhase()
	now := metav1.Now()
	lastAction := &rackStatus.CassandraLastAction
	lastAction.StartTime = &now
	lastAction.Status = api.StatusToDo
	lastAction.Name = api.ActionScaleDown.Name
	rackStatus.PodLastOperation.Status = api.StatusToDo
	rackStatus.PodLastOperation.Name = api.OperationDecommission
	rackStatus.PodLastOperation.StartTime = &now
	rackStatus.PodLastOperation.EndTime = nil
	rackStatus.PodLastOperation.Pods = []string{}
	rackStatus.PodLastOperation.PodsOK = []string{}
	rackStatus.PodLastOperation.PodsKO = []string{}
}
