package dummy

import (
	"io"

	"fmt"

	"github.com/golang/glog"
	apiv1 "k8s.io/api/core/v1"
	"gopkg.in/gcfg.v1"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider"
	"k8s.io/autoscaler/cluster-autoscaler/utils/errors"
	"k8s.io/kubernetes/plugin/pkg/scheduler/schedulercache"
)

const ProviderName = "dummy"

type DummyManager struct {
}

type Config struct {
	Foo struct {
		Bar string
	}
}

type DummyCloudProvider struct {
	rl *cloudprovider.ResourceLimiter
}

type DummyNodeGroup struct {
	min       int
	max       int
	current   int
	name      string
	nodeNames []string
}

func CreateDummyManager(configReader io.Reader, discoveryOpts cloudprovider.NodeGroupDiscoveryOptions) (*DummyManager, error) {
	if configReader != nil {
		var config Config
		if err := gcfg.ReadInto(&config, configReader); err != nil {
			glog.Errorf("Couldn't read config: %v", err)
			return nil, err
		}
	}	
	return &DummyManager{}, nil
}

func BuildDummyCloudProvider(dummyManager *DummyManager, resourceLimiter *cloudprovider.ResourceLimiter) (cloudprovider.CloudProvider, error) {
	return &DummyCloudProvider{rl: resourceLimiter}, nil
}

// Name returns name of the cloud provider.
func (d *DummyCloudProvider) Name() string {
	return ProviderName
}

var ng = &DummyNodeGroup{
	min:       1,
	max:       3,
	current:   2,
	name:      "k8sbmcs",
	nodeNames: []string{"k8s-master-ad1-0.k8smasterad1.k8sbmcs.oraclevcn.com", "k8s-worker-ad1-0.k8sworkerad1.k8sbmcs.oraclevcn.com"},
}

// NodeGroups returns all node groups configured for this cloud provider.
func (d *DummyCloudProvider) NodeGroups() []cloudprovider.NodeGroup {
	glog.Info("NodeGroups() called")

	return []cloudprovider.NodeGroup{ng}
}

// NodeGroupForNode returns the node group for the given node, nil if the node
// should not be processed by cluster autoscaler, or non-nil error if such
// occurred. Must be implemented.
func (d *DummyCloudProvider) NodeGroupForNode(n *apiv1.Node) (cloudprovider.NodeGroup, error) {
	glog.Info("NodeGroupForNode() called ", n.GetName(), " returning ng k8sbmcs")
	return ng, nil
}

// Pricing returns pricing model for this cloud provider or error if not available.
// Implementation optional.
func (d *DummyCloudProvider) Pricing() (cloudprovider.PricingModel, errors.AutoscalerError) {
	return nil, cloudprovider.ErrNotImplemented
}

// GetAvailableMachineTypes get all machine types that can be requested from the cloud provider.
// Implementation optional.
func (d *DummyCloudProvider) GetAvailableMachineTypes() ([]string, error) {
	return []string{}, nil
}

// NewNodeGroup builds a theoretical node group based on the node definition provided. The node group is not automatically
// created on the cloud provider side. The node group is not returned by NodeGroups() until it is created.
// Implementation optional.
func (d *DummyCloudProvider) NewNodeGroup(machineType string, labels map[string]string, systemLabels map[string]string, extraResources map[string]resource.Quantity) (cloudprovider.NodeGroup, error) {
	return nil, cloudprovider.ErrNotImplemented
}

// GetResourceLimiter returns struct containing limits (max, min) for resources (cores, memory etc.).
func (d *DummyCloudProvider) GetResourceLimiter() (*cloudprovider.ResourceLimiter, error) {
	return d.rl, nil
}

// Cleanup cleans up open resources before the cloud provider is destroyed, i.e. go routines etc.
func (d *DummyCloudProvider) Cleanup() error {
	glog.Info("Cleanup() called")
	return nil
}

