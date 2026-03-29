module github.com/zoobz-io/argus

go 1.25.0

toolchain go1.25.3

require (
	cloud.google.com/go/auth v0.18.2
	cloud.google.com/go/storage v1.61.3
	github.com/coreos/go-oidc/v3 v3.17.0
	github.com/go-jose/go-jose/v4 v4.1.3
	github.com/google/uuid v1.6.0
	github.com/jmoiron/sqlx v1.4.0
	github.com/ledongthuc/pdf v0.0.0-20250511090121-5959a4027728
	github.com/minio/minio-go/v7 v7.0.99
	github.com/opensearch-project/opensearch-go/v4 v4.3.0
	github.com/redis/go-redis/v9 v9.18.0
	github.com/xuri/excelize/v2 v2.10.1
	github.com/zoobz-io/aperture v1.0.3
	github.com/zoobz-io/argus/proto v0.0.0-20260324042205-707100498d41
	github.com/zoobz-io/astql v1.0.7
	github.com/zoobz-io/capitan v1.0.2
	github.com/zoobz-io/cereal v0.1.2
	github.com/zoobz-io/check v0.0.5
	github.com/zoobz-io/grub v0.1.17
	github.com/zoobz-io/grub/minio v0.1.11
	github.com/zoobz-io/grub/opensearch v0.1.11
	github.com/zoobz-io/grub/postgres v0.1.11
	github.com/zoobz-io/grub/redis v0.1.11
	github.com/zoobz-io/herald v1.0.4
	github.com/zoobz-io/herald/redis v1.0.4
	github.com/zoobz-io/lucene v0.0.4
	github.com/zoobz-io/pipz v1.0.5
	github.com/zoobz-io/rocco v0.1.16
	github.com/zoobz-io/soy/testing v0.0.0-20260326212003-c968b42d4748
	github.com/zoobz-io/sum v0.0.12
	github.com/zoobz-io/vex v0.0.2
	github.com/zoobz-io/zyn v1.0.2
	github.com/zoobz-io/zyn/openai v0.0.0-20260320210919-408cce8f5047
	go.opentelemetry.io/otel v1.42.0
	go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploghttp v0.14.0
	go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetrichttp v1.38.0
	go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp v1.38.0
	go.opentelemetry.io/otel/log v0.14.0
	go.opentelemetry.io/otel/metric v1.42.0
	go.opentelemetry.io/otel/sdk v1.42.0
	go.opentelemetry.io/otel/sdk/log v0.14.0
	go.opentelemetry.io/otel/sdk/metric v1.42.0
	go.opentelemetry.io/otel/trace v1.42.0
	golang.org/x/oauth2 v0.36.0
	google.golang.org/api v0.273.0
	google.golang.org/grpc v1.79.3
)

