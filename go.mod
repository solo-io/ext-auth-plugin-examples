module github.com/solo-io/ext-auth-plugin-examples

go 1.13

require (
	cloud.google.com/go/pubsub v1.1.0 // indirect
	github.com/Microsoft/go-winio v0.4.14 // indirect
	github.com/OneOfOne/xxhash v1.2.5 // indirect
	github.com/armon/go-metrics v0.3.0 // indirect
	github.com/avast/retry-go v2.4.3+incompatible // indirect
	github.com/bxcodec/faker v2.0.1+incompatible // indirect
	github.com/cespare/xxhash/v2 v2.1.1 // indirect
	github.com/chai2010/gettext-go v0.0.0-20170215093142-bf70f2a70fb1 // indirect
	github.com/docker/cli v0.0.0-20191017083524-a8ff7f821017 // indirect
	github.com/docker/docker-credential-helpers v0.6.3 // indirect
	github.com/elazarl/goproxy v0.0.0-20190421051319-9d40249d3c2f // indirect
	github.com/envoyproxy/go-control-plane v0.9.1
	github.com/ghodss/yaml v1.0.1-0.20190212211648-25d852aebe32 // indirect
	github.com/go-openapi/loads v0.19.4 // indirect
	github.com/go-openapi/runtime v0.19.5 // indirect
	github.com/google/martian v2.1.1-0.20190517191504-25dcb96d9e51+incompatible // indirect
	github.com/goph/emperror v0.17.2 // indirect
	github.com/gophercloud/gophercloud v0.6.0 // indirect
	github.com/gorilla/handlers v1.4.2 // indirect
	github.com/gorilla/mux v1.7.3 // indirect
	github.com/grpc-ecosystem/go-grpc-middleware v1.0.1-0.20190118093823-f849b5445de4 // indirect
	github.com/grpc-ecosystem/grpc-gateway v1.9.2 // indirect
	github.com/hashicorp/go-immutable-radix v1.1.0 // indirect
	github.com/hashicorp/go-msgpack v0.5.5 // indirect
	github.com/hashicorp/go-retryablehttp v0.6.2 // indirect
	github.com/hashicorp/go-rootcerts v1.0.1 // indirect
	github.com/hashicorp/go-sockaddr v1.0.2 // indirect
	github.com/hashicorp/go-uuid v1.0.2-0.20191001231223-f32f5fe8d6a8 // indirect
	github.com/hashicorp/memberlist v0.1.4 // indirect
	github.com/hashicorp/serf v0.8.3 // indirect
	github.com/konsorten/go-windows-terminal-sequences v1.0.2 // indirect
	github.com/mattn/go-colorable v0.1.2 // indirect
	github.com/mattn/go-runewidth v0.0.6 // indirect
	github.com/miekg/dns v1.1.15 // indirect
	github.com/mitchellh/reflectwalk v1.0.1 // indirect
	github.com/morikuni/aec v1.0.0 // indirect
	github.com/onsi/ginkgo v1.10.1
	github.com/onsi/gomega v1.7.1
	github.com/opencontainers/runc v1.0.0-rc2.0.20190611121236-6cc515888830 // indirect
	github.com/pelletier/go-toml v1.4.0 // indirect
	github.com/pkg/errors v0.8.1
	github.com/pquerna/cachecontrol v0.0.0-20180517163645-1555304b9b35 // indirect
	github.com/prometheus/tsdb v0.10.0 // indirect
	github.com/solo-io/ext-auth-plugins v0.1.1
	github.com/solo-io/go-utils v0.11.5
	github.com/solo-io/solo-kit v0.11.13 // indirect
	github.com/spf13/jwalterweatherman v1.1.0 // indirect
	github.com/spf13/viper v1.5.0 // indirect
	github.com/ugorji/go v1.1.5-pre // indirect
	github.com/xeipuuv/gojsonpointer v0.0.0-20190905194746-02993c407bfb // indirect
	github.com/xeipuuv/gojsonschema v1.2.0 // indirect
	go.opencensus.io v0.22.2 // indirect
	go.uber.org/multierr v1.4.0 // indirect
	go.uber.org/zap v1.13.0
	golang.org/x/crypto v0.0.0-20191206172530-e9b2fee46413 // indirect
	golang.org/x/lint v0.0.0-20191125180803-fdd1cda4f05f // indirect
	golang.org/x/mod v0.1.1-0.20191209134235-331c550502dd // indirect
	golang.org/x/sys v0.0.0-20191210023423-ac6580df4449 // indirect
	golang.org/x/tools v0.0.0-20191212051200-825cb0626375 // indirect
	google.golang.org/genproto v0.0.0-20191115221424-83cc0476cb11 // indirect
	google.golang.org/grpc v1.25.1 // indirect
	gopkg.in/AlecAivazis/survey.v1 v1.8.7 // indirect
	gopkg.in/square/go-jose.v2 v2.3.1 // indirect
)

