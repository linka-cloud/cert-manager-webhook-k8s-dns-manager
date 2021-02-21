# [k8s-dns-manager](https://gitlab.com/linka-cloud/k8s/dns) Webhook for Cert Manager

## Prerequisites

* [cert-manager](https://github.com/jetstack/cert-manager) version 0.11.0 or higher (*tested with 0.12.0*):
    - [Installing on Kubernetes](https://cert-manager.io/docs/installation/kubernetes/#installing-with-helm)
    
* [k8s-dns-manager](https://gitlab.com/linka-cloud/k8s/dns) installed and configured

## Installation
1. Clone this repository:
   ```bash
   $ git clone https://gitlab.com/linka-cloud/k8s/cert-manager-webhook-k8s-dns-manager.git && \
        cd cert-manager-webhook-k8s-dns-manager
   ```
2. Run:
    ```bash
    $ helm install cert-manager-webhook-k8s-dns ./deploy/cert-manager-webhook-k8s-dns
    ```

## How to use it

Here is an example using the [Let's Encrypt staging environment](https://letsencrypt.org/docs/staging-environment/).
To go to the production environment, replace `https://acme-staging-v02.api.letsencrypt.org/directory` with
`https://acme-v02.api.letsencrypt.org/directory`

1. Create a certificate issuer:

    ```yaml
    apiVersion: cert-manager.io/v1alpha2
    kind: Issuer # or ClusterIssuer to have it available in every namespaces
    metadata:
      name: letsencrypt
    spec:
      acme:
        server: https://acme-staging-v02.api.letsencrypt.org/directory
        email: '<YOUR_EMAIL_ADDRESS>'
        privateKeySecretRef:
          name: letsencrypt-account-key
        solvers:
        - dns01:
            webhook:
              groupName: acme.dns.linka.cloud
              solverName: k8s-dns
              config:
                namespace: cert-manager
    ```

2. Issue a certificate:
    
    ```yaml
    apiVersion: cert-manager.io/v1alpha2
    kind: Certificate
    metadata:
      name: example-com
    spec:
      dnsNames:
      - example.com
      - *.example.com
      issuerRef:
        name: letsencrypt
      secretName: example-com-tls
    ```

### Running the test suite

All DNS providers **must** run the DNS01 provider conformance testing suite,
else they will have undetermined behaviour when used with cert-manager.

The tests require Docker to be installed on the local machine, and 
[Kind](https://kind.sigs.k8s.io/docs/user/quick-start/), which is
downloaded when the tests are launched.

You can run the test suite with:

```bash
$ make verify
```

**The tests may fail at the first run, but should pass the next time.**
