module miniflux.app

// +heroku goVersion go1.19

require (
	github.com/PuerkitoBio/goquery v1.8.0
	github.com/coreos/go-oidc v2.2.1+incompatible
	github.com/duo-labs/webauthn v0.0.0-20220815211337-00c9fb5711f5
	github.com/go-telegram-bot-api/telegram-bot-api v4.6.4+incompatible
	github.com/gorilla/mux v1.8.0
	github.com/lib/pq v1.10.7
	github.com/matrix-org/gomatrix v0.0.0-20220926102614-ceba4d9f7530
	github.com/prometheus/client_golang v1.14.0
	github.com/rylans/getlang v0.0.0-20201227074721-9e7f44ff8aa0
	github.com/tdewolff/minify/v2 v2.12.4
	github.com/yuin/goldmark v1.5.3
	golang.org/x/crypto v0.3.0
	golang.org/x/net v0.2.0
	golang.org/x/oauth2 v0.2.0
	golang.org/x/term v0.2.0
	mvdan.cc/xurls/v2 v2.4.0
)

require (
	cloud.google.com/go/compute v1.7.0 // indirect
	github.com/andybalholm/cascadia v1.3.1 // indirect
	github.com/beorn7/perks v1.0.1 // indirect
	github.com/bgentry/speakeasy v0.1.0 // indirect
	github.com/census-instrumentation/opencensus-proto v0.3.0 // indirect
	github.com/cespare/xxhash/v2 v2.1.2 // indirect
	github.com/cloudflare/cfssl v1.6.1 // indirect
	github.com/cncf/udpa/go v0.0.0-20210930031921-04548b0d99d4 // indirect
	github.com/cncf/xds/go v0.0.0-20211011173535-cb28da3451f1 // indirect
	github.com/coreos/go-semver v0.3.0 // indirect
	github.com/coreos/go-systemd/v22 v22.3.2 // indirect
	github.com/cpuguy83/go-md2man/v2 v2.0.0 // indirect
	github.com/dustin/go-humanize v1.0.0 // indirect
	github.com/envoyproxy/go-control-plane v0.10.2-0.20220325020618-49ff273808a1 // indirect
	github.com/envoyproxy/protoc-gen-validate v0.6.1 // indirect
	github.com/form3tech-oss/jwt-go v3.2.3+incompatible // indirect
	github.com/fullstorydev/grpcurl v1.8.1 // indirect
	github.com/fxamacker/cbor/v2 v2.4.0 // indirect
	github.com/gogo/protobuf v1.3.2 // indirect
	github.com/golang-jwt/jwt/v4 v4.1.0 // indirect
	github.com/golang/groupcache v0.0.0-20210331224755-41bb18bfe9da // indirect
	github.com/golang/mock v1.6.0 // indirect
	github.com/golang/protobuf v1.5.2 // indirect
	github.com/google/btree v1.0.1 // indirect
	github.com/google/certificate-transparency-go v1.1.2-0.20210511102531-373a877eec92 // indirect
	github.com/google/uuid v1.3.0 // indirect
	github.com/gorilla/websocket v1.4.2 // indirect
	github.com/grpc-ecosystem/go-grpc-middleware v1.3.0 // indirect
	github.com/grpc-ecosystem/go-grpc-prometheus v1.2.0 // indirect
	github.com/grpc-ecosystem/grpc-gateway v1.16.0 // indirect
	github.com/inconshreveable/mousetrap v1.0.0 // indirect
	github.com/jhump/protoreflect v1.8.2 // indirect
	github.com/jonboulle/clockwork v0.2.2 // indirect
	github.com/json-iterator/go v1.1.12 // indirect
	github.com/mattn/go-runewidth v0.0.12 // indirect
	github.com/matttproud/golang_protobuf_extensions v1.0.4 // indirect
	github.com/mitchellh/mapstructure v1.1.2 // indirect
	github.com/modern-go/concurrent v0.0.0-20180306012644-bacd9c7ef1dd // indirect
	github.com/modern-go/reflect2 v1.0.2 // indirect
	github.com/olekukonko/tablewriter v0.0.5 // indirect
	github.com/pquerna/cachecontrol v0.1.0 // indirect
	github.com/prometheus/client_model v0.3.0 // indirect
	github.com/prometheus/common v0.37.0 // indirect
	github.com/prometheus/procfs v0.8.0 // indirect
	github.com/rivo/uniseg v0.2.0 // indirect
	github.com/russross/blackfriday/v2 v2.1.0 // indirect
	github.com/sirupsen/logrus v1.8.1 // indirect
	github.com/soheilhy/cmux v0.1.5 // indirect
	github.com/spf13/cobra v1.1.3 // indirect
	github.com/spf13/pflag v1.0.5 // indirect
	github.com/tdewolff/parse/v2 v2.6.4 // indirect
	github.com/technoweenie/multipartstreamer v1.0.1 // indirect
	github.com/tmc/grpc-websocket-proxy v0.0.0-20201229170055-e5319fda7802 // indirect
	github.com/urfave/cli v1.22.5 // indirect
	github.com/x448/float16 v0.8.4 // indirect
	github.com/xiang90/probing v0.0.0-20190116061207-43a291ad63a2 // indirect
	go.etcd.io/bbolt v1.3.5 // indirect
	go.etcd.io/etcd/api/v3 v3.5.0-alpha.0 // indirect
	go.etcd.io/etcd/client/v2 v2.305.0-alpha.0 // indirect
	go.etcd.io/etcd/client/v3 v3.5.0-alpha.0 // indirect
	go.etcd.io/etcd/etcdctl/v3 v3.5.0-alpha.0 // indirect
	go.etcd.io/etcd/pkg/v3 v3.5.0-alpha.0 // indirect
	go.etcd.io/etcd/raft/v3 v3.5.0-alpha.0 // indirect
	go.etcd.io/etcd/server/v3 v3.5.0-alpha.0 // indirect
	go.etcd.io/etcd/tests/v3 v3.5.0-alpha.0 // indirect
	go.etcd.io/etcd/v3 v3.5.0-alpha.0 // indirect
	go.uber.org/atomic v1.7.0 // indirect
	go.uber.org/multierr v1.7.0 // indirect
	go.uber.org/zap v1.16.0 // indirect
	golang.org/x/mod v0.6.0-dev.0.20220419223038-86c51ed26bb4 // indirect
	golang.org/x/sys v0.2.0 // indirect
	golang.org/x/text v0.4.0 // indirect
	golang.org/x/time v0.0.0-20210220033141-f8bda1e9f3ba // indirect
	golang.org/x/tools v0.1.12 // indirect
	google.golang.org/appengine v1.6.7 // indirect
	google.golang.org/genproto v0.0.0-20220616135557-88e70c0c3a90 // indirect
	google.golang.org/grpc v1.47.0 // indirect
	google.golang.org/protobuf v1.28.1 // indirect
	gopkg.in/cheggaaa/pb.v1 v1.0.28 // indirect
	gopkg.in/square/go-jose.v2 v2.6.0 // indirect
	gopkg.in/yaml.v2 v2.4.0 // indirect
	sigs.k8s.io/yaml v1.2.0 // indirect
)

go 1.19
