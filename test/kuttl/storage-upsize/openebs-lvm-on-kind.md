# Setting up OpenEBS LVM LocalPV on kind for Online PVC Expansion

This guide explains how to set up LVM (Logical Volume Manager) inside kind (Kubernetes in Docker) containers to enable online PVC expansion testing with OpenEBS LVM LocalPV.

## Overview

Online PVC expansion allows you to resize persistent volume claims without restarting pods. This is useful for testing applications that need to handle storage expansion dynamically.

In casskop domain it is used to test the storage upsize feature.
This feature requires online PVC expansion capability from the underlying storage provider which is available on all major cloud providers.

## Prerequisites

- Docker installed and running
- kind installed (`go install sigs.k8s.io/kind@latest` or via package manager)
- kubectl installed and configured
- sudo access (for creating directories)

## Step 1: Create a kind Cluster with Extra Mounts

First, create a kind configuration file that mounts extra storage paths:

```yaml
# kind-lvm-config.yaml
kind: Cluster
apiVersion: kind.x-k8s.io/v1alpha4
nodes:
- role: control-plane
- role: worker
  extraMounts:
  - hostPath: /tmp/kind-worker1-lvm
    containerPath: /mnt/disks
- role: worker
  extraMounts:
  - hostPath: /tmp/kind-worker2-lvm
    containerPath: /mnt/disks
```

Create the necessary directories and the cluster:

```bash
sudo mkdir -p /tmp/kind-worker1-lvm /tmp/kind-worker2-lvm
kind create cluster --config kind-lvm-config.yaml --name lvm-test
```

## Step 2: Set up LVM Inside Each kind Worker Node

For each worker node, you need to install LVM tools and create a volume group.

### Manual Setup

```bash
# Get the worker node names
kubectl get nodes

# For the first worker node
docker exec -it kind-lvm-test-worker bash

# Inside the container, run:
apt-get update
apt-get install -y lvm2

# Create a loop device (simulating a physical disk)
truncate -s 10G /mnt/disks/disk.img
losetup -f /mnt/disks/disk.img
LOOP_DEVICE=$(losetup -j /mnt/disks/disk.img | cut -d: -f1)
echo "Loop device: $LOOP_DEVICE"

# Create LVM physical volume
pvcreate $LOOP_DEVICE

# Create LVM volume group (name must match StorageClass configuration)
vgcreate lvmvg $LOOP_DEVICE

# Verify the setup
vgs
pvs
lvs

# Exit the container
exit
```

Repeat the same process for the second worker node:

```bash
docker exec -it kind-lvm-test-worker2 bash
# ... repeat the same commands above ...
exit
```

### Automated Setup Script

Alternatively, use this script to automate the LVM setup:

```bash
#!/bin/bash
# setup-lvm-in-kind.sh

CLUSTER_NAME=${1:-lvm-test}
WORKER_NODES=$(kind get nodes --name $CLUSTER_NAME | grep worker)

for NODE in $WORKER_NODES; do
  echo "Setting up LVM on $NODE..."
  
  docker exec $NODE bash -c '
    apt-get update -qq && apt-get install -y -qq lvm2 > /dev/null 2>&1
    mkdir -p /mnt/disks
    truncate -s 10G /mnt/disks/disk.img
    LOOP_DEVICE=$(losetup -f)
    losetup $LOOP_DEVICE /mnt/disks/disk.img
    pvcreate $LOOP_DEVICE
    vgcreate lvmvg $LOOP_DEVICE
    echo "LVM setup complete on $(hostname)"
    vgs
  '
done

echo "LVM setup complete on all worker nodes!"
```

Make it executable and run:

```bash
chmod +x setup-lvm-in-kind.sh
./setup-lvm-in-kind.sh lvm-test
```

## Step 3: Install OpenEBS LVM LocalPV

Install the OpenEBS LVM operator:

```bash
kubectl apply -f https://openebs.github.io/charts/lvm-operator.yaml
```

Wait for all OpenEBS pods to be ready:

```bash
kubectl get pods -n openebs -w
```

You should see pods like:
- `openebs-lvm-localpv-controller`
- `openebs-lvm-localpv-node` (one per worker node)

## Step 4: Create a StorageClass with Volume Expansion Enabled

Create a StorageClass that uses the LVM volume group:

```yaml
# openebs-lvm-sc.yaml
apiVersion: storage.k8s.io/v1
kind: StorageClass
metadata:
  name: openebs-lvm-sc
provisioner: local.csi.openebs.io
parameters:
  storage: "lvm"
  volgroup: "lvmvg"
allowVolumeExpansion: true
volumeBindingMode: WaitForFirstConsumer
```

Apply the StorageClass:

```bash
kubectl apply -f openebs-lvm-sc.yaml
```

**Important:** The `volgroup` parameter must match the volume group name you created in Step 2 (`lvmvg`).

## Step 5: Create a Test PVC

Create a PersistentVolumeClaim using the new StorageClass:

```yaml
# test-lvm-pvc.yaml
apiVersion: v1
kind: PersistentVolumeClaim
metadata:
  name: lvm-pvc
spec:
  storageClassName: openebs-lvm-sc
  accessModes:
    - ReadWriteOnce
  resources:
    requests:
      storage: 1Gi
```

Apply the PVC:

