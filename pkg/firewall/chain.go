// Copyright 2019-2023 The Liqo Authors
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

package firewall

import (
	"github.com/google/nftables"

	firewallapi "github.com/liqotech/liqo/apis/networking/v1alpha1/firewall"
)

func addChain(nftconn *nftables.Conn, chain *firewallapi.Chain, table *nftables.Table) error {
	nftChain, err := getChain(nftconn, table, chain)
	if err != nil {
		return err
	}
	if nftChain != nil {
		return nil
	}

	nftChain = &nftables.Chain{Table: table}
	if chain.Name != nil {
		setChainName(nftChain, *chain.Name)
	}
	setHooknum(nftChain, *chain.Hook)
	if chain.Priority != nil {
		setPriority(nftChain, *chain.Priority)
	}
	if chain.Type != nil {
		setType(nftChain, *chain.Type)
	}
	if chain.Policy != nil {
		setPolicy(nftChain, *chain.Policy)
	}
	nftconn.AddChain(nftChain)
	return nil
}

func getChain(nftConn *nftables.Conn, table *nftables.Table,
	chain *firewallapi.Chain) (*nftables.Chain, error) {
	chainlist, err := nftConn.ListChainsOfTableFamily(table.Family)
	if err != nil {
		return nil, err
	}
	for i := range chainlist {
		if chainlist[i].Table.Name != table.Name {
			continue
		}
		if chainlist[i].Name == *chain.Name {
			return chainlist[i], nil
		}
	}
	return nil, nil
}

func setChainName(chain *nftables.Chain, name string) {
	chain.Name = name
}

func setHooknum(chain *nftables.Chain, hooknum firewallapi.ChainHook) {
	switch hooknum {
	case firewallapi.ChainHookPrerouting:
		chain.Hooknum = nftables.ChainHookPrerouting
	case firewallapi.ChainHookInput:
		chain.Hooknum = nftables.ChainHookInput
	case firewallapi.ChainHookForward:
		chain.Hooknum = nftables.ChainHookForward
	case firewallapi.ChainHookOutput:
		chain.Hooknum = nftables.ChainHookOutput
	case firewallapi.ChainHookPostrouting:
		chain.Hooknum = nftables.ChainHookPostrouting
	case firewallapi.ChainHookIngress:
		chain.Hooknum = nftables.ChainHookIngress
	}
}

func getHooknum(hooknum nftables.ChainHook) firewallapi.ChainHook {
	switch hooknum {
	case *nftables.ChainHookPrerouting:
		return firewallapi.ChainHookPrerouting
	case *nftables.ChainHookInput:
		return firewallapi.ChainHookInput
	case *nftables.ChainHookForward:
		return firewallapi.ChainHookForward
	case *nftables.ChainHookOutput:
		return firewallapi.ChainHookOutput
	case *nftables.ChainHookPostrouting:
		return firewallapi.ChainHookPostrouting
	case *nftables.ChainHookIngress:
		return firewallapi.ChainHookIngress
	default:
		return ""
	}
}

func setPriority(chain *nftables.Chain, priority firewallapi.ChainPriority) {
	chain.Priority = (*nftables.ChainPriority)(&priority)
}

func setType(chain *nftables.Chain, chainType firewallapi.ChainType) {
	switch chainType {
	case firewallapi.ChainTypeFilter:
		chain.Type = nftables.ChainTypeFilter
	case firewallapi.ChainTypeRoute:
		chain.Type = nftables.ChainTypeRoute
	case firewallapi.ChainTypeNAT:
		chain.Type = nftables.ChainTypeNAT
	}
}

func getType(chaintype nftables.ChainType) firewallapi.ChainType {
	switch chaintype {
	case nftables.ChainTypeFilter:
		return firewallapi.ChainTypeFilter
	case nftables.ChainTypeRoute:
		return firewallapi.ChainTypeRoute
	case nftables.ChainTypeNAT:
		return firewallapi.ChainTypeNAT
	default:
		return ""
	}
}

func setPolicy(chain *nftables.Chain, policy firewallapi.ChainPolicy) {
	switch policy {
	case firewallapi.ChainPolicyDrop:
		p := nftables.ChainPolicyDrop
		chain.Policy = &p
	case firewallapi.ChainPolicyAccept:
		p := nftables.ChainPolicyAccept
		chain.Policy = &p
	}
}

func getPolicy(policy nftables.ChainPolicy) firewallapi.ChainPolicy {
	switch policy {
	case nftables.ChainPolicyDrop:
		return firewallapi.ChainPolicyDrop
	case nftables.ChainPolicyAccept:
		return firewallapi.ChainPolicyAccept
	default:
		return ""
	}
}

// isChainOutdated checks if the chain has to be deleted.
// A chain must be deleted when it's properties change
// or when it is not contained in the FirewallConfiguration CRD.

func isChainOutdated(nftChain *nftables.Chain, chains []firewallapi.Chain) bool {
	for i := range chains {
		// if chain names are not equal, continue till the end of the list
		// if the chain is not present, delete it
		if chains[i].Name == nil || *chains[i].Name != nftChain.Name {
			continue
		}
		// if chain names are equal, check if the chain has been modified
		if isChainModified(nftChain, &chains[i]) {
			// if the chain has been modified, delete it
			return true
		}
		// if the chain has not been modified, do not delete it
		return false
	}
	return true
}

// isChainModified checks if the chain has been modified.
// It does not consider policies since they can be modified without deleting the chain.
func isChainModified(nftChain *nftables.Chain, chain *firewallapi.Chain) bool {
	if chain.Type != nil && *chain.Type != getType(nftChain.Type) {
		return true
	}
	if chain.Hook != nil && *chain.Hook != getHooknum(*nftChain.Hooknum) {
		return true
	}
	if chain.Priority != nil && *chain.Priority != firewallapi.ChainPriority(*nftChain.Priority) {
		return true
	}
	if chain.Policy != nil && *chain.Policy != getPolicy(*nftChain.Policy) {
		return true
	}
	return false
}