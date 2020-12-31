module go.linka.cloud/k8s/cert-manager-webhook-k8s-dns

go 1.13

require (
	github.com/bombsimon/logrusr v1.0.0
	github.com/go-logr/logr v0.1.0
	github.com/jetstack/cert-manager v0.13.1
	github.com/miekg/dns v0.0.0-20170721150254-0f3adef2e220
	github.com/sirupsen/logrus v1.4.2
	k8s.io/api v0.17.0
	k8s.io/apiextensions-apiserver v0.17.0
	k8s.io/apimachinery v0.17.0
	k8s.io/client-go v0.17.0
	k8s.io/klog v1.0.0
	sigs.k8s.io/controller-runtime v0.3.1-0.20191022174215-ad57a976ffa1
	sigs.k8s.io/testing_frameworks v0.1.1
)
