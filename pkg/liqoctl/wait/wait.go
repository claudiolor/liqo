// Copyright 2019-2022 The Liqo Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package wait

import (
	"context"
	"fmt"
	"time"

	"k8s.io/apimachinery/pkg/util/wait"
	"sigs.k8s.io/controller-runtime/pkg/client"

	discoveryv1alpha1 "github.com/liqotech/liqo/apis/discovery/v1alpha1"
	"github.com/liqotech/liqo/pkg/liqoctl/factory"
	"github.com/liqotech/liqo/pkg/utils"
	fcutils "github.com/liqotech/liqo/pkg/utils/foreignCluster"
	getters "github.com/liqotech/liqo/pkg/utils/getters"
)

// Waiter is a struct that contains the necessary information to wait for resource events.
type Waiter struct {
	// Printer is the object used to output messages in the appropriate format.
	Printer *factory.Printer
	// crClient is the controller runtime client.
	CRClient client.Client
}

// NewWaiterFromFactory creates a new Waiter object from the given factory.
func NewWaiterFromFactory(f *factory.Factory) *Waiter {
	return &Waiter{
		Printer:  f.Printer,
		CRClient: f.CRClient,
	}
}

// ForUnpeering waits until the status on the foreiglcusters resource states that the in/outgoing peering has been successfully
// set to None or the timeout expires.
func (w *Waiter) ForUnpeering(ctx context.Context, remoteClusterID *discoveryv1alpha1.ClusterIdentity) error {
	remName := remoteClusterID.ClusterName
	s := w.Printer.StartSpinner(fmt.Sprintf("Unpeering from the remote cluster %q", remName))
	err := fcutils.PollForEvent(ctx, w.CRClient, remoteClusterID, fcutils.IsUnpeered, 1*time.Second)
	if client.IgnoreNotFound(err) != nil {
		s.Fail(fmt.Sprintf("Failed unpeering from remote cluster %q: %s", remName, err.Error()))
		return err
	}
	s.Success(fmt.Sprintf("Successfully unpeered from remote cluster %q", remName))
	return nil
}

// ForOutgoingUnpeering waits until the status on the foreiglcusters resource states that the outgoing peering has been successfully
// set to None or the timeout expires.
func (w *Waiter) ForOutgoingUnpeering(ctx context.Context, remoteClusterID *discoveryv1alpha1.ClusterIdentity) error {
	remName := remoteClusterID.ClusterName
	s := w.Printer.StartSpinner(fmt.Sprintf("Disabling outgoing peering to the remote cluster %q", remName))
	err := fcutils.PollForEvent(ctx, w.CRClient, remoteClusterID, fcutils.IsOutgoingPeeringNone, 1*time.Second)
	if client.IgnoreNotFound(err) != nil {
		s.Fail(fmt.Sprintf("Failed disabling outgoing peering to the remote cluster %q: %s", remName, err.Error()))
		return err
	}
	s.Success(fmt.Sprintf("Successfully disabled outgoing peering to the remote cluster %q", remName))
	return nil
}

// ForAuth waits until the authentication has been established with the remote cluster or the timeout expires.
func (w *Waiter) ForAuth(ctx context.Context, remoteClusterID *discoveryv1alpha1.ClusterIdentity) error {
	remName := remoteClusterID.ClusterName
	s := w.Printer.StartSpinner(fmt.Sprintf("Waiting for authentication to the cluster %q", remName))
	err := fcutils.PollForEvent(ctx, w.CRClient, remoteClusterID, fcutils.IsAuthenticated, 1*time.Second)
	if err != nil {
		s.Fail(fmt.Sprintf("Authentication to the remote cluster %q failed: %s", remName, err.Error()))
		return err
	}
	s.Success(fmt.Sprintf("Authenticated to cluster %q", remName))
	return nil
}

// ForNetwork waits until the networking has been established with the remote cluster or the timeout expires.
func (w *Waiter) ForNetwork(ctx context.Context, remoteClusterID *discoveryv1alpha1.ClusterIdentity) error {
	remName := remoteClusterID.ClusterName
	s := w.Printer.StartSpinner(fmt.Sprintf("Waiting for network to the remote cluster %q", remName))
	err := fcutils.PollForEvent(ctx, w.CRClient, remoteClusterID, fcutils.IsNetworkingEstablished, 1*time.Second)
	if err != nil {
		s.Fail(fmt.Sprintf("Failed establishing networking to the remote cluster %q: %s", remName, err.Error()))
		return err
	}
	s.Success(fmt.Sprintf("Network established to the remote cluster %q", remName))
	return nil
}

// ForOutgoingPeering waits until the status on the foreiglcusters resource states that the outgoing peering has been successfully
// established or the timeout expires.
func (w *Waiter) ForOutgoingPeering(ctx context.Context, remoteClusterID *discoveryv1alpha1.ClusterIdentity) error {
	remName := remoteClusterID.ClusterName
	s := w.Printer.StartSpinner(fmt.Sprintf("Activating outgoing peering to the remote cluster %q", remName))
	err := fcutils.PollForEvent(ctx, w.CRClient, remoteClusterID, fcutils.IsOutgoingJoined, 1*time.Second)
	if err != nil {
		s.Fail(fmt.Sprintf("Failed activating outgoing peering to the remote cluster %q: %s", remName, err.Error()))
		return err
	}
	s.Success(fmt.Sprintf("Outgoing peering activated to the remote cluster %q", remName))
	return nil
}

// ForNode waits until the node has been added to the cluster or the timeout expires.
func (w *Waiter) ForNode(ctx context.Context, remoteClusterID *discoveryv1alpha1.ClusterIdentity) error {
	remName := remoteClusterID.ClusterName
	s := w.Printer.StartSpinner(fmt.Sprintf("Waiting for node to be created for the remote cluster %q", remName))

	err := wait.PollImmediateUntilWithContext(ctx, 1*time.Second, func(ctx context.Context) (done bool, err error) {
		node, err := getters.GetNodeByClusterID(ctx, w.CRClient, remoteClusterID)
		if err != nil {
			return false, client.IgnoreNotFound(err)
		}

		return utils.IsNodeReady(node), nil
	})
	if err != nil {
		s.Fail(fmt.Sprintf("Failed waiting for node to be created for remote cluster %q: %s", remName, err.Error()))
		return err
	}
	s.Success(fmt.Sprintf("Node created for remote cluster %q", remName))
	return nil
}
