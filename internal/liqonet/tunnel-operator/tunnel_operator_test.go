package tunneloperator

import (
	"net"

	"github.com/containernetworking/plugins/pkg/ns"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/vishvananda/netlink"

	"github.com/liqotech/liqo/pkg/liqonet/netns"
)

var _ = Describe("TunnelOperator", func() {
	Describe("setup gateway namespace", func() {
		Context("creating a new gateway namespace", func() {
			JustAfterEach(func() {
				link, err := netlink.LinkByName(hostVethName)
				if err != nil {
					Expect(err).Should(MatchError("Link not found"))
				}
				if err != nil && err.Error() != "Link not found" {
					Expect(err).ShouldNot(HaveOccurred())
				}
				if link != nil {
					Expect(netlink.LinkDel(link)).ShouldNot(HaveOccurred())
				}
				Expect(netns.DeleteNetns(gatewayNetnsName)).ShouldNot(HaveOccurred())
			})
			It("should return nil", func() {
				tc := &TunnelController{}
				err := tc.setUpGWNetns(gatewayNetnsName, hostVethName, gatewayVethName, "169.254.1.134/32", 1420)
				Expect(err).ShouldNot(HaveOccurred())
				// Check that we have the veth interface in host namespace
				err = tc.hostNetns.Do(func(ns ns.NetNS) error {
					defer GinkgoRecover()
					link, err := netlink.LinkByName(hostVethName)
					Expect(err).ShouldNot(HaveOccurred())
					Expect(link.Attrs().MTU).Should(BeNumerically("==", 1420))
					return nil
				})
				// Check that we have the veth interface in gateway namespace
				err = tc.gatewayNetns.Do(func(ns ns.NetNS) error {
					defer GinkgoRecover()
					link, err := netlink.LinkByName(gatewayVethName)
					Expect(err).ShouldNot(HaveOccurred())
					Expect(link.Attrs().MTU).Should(BeNumerically("==", 1420))
					addresses, err := netlink.AddrList(link, netlink.FAMILY_V4)
					Expect(addresses[0].IPNet.String()).Should(Equal("169.254.1.134/32"))
					Expect(err).ShouldNot(HaveOccurred())

					return nil
				})
				Expect(tc.hostNetns.Close()).ShouldNot(HaveOccurred())
				Expect(tc.gatewayNetns.Close()).ShouldNot(HaveOccurred())

			})

			It("incorrect name for veth interface, should return error", func() {
				err := tc.setUpGWNetns(gatewayNetnsName, "", gatewayVethName, "169.254.1.134/24", 1420)
				Expect(err).Should(HaveOccurred())
				Expect(err).Should(MatchError("failed to make veth pair: LinkAttrs.Name cannot be empty"))
			})

			It("incorrect ip address for veth interface, should return error", func() {
				tc := &TunnelController{}
				err := tc.setUpGWNetns(gatewayNetnsName, hostVethName, gatewayVethName, "169.254.1.1.34/24", 1420)
				Expect(err).Should(HaveOccurred())
				Expect(err).Should(MatchError(&net.ParseError{Text: "169.254.1.1.34/24", Type: "CIDR address"}))
			})
		})
	})
})