package main

import (
	"os"
	"testing"

	"go.linka.cloud/k8s/cert-manager-webhook-k8s-dns/pkg/test/dns"
	"go.linka.cloud/k8s/cert-manager-webhook-k8s-dns/pkg/test/dns/util"
)

var (
	zone = os.Getenv("TEST_ZONE_NAME")
)

func TestRunsSuite(t *testing.T) {
	dnsServer := "127.0.0.1:30053"
	util.RecursiveNameservers = []string{dnsServer}
	util.DefaultDNSPort = "30053"
	// disable propagation check as we do not use a public dns server
	util.PreCheckDNS = func(fqdn, value string, nameservers []string, useAuthoritative bool) (bool, error) {
		return true, nil
	}
	// The manifest path should contain a file named config.json that is a
	// snippet of valid configuration that should be included on the
	// ChallengeRequest passed as part of the test cases.
	if zone == "" {
		zone = "example.com."
	}
	f := dns.NewFixture(&k8sDNSProviderSolver{},
		dns.SetBinariesPath("_out/kubebuilder/bin"),
		dns.SetResolvedZone(zone),
		dns.SetAllowAmbientCredentials(false),
		dns.SetManifestPath("testdata/config"),
		dns.SetDNSServer(dnsServer),
		dns.SetStrict(true),
	)
	f.RunConformance(t)
}