// Refresh is called before every main loop and can be used to dynamically update cloud provider state.
// In particular the list of node groups returned by NodeGroups can change as a result of CloudProvider.Refresh().
func (d *DummyCloudProvider) Refresh() error {
	glog.Info("Refresh() called")
	return nil
}

// MaxSize returns maximum size of the node group.
func (n *DummyNodeGroup) MaxSize() int {
	return n.max
}

// MinSize returns minimum size of the node group.
func (n *DummyNodeGroup) MinSize() int {
	return n.min
}

// TargetSize returns the current target size of the node group. It is possible that the
// number of nodes in Kubernetes is different at the moment but should be equal
// to Size() once everything stabilizes (new nodes finish startup and registration or
// removed nodes are deleted completely). Implementation required.
func (n *DummyNodeGroup) TargetSize() (int, error) {
	return n.current, nil
}

// IncreaseSize increases the size of the node group. To delete a node you need
// to explicitly name it and use DeleteNode. This function should wait until
// node group size is updated. Implementation required.
func (n *DummyNodeGroup) IncreaseSize(delta int) error {
	glog.Infof("IncreaseSize(%d) called", delta)
	n.current++
	return nil
}

// DeleteNodes deletes nodes from this node group. Error is returned either on
// failure or if the given node doesn't belong to this node group. This function
// should wait until node group size is updated. Implementation required.
func (n *DummyNodeGroup) DeleteNodes(nodes []*apiv1.Node) error {
	glog.Info("DeleteNodes() called ", nodes)
	return cloudprovider.ErrNotImplemented
}

// DecreaseTargetSize decreases the target size of the node group. This function
// doesn't permit to delete any existing node and can be used only to reduce the
// request for new nodes that have not been yet fulfilled. Delta should be negative.
// It is assumed that cloud provider will not delete the existing nodes when there
// is an option to just decrease the target. Implementation required.
func (n *DummyNodeGroup) DecreaseTargetSize(delta int) error {
	glog.Infof("DecreaseTargetSize(%d) called", delta)
	n.current--
	return nil
}

// Id returns an unique identifier of the node group.
func (n *DummyNodeGroup) Id() string {
	return n.name
}

// Debug returns a string containing all information regarding this node group.
func (n *DummyNodeGroup) Debug() string {
	return fmt.Sprintf("%v", n)
}

// Nodes returns a list of all nodes that belong to this node group.
func (n *DummyNodeGroup) Nodes() ([]string, error) {
	return n.nodeNames, nil
}

// TemplateNodeInfo returns a schedulercache.NodeInfo structure of an empty
// (as if just started) node. This will be used in scale-up simulations to
// predict what would a new node look like if a node group was expanded. The returned
// NodeInfo is expected to have a fully populated Node object, with all of the labels,
// capacity and allocatable information as well as all pods that are started on
// the node by default, using manifest (most likely only kube-proxy). Implementation optional.
func (n *DummyNodeGroup) TemplateNodeInfo() (*schedulercache.NodeInfo, error) {
	return nil, cloudprovider.ErrNotImplemented
}

// Exist checks if the node group really exists on the cloud provider side. Allows to tell the
// theoretical node group from the real one. Implementation required.
func (n *DummyNodeGroup) Exist() bool {
	return true
}

// Create creates the node group on the cloud provider side. Implementation optional.
func (n *DummyNodeGroup) Create() error {
	return cloudprovider.ErrNotImplemented
}

// Delete deletes the node group on the cloud provider side.
// This will be executed only for autoprovisioned node groups, once their size drops to 0.
// Implementation optional.
func (n *DummyNodeGroup) Delete() error {
	return cloudprovider.ErrNotImplemented
}

// Autoprovisioned returns true if the node group is autoprovisioned. An autoprovisioned group
// was created by CA and can be deleted when scaled to 0.
func (n *DummyNodeGroup) Autoprovisioned() bool {
	return false
}
