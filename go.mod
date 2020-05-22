module github.com/solo-io/ext-auth-plugin-examples

go 1.14

require (
	// Merged 'require' section of the Gloo depenencies and your go.mod file:
	cloud.google.com/go v0.46.3
	cloud.google.com/go/pubsub v1.1.0
	github.com/Azure/azure-sdk-for-go v35.0.0+incompatible
	github.com/Azure/go-autorest/autorest v0.9.3
	github.com/Azure/go-autorest/autorest/adal v0.8.1
	github.com/Azure/go-autorest/autorest/date v0.2.0
	github.com/Azure/go-autorest/autorest/mocks v0.3.0
	github.com/Masterminds/semver/v3 v3.0.3
	github.com/Masterminds/sprig/v3 v3.0.2
	github.com/Masterminds/vcs v1.13.1
	github.com/Microsoft/go-winio v0.4.15-0.20190919025122-fc70bd9a86b5
	github.com/Microsoft/hcsshim v0.8.7
	github.com/OneOfOne/xxhash v1.2.7
	github.com/armon/go-metrics v0.3.0
	github.com/asaskevich/govalidator v0.0.0-20200108200545-475eaeb16496
	github.com/avast/retry-go v2.4.3+incompatible
	github.com/aws/aws-sdk-go v1.26.2
	github.com/cespare/xxhash/v2 v2.1.1
	github.com/chai2010/gettext-go v0.0.0-20170215093142-bf70f2a70fb1
	github.com/containerd/containerd v1.3.2
	github.com/containerd/continuity v0.0.0-20200107194136-26c1120b8d41
	github.com/containerd/typeurl v0.0.0-20190228175220-2a93cfde8c20
	github.com/coreos/etcd v3.3.15+incompatible
	github.com/deislabs/oras v0.8.1
	github.com/docker/cli v0.0.0-20200130152716-5d0cf8839492
	github.com/docker/docker-credential-helpers v0.6.3
	github.com/elazarl/goproxy v0.0.0-20190421051319-9d40249d3c2f
	github.com/envoyproxy/go-control-plane v0.9.6-0.20200401235947-be7fefdaf0df
	github.com/ghodss/yaml v1.0.1-0.20190212211648-25d852aebe32
	github.com/go-openapi/runtime v0.19.5
	github.com/godbus/dbus v4.1.0+incompatible
	github.com/golang/mock v1.4.3
	github.com/golang/protobuf v1.3.5
	github.com/google/go-cmp v0.4.0
	github.com/google/martian v2.1.1-0.20190517191504-25dcb96d9e51+incompatible
	github.com/google/pprof v0.0.0-20190515194954-54271f7e092f
	github.com/googleapis/gax-go/v2 v2.0.5
	github.com/goph/emperror v0.17.2
	github.com/gophercloud/gophercloud v0.6.0
	github.com/gorilla/handlers v1.4.2
	github.com/gorilla/mux v1.7.3
	github.com/gorilla/websocket v1.4.1
	github.com/gosuri/uitable v0.0.4
	github.com/gregjones/httpcache v0.0.0-20190611155906-901d90724c79
	github.com/hashicorp/consul/sdk v0.3.0
	github.com/hashicorp/go-hclog v0.10.0
	github.com/hashicorp/go-immutable-radix v1.1.0
	github.com/hashicorp/go-msgpack v0.5.5
	github.com/hashicorp/go-retryablehttp v0.6.4
	github.com/hashicorp/go-rootcerts v1.0.1
	github.com/hashicorp/go-sockaddr v1.0.2
	github.com/hashicorp/go-uuid v1.0.2-0.20191001231223-f32f5fe8d6a8
	github.com/hashicorp/memberlist v0.1.5
	github.com/hashicorp/serf v0.8.5
	github.com/jmespath/go-jmespath v0.0.0-20180206201540-c2b33e8439af
	github.com/karrick/godirwalk v1.14.1
	github.com/konsorten/go-windows-terminal-sequences v1.0.2
	github.com/kylelemons/godebug v1.1.0
	github.com/magiconair/properties v1.8.1
	github.com/mattn/go-colorable v0.1.6
	github.com/mattn/go-runewidth v0.0.7
	github.com/mattn/go-shellwords v1.0.9
	github.com/miekg/dns v1.1.15
	github.com/mitchellh/reflectwalk v1.0.1
	github.com/morikuni/aec v1.0.0
	github.com/olekukonko/tablewriter v0.0.4
	github.com/onsi/ginkgo v1.11.0
	github.com/onsi/gomega v1.8.1
	github.com/opencontainers/runc v1.0.0-rc9
	github.com/opencontainers/runtime-spec v1.0.0
	github.com/pascaldekloe/goe v0.1.0
	github.com/pelletier/go-toml v1.4.0
	github.com/phayes/freeport v0.0.0-20180830031419-95f893ade6f2
	github.com/pkg/errors v0.9.1
	github.com/pquerna/cachecontrol v0.0.0-20180517163645-1555304b9b35
	github.com/prometheus/tsdb v0.10.0
	github.com/rcrowley/go-metrics v0.0.0-20190706150252-9beb055b7962
	github.com/rogpeppe/go-internal v1.5.2
	github.com/ryanuber/columnize v2.1.0+incompatible
	github.com/solo-io/ext-auth-plugins v0.1.2
	github.com/solo-io/go-utils v0.14.2
	github.com/spf13/jwalterweatherman v1.1.0
	github.com/spf13/viper v1.5.0
	github.com/syndtr/gocapability v0.0.0-20180916011248-d98352740cb2
	github.com/tmc/grpc-websocket-proxy v0.0.0-20190109142713-0ad062ec5ee5
	github.com/ugorji/go/codec v1.1.5-pre
	github.com/xeipuuv/gojsonpointer v0.0.0-20190905194746-02993c407bfb
	github.com/xeipuuv/gojsonschema v1.2.0
	github.com/yuin/goldmark v1.1.27
	go.opencensus.io v0.22.2
	go.uber.org/multierr v1.4.0
	go.uber.org/zap v1.13.0
	golang.org/x/crypto v0.0.0-20200311171314-f7b00557c8c4
	golang.org/x/exp v0.0.0-20191030013958-a1ab85dbe136
	golang.org/x/lint v0.0.0-20191125180803-fdd1cda4f05f
	golang.org/x/mobile v0.0.0-20190719004257-d2bd2a29d028
	golang.org/x/mod v0.2.0
	golang.org/x/net v0.0.0-20200301022130-244492dfa37a
	golang.org/x/sys v0.0.0-20200302150141-5c8b2ff67527
	golang.org/x/tools v0.0.0-20200423204450-38a97e00a8a1
	golang.org/x/xerrors v0.0.0-20191204190536-9bdfabe68543
	google.golang.org/api v0.14.0
	google.golang.org/genproto v0.0.0-20200309141739-5b75447e413d
	google.golang.org/grpc v1.28.0-pre.0.20200226185027-6cd03861bfd2
	gopkg.in/AlecAivazis/survey.v1 v1.8.7
	gopkg.in/square/go-jose.v2 v2.3.1
	gopkg.in/yaml.v2 v2.2.8
	helm.sh/helm/v3 v3.1.2
	k8s.io/kubernetes v1.17.1
	sigs.k8s.io/controller-runtime v0.5.1
)

