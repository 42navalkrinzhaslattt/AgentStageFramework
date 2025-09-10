module presidential-simulator

go 1.25.1

require github.com/emergent-world-engine/backend v0.0.0

replace github.com/emergent-world-engine/backend => ../

require (
	github.com/cespare/xxhash/v2 v2.3.0 // indirect
	github.com/dgryski/go-rendezvous v0.0.0-20200823014737-9f7001d12a5f // indirect
	github.com/redis/go-redis/v9 v9.13.0 // indirect
)
