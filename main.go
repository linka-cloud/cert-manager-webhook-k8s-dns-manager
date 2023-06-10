package main

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/bombsimon/logrusr"
	"github.com/go-logr/logr"
	"github.com/jetstack/cert-manager/pkg/acme/webhook/apis/acme/v1alpha1"
	"github.com/jetstack/cert-manager/pkg/acme/webhook/cmd"
	"github.com/sirupsen/logrus"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"

	dnsv1alpha1 "go.linka.cloud/k8s/cert-manager-webhook-k8s-dns/pkg/api/dns/v1alpha1"
)

var GroupName = os.Getenv("GROUP_NAME")

const (
	AnnotationKey = "dns.linka.cloud/acme-challenge"
)

func main() {
	if GroupName == "" {
		panic("GROUP_NAME must be specified")
	}

	// This will register our custom DNS provider with the webhook serving
	// library, making it available as an API under the provided GroupName.
	// You can register multiple DNS provider implementations with a single
	// webhook, where the Name() method will be used to disambiguate between
	// the different implementations.
	cmd.RunWebhookServer(GroupName,
		&k8sDNSProviderSolver{},
	)
}

// k8sDNSProviderSolver implements the provider-specific logic needed to
// 'present' an ACME challenge TXT record for your own DNS provider.
// To do so, it must implement the `github.com/jetstack/cert-manager/pkg/acme/webhook.Solver`
// interface.
type k8sDNSProviderSolver struct {
	// If a Kubernetes 'clientset' is needed, you must:
	// 1. uncomment the additional `client` field in this structure below
	// 2. uncomment the "k8s.io/client-go/kubernetes" import at the top of the file
	// 3. uncomment the relevant code in the Initialize method below
	// 4. ensure your webhook's service account has the required RBAC role
	//    assigned to it for interacting with the Kubernetes APIs you need.
	client *kubernetes.Clientset
	log    logr.Logger
}

// k8sDNSProviderConfig is a structure that is used to decode into when
// solving a DNS01 challenge.
// This information is provided by cert-manager, and may be a reference to
// additional configuration that's needed to solve the challenge for this
// particular certificate or issuer.
// This typically includes references to Secret resources containing DNS
// provider credentials, in cases where a 'multi-tenant' DNS solver is being
// created.
// If you do *not* require per-issuer or per-certificate configuration to be
// provided to your webhook, you can skip decoding altogether in favour of
// using CLI flags or similar to provide configuration.
// You should not include sensitive information here. If credentials need to
// be used by your provider here, you should reference a Kubernetes Secret
// resource and fetch these credentials using a Kubernetes clientset.
type k8sDNSProviderConfig struct {
	// Change the two fields below according to the format of the configuration
	// to be decoded.
	// These fields will be set by users in the
	// `issuer.spec.acme.dns01.providers.webhook.config` field.
	Namespace string `json:"namespace"`
}

// Name is used as the name for this DNS solver when referencing it on the ACME
// Issuer resource.
// This should be unique **within the group name**, i.e. you can have two
// solvers configured with the same Name() **so long as they do not co-exist
// within a single webhook deployment**.
// For example, `cloudflare` may be used as the name of a solver.
func (c *k8sDNSProviderSolver) Name() string {
	return "k8s-dns"
}

// Present is responsible for actually presenting the DNS record with the
// DNS provider.
// This method should tolerate being called multiple times with the same value.
// cert-manager itself will later perform a self check to ensure that the
// solver has correctly configured the DNS provider.
func (c *k8sDNSProviderSolver) Present(ch *v1alpha1.ChallengeRequest) error {
	cfg, err := loadConfig(ch)
	if err != nil {
		return err
	}
	c.log.Info("decoded configuration", "config", cfg)
	c.log.Info("create TXT record", "id", ch.UID, "name", ch.ResolvedFQDN, "target", ch.Key)

	a := map[string]string{
		AnnotationKey: ch.ResolvedFQDN,
	}
	r := dnsv1alpha1.DNSRecord{
		TypeMeta: v1.TypeMeta{
			APIVersion: dnsv1alpha1.GroupVersion.String(),
			Kind:       "DNSRecord",
		},
		ObjectMeta: v1.ObjectMeta{
			Name:        recordName(ch),
			Namespace:   cfg.Namespace,
			Annotations: a,
		},
		Spec: dnsv1alpha1.DNSRecordSpec{
			TXT: &dnsv1alpha1.TXTRecord{
				Name:    ch.ResolvedFQDN,
				Ttl:     30,
				Targets: []string{ch.Key},
			},
		},
	}
	b, err := json.Marshal(r)
	if err != nil {
		return err
	}
	// We cannot use controller-runtime client as this would introduce dependency issues
	return c.client.RESTClient().
		Post().
		AbsPath("apis", dnsv1alpha1.GroupVersion.Group, dnsv1alpha1.GroupVersion.Version, "namespaces", r.Namespace, "dnsrecords").
		Body(b).
		Do().
		Error()
}