require (
	cel.dev/expr v0.25.1 // indirect
	cloud.google.com/go v0.123.0 // indirect
	cloud.google.com/go/auth/oauth2adapt v0.2.8 // indirect
	cloud.google.com/go/compute/metadata v0.9.0 // indirect
	cloud.google.com/go/iam v1.5.3 // indirect
	cloud.google.com/go/monitoring v1.24.3 // indirect
	github.com/GoogleCloudPlatform/opentelemetry-operations-go/detectors/gcp v1.30.0 // indirect
	github.com/GoogleCloudPlatform/opentelemetry-operations-go/exporter/metric v0.55.0 // indirect
	github.com/GoogleCloudPlatform/opentelemetry-operations-go/internal/resourcemapping v0.55.0 // indirect
	github.com/cenkalti/backoff/v5 v5.0.3 // indirect
	github.com/cespare/xxhash/v2 v2.3.0 // indirect
	github.com/cncf/xds/go v0.0.0-20251210132809-ee656c7534f5 // indirect
	github.com/dgryski/go-rendezvous v0.0.0-20200823014737-9f7001d12a5f // indirect
	github.com/dustin/go-humanize v1.0.1 // indirect
	github.com/envoyproxy/go-control-plane/envoy v1.36.0 // indirect
	github.com/envoyproxy/protoc-gen-validate v1.3.0 // indirect
	github.com/felixge/httpsnoop v1.0.4 // indirect
	github.com/go-ini/ini v1.67.0 // indirect
	github.com/go-logr/logr v1.4.3 // indirect
	github.com/go-logr/stdr v1.2.2 // indirect
	github.com/google/s2a-go v0.1.9 // indirect
	github.com/googleapis/enterprise-certificate-proxy v0.3.14 // indirect
	github.com/googleapis/gax-go/v2 v2.19.0 // indirect
	github.com/grpc-ecosystem/grpc-gateway/v2 v2.28.0 // indirect
	github.com/klauspost/compress v1.18.2 // indirect
	github.com/klauspost/cpuid/v2 v2.2.11 // indirect
	github.com/klauspost/crc32 v1.3.0 // indirect
	github.com/lib/pq v1.10.9 // indirect
	github.com/minio/crc64nvme v1.1.1 // indirect
	github.com/minio/md5-simd v1.1.2 // indirect
	github.com/philhofer/fwd v1.2.0 // indirect
	github.com/planetscale/vtprotobuf v0.6.1-0.20240319094008-0393e58bdf10 // indirect
	github.com/richardlehane/mscfb v1.0.6 // indirect
	github.com/richardlehane/msoleps v1.0.6 // indirect
	github.com/rs/xid v1.6.0 // indirect
	github.com/spiffe/go-spiffe/v2 v2.6.0 // indirect
	github.com/tiendc/go-deepcopy v1.7.2 // indirect
	github.com/tinylib/msgp v1.6.1 // indirect
	github.com/xuri/efp v0.0.1 // indirect
	github.com/xuri/nfp v0.0.2-0.20250530014748-2ddeb826f9a9 // indirect
	github.com/zoobz-io/atom v1.0.1 // indirect
	github.com/zoobz-io/clockz v1.0.2 // indirect
	github.com/zoobz-io/dbml v1.0.1 // indirect
	github.com/zoobz-io/edamame v1.0.2 // indirect
	github.com/zoobz-io/fig v0.0.3 // indirect
	github.com/zoobz-io/openapi v1.0.2 // indirect
	github.com/zoobz-io/scio v0.0.5 // indirect
	github.com/zoobz-io/sentinel v1.0.4 // indirect
	github.com/zoobz-io/slush v0.0.3 // indirect
	github.com/zoobz-io/soy v1.0.8 // indirect
	github.com/zoobz-io/vecna v0.0.3 // indirect
	go.opentelemetry.io/auto/sdk v1.2.1 // indirect
	go.opentelemetry.io/contrib/detectors/gcp v1.39.0 // indirect
	go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc v0.63.0 // indirect
	go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp v0.63.0 // indirect
	go.opentelemetry.io/otel/exporters/otlp/otlptrace v1.42.0 // indirect
	go.opentelemetry.io/proto/otlp v1.10.0 // indirect
	go.uber.org/atomic v1.11.0 // indirect
	go.yaml.in/yaml/v3 v3.0.4 // indirect
	golang.org/x/crypto v0.49.0 // indirect
	golang.org/x/exp v0.0.0-20260112195511-716be5621a96 // indirect
	golang.org/x/image v0.32.0 // indirect
	golang.org/x/net v0.52.0 // indirect
	golang.org/x/sync v0.20.0 // indirect
	golang.org/x/sys v0.42.0 // indirect
	golang.org/x/text v0.35.0 // indirect
	golang.org/x/time v0.15.0 // indirect
	google.golang.org/genproto v0.0.0-20260316180232-0b37fe3546d5 // indirect
	google.golang.org/genproto/googleapis/api v0.0.0-20260316180232-0b37fe3546d5 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20260319201613-d00831a3d3e7 // indirect
	google.golang.org/protobuf v1.36.11 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

replace github.com/zoobz-io/argus/proto => ./proto
