IMAGE_NAME := "registry.gitlab.com/linka-cloud/k8s/cert-manager-webhook-k8s-dns"
IMAGE_TAG := "latest"

OUT := $(shell pwd)/_out
BIN := $(OUT)/kubebuilder/bin
KUBECTL := $(BIN)/kubectl
KIND := $(BIN)/kind

$(shell mkdir -p "$(OUT)")

TEST_DATA := testdata/k8s-dns
KIND_NAME := cert-manager-e2e

verify:
	@./scripts/fetch-test-binaries.sh
	$(KIND) get clusters | grep $(KIND_NAME) || $(KIND) create cluster --config=testdata/k8s-dns/kind.yaml --name $(KIND_NAME)
	$(KIND) export kubeconfig --name $(KIND_NAME)
	$(KUBECTL) apply -f $(TEST_DATA)/dns.linka.cloud_dnsrecord.yaml
	$(KUBECTL) apply -f $(TEST_DATA)/k8s-dns.yml --wait
	@TEST_ZONE_NAME=example.com. go test -v .
	$(KIND) delete cluster --name $(KIND_NAME)

build:
	docker build -t "$(IMAGE_NAME):$(IMAGE_TAG)" .

push:
	docker push $(IMAGE_NAME):$(IMAGE_TAG)

docker: build push

.PHONY: rendered-manifest.yaml
rendered-manifest.yaml:
	helm template \
	    cert-manager-webhook-k8s-dns \
        --set image.repository=$(IMAGE_NAME) \
        --set image.tag=$(IMAGE_TAG) \
        deploy/cert-manager-webhook-k8s-dns > "$(OUT)/rendered-manifest.yaml"
