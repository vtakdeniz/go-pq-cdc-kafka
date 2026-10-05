module github.com/Trendyol/go-pq-cdc-kafka/example/metadata-demo

go 1.25.0

require (
	github.com/Trendyol/go-pq-cdc v1.11.14
	github.com/Trendyol/go-pq-cdc-kafka v0.0.0
	github.com/lib/pq v1.10.9
	github.com/segmentio/kafka-go v0.4.51
)

require (
	github.com/avast/retry-go/v4 v4.6.0 // indirect
	github.com/beorn7/perks v1.0.1 // indirect
	github.com/cespare/xxhash/v2 v2.3.0 // indirect
	github.com/go-playground/errors v3.3.0+incompatible // indirect
	github.com/jackc/pgpassfile v1.0.0 // indirect
	github.com/jackc/pgservicefile v0.0.0-20240606120523-5a60cdf6a761 // indirect
	github.com/jackc/pgx/v5 v5.9.2 // indirect
	github.com/klauspost/compress v1.18.0 // indirect
	github.com/munnerz/goautoneg v0.0.0-20191010083416-a7dc8b61c822 // indirect
	github.com/pierrec/lz4/v4 v4.1.21 // indirect
	github.com/pkg/errors v0.9.1 // indirect
	github.com/prometheus/client_golang v1.20.5 // indirect
	github.com/prometheus/client_model v0.6.1 // indirect
	github.com/prometheus/common v0.60.1 // indirect
	github.com/prometheus/procfs v0.15.1 // indirect
	github.com/xdg-go/pbkdf2 v1.0.0 // indirect
	github.com/xdg-go/scram v1.1.2 // indirect
	github.com/xdg-go/stringprep v1.0.4 // indirect
	golang.org/x/sys v0.35.0 // indirect
	golang.org/x/text v0.29.0 // indirect
	google.golang.org/protobuf v1.36.5 // indirect
	gopkg.in/yaml.v2 v2.4.0 // indirect
)

replace github.com/Trendyol/go-pq-cdc-kafka => ../..

// Temporary: ListenerContext.Xid hosted on cursor/expose-listener-xid-fb7d until upstream
replace github.com/Trendyol/go-pq-cdc => github.com/vtakdeniz/go-pq-cdc-kafka v0.0.0-20261005193026-fd7107ee21f4