replace (
	github.com/Azure/go-autorest => github.com/Azure/go-autorest v13.0.0+incompatible
	github.com/Sirupsen/logrus => github.com/sirupsen/logrus v1.0.5
	github.com/docker/docker => github.com/moby/moby v0.7.3-0.20190826074503-38ab9da00309
	github.com/hashicorp/consul => github.com/hashicorp/consul v1.5.1
	github.com/hashicorp/consul/api => github.com/hashicorp/consul/api v1.1.0
	github.com/hashicorp/vault => github.com/hashicorp/vault v1.3.0
	github.com/hashicorp/vault/api => github.com/hashicorp/vault/api v1.0.5-0.20191108163347-bdd38fca2cff
	k8s.io/api => k8s.io/api v0.0.0-20191004120104-195af9ec3521
	k8s.io/apiextensions-apiserver => k8s.io/apiextensions-apiserver v0.0.0-20191204090712-e0e829f17bab
	k8s.io/apimachinery => k8s.io/apimachinery v0.0.0-20191028221656-72ed19daf4bb
	k8s.io/apiserver => k8s.io/apiserver v0.0.0-20191109104512-b243870e034b
	k8s.io/cli-runtime => k8s.io/cli-runtime v0.0.0-20191004123735-6bff60de4370
	k8s.io/client-go => k8s.io/client-go v0.0.0-20191016111102-bec269661e48
	k8s.io/cloud-provider => k8s.io/cloud-provider v0.0.0-20191004125000-f72359dfc58e
	k8s.io/cluster-bootstrap => k8s.io/cluster-bootstrap v0.0.0-20191004124811-493ca03acbc1
	k8s.io/code-generator => k8s.io/code-generator v0.0.0-20191004115455-8e001e5d1894
	k8s.io/component-base => k8s.io/component-base v0.0.0-20191004121439-41066ddd0b23
	k8s.io/cri-api => k8s.io/cri-api v0.0.0-20190828162817-608eb1dad4ac
	k8s.io/csi-translation-lib => k8s.io/csi-translation-lib v0.0.0-20191004125145-7118cc13aa0a
	k8s.io/gengo => k8s.io/gengo v0.0.0-20190822140433-26a664648505
	k8s.io/heapster => k8s.io/heapster v1.2.0-beta.1
	k8s.io/klog => github.com/stefanprodan/klog v0.0.0-20190418165334-9cbb78b20423
	k8s.io/kube-aggregator => k8s.io/kube-aggregator v0.0.0-20191104231939-9e18019dec40
	k8s.io/kube-controller-manager => k8s.io/kube-controller-manager v0.0.0-20191004124629-b9859bb1ce71
	k8s.io/kube-openapi => k8s.io/kube-openapi v0.0.0-20190816220812-743ec37842bf
	k8s.io/kube-proxy => k8s.io/kube-proxy v0.0.0-20191004124112-c4ee2f9e1e0a
	k8s.io/kube-scheduler => k8s.io/kube-scheduler v0.0.0-20191004124444-89f3bbd82341
	k8s.io/kubectl => k8s.io/kubectl v0.0.0-20191004125858-14647fd13a8b
	k8s.io/kubelet => k8s.io/kubelet v0.0.0-20191004124258-ac1ea479bd3a
	k8s.io/legacy-cloud-providers => k8s.io/legacy-cloud-providers v0.0.0-20191203122058-2ae7e9ca8470
	k8s.io/metrics => k8s.io/metrics v0.0.0-20191004123543-798934cf5e10
	k8s.io/node-api => k8s.io/node-api v0.0.0-20191004125527-f5592a7bd6b6
	k8s.io/repo-infra => k8s.io/repo-infra v0.0.0-20181204233714-00fe14e3d1a3
	k8s.io/sample-apiserver => k8s.io/sample-apiserver v0.0.0-20191028231949-ceef03da3009
	k8s.io/sample-cli-plugin => k8s.io/sample-cli-plugin v0.0.0-20191004123926-88de2937c61b
	k8s.io/sample-controller => k8s.io/sample-controller v0.0.0-20191004122958-d040c2be0d0b
	k8s.io/utils => k8s.io/utils v0.0.0-20190801114015-581e00157fb1
)