replace (
	// Merged 'replace' section of the Gloo depenencies and your go.mod file:
	github.com/Azure/go-autorest => github.com/Azure/go-autorest v13.0.0+incompatible
	github.com/Sirupsen/logrus => github.com/sirupsen/logrus v1.0.5
	github.com/docker/docker => github.com/moby/moby v0.7.3-0.20190826074503-38ab9da00309
	github.com/hashicorp/consul => github.com/hashicorp/consul v1.5.1
	github.com/hashicorp/consul/api => github.com/hashicorp/consul/api v1.1.0
	github.com/hashicorp/vault => github.com/hashicorp/vault v1.3.0
	github.com/hashicorp/vault/api => github.com/hashicorp/vault/api v1.0.5-0.20191108163347-bdd38fca2cff
	github.com/pseudomuto/protoc-gen-doc => github.com/pseudomuto/protoc-gen-doc v1.0.0
	github.com/sclevine/agouti => github.com/yuval-k/agouti v0.0.0-20190109124522-0e71d6bad483
	k8s.io/api => k8s.io/api v0.17.1
	k8s.io/apiextensions-apiserver => k8s.io/apiextensions-apiserver v0.17.1
	k8s.io/apimachinery => k8s.io/apimachinery v0.17.1
	k8s.io/apiserver => k8s.io/apiserver v0.17.1
	k8s.io/cli-runtime => k8s.io/cli-runtime v0.17.1
	k8s.io/client-go => k8s.io/client-go v0.17.1
	k8s.io/cloud-provider => k8s.io/cloud-provider v0.17.1
	k8s.io/cluster-bootstrap => k8s.io/cluster-bootstrap v0.17.1
	k8s.io/code-generator => k8s.io/code-generator v0.17.1
	k8s.io/component-base => k8s.io/component-base v0.17.1
	k8s.io/cri-api => k8s.io/cri-api v0.17.1
	k8s.io/csi-translation-lib => k8s.io/csi-translation-lib v0.17.1
	k8s.io/gengo => k8s.io/gengo v0.0.0-20190822140433-26a664648505
	k8s.io/heapster => k8s.io/heapster v1.2.0-beta.1
	k8s.io/klog => github.com/stefanprodan/klog v0.0.0-20190418165334-9cbb78b20423
	k8s.io/kube-aggregator => k8s.io/kube-aggregator v0.17.1
	k8s.io/kube-controller-manager => k8s.io/kube-controller-manager v0.17.1
	k8s.io/kube-openapi => k8s.io/kube-openapi v0.0.0-20190816220812-743ec37842bf
	k8s.io/kube-proxy => k8s.io/kube-proxy v0.17.1
	k8s.io/kube-scheduler => k8s.io/kube-scheduler v0.17.1
	k8s.io/kubectl => k8s.io/kubectl v0.17.1
	k8s.io/kubelet => k8s.io/kubelet v0.17.1
	k8s.io/legacy-cloud-providers => k8s.io/legacy-cloud-providers v0.17.1
	k8s.io/metrics => k8s.io/metrics v0.17.1
	k8s.io/node-api => k8s.io/node-api v0.17.1
	k8s.io/repo-infra => k8s.io/repo-infra v0.0.0-20181204233714-00fe14e3d1a3
	k8s.io/sample-apiserver => k8s.io/sample-apiserver v0.17.1
	k8s.io/sample-cli-plugin => k8s.io/sample-cli-plugin v0.17.1
	k8s.io/sample-controller => k8s.io/sample-controller v0.17.1
	k8s.io/utils => k8s.io/utils v0.0.0-20190801114015-581e00157fb1
)