// CleanUp should delete the relevant TXT record from the DNS provider console.
// If multiple TXT records exist with the same record name (e.g.
// _acme-challenge.example.com) then **only** the record with the same `key`
// value provided on the ChallengeRequest should be cleaned up.
// This is in order to facilitate multiple DNS validations for the same domain
// concurrently.
func (c *k8sDNSProviderSolver) CleanUp(ch *v1alpha1.ChallengeRequest) error {
	cfg, err := loadConfig(ch)
	if err != nil {
		return err
	}
	c.log.Info("decoded configuration", "config", cfg)
	c.log.Info("cleaning TXT record", "id", ch.UID, "name", ch.ResolvedFQDN, "target", ch.Key)
	a := map[string]string{
		AnnotationKey: ch.ResolvedFQDN,
	}
	r := dnsv1alpha1.DNSRecord{
		TypeMeta: v1.TypeMeta{
			APIVersion: dnsv1alpha1.GroupVersion.String(),
			Kind:       "DNSRecord",
		},
		ObjectMeta: v1.ObjectMeta{
			Name:        recordName(ch),
			Namespace:   cfg.Namespace,
			Annotations: a,
		},
		Spec: dnsv1alpha1.DNSRecordSpec{
			TXT: &dnsv1alpha1.TXTRecord{
				Name:    ch.ResolvedFQDN,
				Ttl:     30,
				Targets: []string{ch.Key},
			},
		},
	}
	return c.client.RESTClient().
		Delete().
		AbsPath("apis", dnsv1alpha1.GroupVersion.Group, dnsv1alpha1.GroupVersion.Version, "namespaces", r.Namespace, "dnsrecords", r.Name).
		Do().
		Error()
}

// Initialize will be called when the webhook first starts.
// This method can be used to instantiate the webhook, i.e. initialising
// connections or warming up caches.
// Typically, the kubeClientConfig parameter is used to build a Kubernetes
// client that can be used to fetch resources from the Kubernetes API, e.g.
// Secret resources containing credentials used to authenticate with DNS
// provider accounts.
// The stopCh can be used to handle early termination of the webhook, in cases
// where a SIGTERM or similar signal is sent to the webhook process.
func (c *k8sDNSProviderSolver) Initialize(kubeClientConfig *rest.Config, stopCh <-chan struct{}) error {
	var err error
	if c.client, err = kubernetes.NewForConfig(kubeClientConfig); err != nil {
		return err
	}
	c.log = logrusr.NewLogger(logrus.New())
	return nil
}

// loadConfig is a small helper function that decodes JSON configuration into
// the typed config struct.
func loadConfig(ch *v1alpha1.ChallengeRequest) (k8sDNSProviderConfig, error) {
	cfg := k8sDNSProviderConfig{
		Namespace: ch.ResourceNamespace,
	}
	// handle the 'base case' where no configuration has been provided
	if ch.Config == nil {
		return cfg, nil
	}
	if err := json.Unmarshal(ch.Config.Raw, &cfg); err != nil {
		return cfg, fmt.Errorf("error decoding solver config: %v", err)
	}
	if cfg.Namespace == "" {
		cfg.Namespace = ch.ResourceNamespace
	}
	return cfg, nil
}

var regx = regexp.MustCompile("[a-z0-9]([-a-z0-9]*[a-z0-9])?(\\.[a-z0-9]([-a-z0-9]*[a-z0-9])?)*")

func recordName(ch *v1alpha1.ChallengeRequest) string {
	name := strings.ToLower(base64.RawStdEncoding.EncodeToString([]byte(fmt.Sprintf("%s-%s", ch.ResolvedFQDN, ch.Key))))
	name = strings.Replace(name, "/", "-", -1)
	name = strings.Replace(name, "+", ".", -1)
	if !regx.MatchString(name) {
		name = strings.ToLower(base64.RawStdEncoding.EncodeToString([]byte(ch.Key)))
	}
	return name
}

func normalizeRecordName(n string) string {
	name := strings.TrimSuffix(n, ".")
	name = strings.Replace(name, ".", "-", -1)
	name = strings.Replace(name, "_", "", -1)
	name = strings.Replace(name, "*", "wildcard", -1)
	return name
}
