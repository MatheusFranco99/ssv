module github.com/MatheusFranco99/ssv

go 1.20

require (
	github.com/MatheusFranco99/ssv-spec-AleaBFT v0.3.9
	github.com/aquasecurity/table v1.8.0
	github.com/attestantio/go-eth2-client v0.15.2
	github.com/bloxapp/eth2-key-manager v1.2.0
	github.com/bloxapp/ssv v0.4.0
	github.com/btcsuite/btcd/btcec/v2 v2.2.1
	github.com/dgraph-io/badger/v3 v3.2103.2
	github.com/ethereum/go-ethereum v1.10.23
	github.com/ferranbt/fastssz v0.1.2
	github.com/golang/mock v1.6.0
	github.com/google/uuid v1.3.0
	github.com/gorilla/websocket v1.5.0
	github.com/grpc-ecosystem/go-grpc-middleware v1.2.2
	github.com/grpc-ecosystem/go-grpc-prometheus v1.2.0
	github.com/hashicorp/go-multierror v1.1.1
	github.com/herumi/bls-eth-go-binary v1.28.1
	github.com/ilyakaznacheev/cleanenv v1.2.5
	github.com/ipfs/go-log v1.0.5
	github.com/libp2p/go-libp2p v0.24.2
	github.com/libp2p/go-libp2p-kad-dht v0.20.0
	github.com/libp2p/go-libp2p-pubsub v0.8.2
	github.com/multiformats/go-multiaddr v0.8.0
	github.com/multiformats/go-multistream v0.3.3
	github.com/patrickmn/go-cache v2.1.0+incompatible
	github.com/pkg/errors v0.9.1
	github.com/prometheus/client_golang v1.14.0
	github.com/prysmaticlabs/go-bitfield v0.0.0-20210809151128-385d8c5e3fb7
	github.com/prysmaticlabs/prysm v1.4.4
	github.com/rs/zerolog v1.26.1
	github.com/spf13/cobra v1.5.0
	github.com/stretchr/testify v1.8.1
	github.com/wealdtech/go-eth2-util v1.6.3
	go.opencensus.io v0.24.0
	go.uber.org/zap v1.24.0
	golang.org/x/mod v0.7.0
	google.golang.org/grpc v1.40.0
	gopkg.in/yaml.v3 v3.0.1
)

replace github.com/prysmaticlabs/prysm => github.com/prysmaticlabs/prysm v1.4.2-0.20211101172615-63308239d94f

replace github.com/google/flatbuffers => github.com/google/flatbuffers v1.11.0

replace github.com/dgraph-io/ristretto => github.com/dgraph-io/ristretto v0.1.1-0.20211108053508-297c39e6640f