```bash
kubectl apply -f test-lvm-pvc.yaml
```

## Step 6: Create a Test Application

Create a deployment that uses the PVC:

```yaml
# test-lvm-pod.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: test-lvm-app
spec:
  replicas: 1
  selector:
    matchLabels:
      app: test-lvm
  template:
    metadata:
      labels:
        app: test-lvm
    spec:
      containers:
      - name: test-container
        image: nginx
        volumeMounts:
        - name: data
          mountPath: /data
        command: ["/bin/sh"]
        args: ["-c", "while true; do df -h /data; sleep 30; done"]
      volumes:
      - name: data
        persistentVolumeClaim:
          claimName: lvm-pvc
```

Apply the deployment:

```bash
kubectl apply -f test-lvm-pod.yaml
```

Wait for the pod to be running:

```bash
kubectl get pods -l app=test-lvm -w
```

## Step 7: Test Online PVC Expansion

Now you can test the online expansion feature:

### Check Current Size

```bash
# Check PVC size
kubectl get pvc lvm-pvc

# Get the pod name
POD_NAME=$(kubectl get pod -l app=test-lvm -o jsonpath='{.items[0].metadata.name}')

# Check filesystem size inside the pod
kubectl exec $POD_NAME -- df -h /data
```

### Expand the PVC

Expand the PVC while the pod is still running:

```bash
kubectl patch pvc lvm-pvc -p '{"spec":{"resources":{"requests":{"storage":"3Gi"}}}}'
```

### Monitor the Expansion

Watch the PVC status change:

```bash
kubectl get pvc lvm-pvc -w
```

You should see the status transition through:
1. `Resizing` - Expansion in progress
2. `FileSystemResizePending` - Volume expanded, filesystem resize pending
3. `Bound` - Expansion complete

### Verify the New Size

Check that the filesystem has been resized without pod restart:

```bash
kubectl exec $POD_NAME -- df -h /data
```

You should see the increased storage size reflected in the output.

### Verify Pod Was Not Restarted

```bash
kubectl get pod $POD_NAME -o jsonpath='{.status.containerStatuses[0].restartCount}'
```

This should return `0`, confirming no restart occurred.

## Verification Commands

Here's a quick reference of useful verification commands:

```bash
# Check LVM setup on a worker node
docker exec kind-lvm-test-worker vgs
docker exec kind-lvm-test-worker pvs
docker exec kind-lvm-test-worker lvs

# Check OpenEBS components
kubectl get pods -n openebs
kubectl get sc
kubectl get pvc
kubectl get pv

# Check PVC events
kubectl describe pvc lvm-pvc

# Watch pod logs
kubectl logs -l app=test-lvm -f
```

## Troubleshooting

### PVC Stuck in Pending

**Problem:** PVC remains in `Pending` state.

**Solutions:**
- Verify LVM is properly set up on worker nodes: `docker exec kind-lvm-test-worker vgs`
- Check OpenEBS LVM controller logs: `kubectl logs -n openebs -l app=openebs-lvm-controller`
- Ensure volume group name in StorageClass matches: `volgroup: "lvmvg"`

### Expansion Not Working

**Problem:** PVC expansion is stuck or fails.

**Solutions:**
- Check if `allowVolumeExpansion: true` is set in StorageClass
- Verify there's enough space in the volume group: `docker exec kind-lvm-test-worker vgs`
- Check OpenEBS node pod logs: `kubectl logs -n openebs -l app=openebs-lvm-node`

### Loop Device Not Found After Restart

**Problem:** Loop devices disappear after kind container restart.

**Solution:** Loop devices are ephemeral in containers. After restarting kind, you'll need to re-run the LVM setup commands or script.

## Important Notes

1. **Ephemeral Setup:** Loop devices created in kind containers don't persist across container restarts.  For permanent testing environments, consider using a VM with actual block devices.

2. **Volume Group Naming:** The volume group name (`lvmvg`) must match exactly between:
    - The LVM setup commands (`vgcreate lvmvg`)
    - The StorageClass parameters (`volgroup: "lvmvg"`)

3. **Storage Limits:** The loop device size (10G in this example) limits the total storage available for all PVCs on that node.

4. **Production Use:** This setup is intended for development and testing only. For production, use actual block devices or cloud provider storage solutions.

5. **Multiple Workers:** Each worker node needs its own LVM volume group setup.  The automated script handles this automatically.

## Cleanup

To clean up the environment:

```bash
# Delete the test resources
kubectl delete deployment test-lvm-app
kubectl delete pvc lvm-pvc
kubectl delete sc openebs-lvm-sc

# Delete OpenEBS
kubectl delete -f https://openebs.github.io/charts/lvm-operator.yaml

# Delete the kind cluster
kind delete cluster --name lvm-test

# Remove temporary directories
sudo rm -rf /tmp/kind-worker1-lvm /tmp/kind-worker2-lvm
```

## Additional Resources

- [OpenEBS LVM LocalPV Documentation](https://github.com/openebs/lvm-localpv)
- [Kubernetes Volume Expansion Documentation](https://kubernetes.io/docs/concepts/storage/persistent-volumes/#expanding-persistent-volumes-claims)
- [kind Documentation](https://kind.sigs.k8s.io/)

## License

This documentation is provided as-is for educational and testing purposes. 