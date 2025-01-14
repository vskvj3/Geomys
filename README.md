# Geomys
A Distributed In-memory Cache written in Go

## Suggested Structure and TODO
```
redis-clone/
├── cmd/
│   ├── server/
│   │   └── main.go       # Entry point for the server
│   ├── client/
│   │   └── main.go       # Entry point for the client
├── internal/
│   ├── core/
│   │   ├── database.go   # Core database implementation
│   │   ├── commands.go   # Command handlers
│   │   ├── replication.go # Replication logic
│   │   └── persistence.go # Persistence layer
│   ├── network/
│   │   ├── server.go     # Network server implementation
│   │   ├── client.go     # Network client implementation
│   │   └── protocol.go   # Serialization and command protocol
│   ├── cluster/
│   │   ├── shard.go      # Sharding logic
│   │   └── membership.go # Cluster membership and discovery
│   ├── utils/
│   │   ├── logger.go     # Logging utilities
│   │   └── config.go     # Configuration parsing
├── pkg/
│   └── api/              # Public APIs for extensions or clients
├── tests/
│   ├── integration/
│   │   └── integration_test.go # Integration tests for the database
│   ├── unit/
│   │   └── unit_test.go        # Unit tests for individual components
├── Dockerfile            # Dockerfile for containerization
├── go.mod                # Go module file
└── README.md             # Project description and usage
```