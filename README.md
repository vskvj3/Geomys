# Geomys 
Geomys is a **distributed in-memory key-value store** that supports **leader-follower replication, persistence, and multi-node clustering**. It ensures **high availability and eventual consistency** across nodes using **gRPC-based data replication**.  

---

## **Table of Contents**  
- [Features](#features)  
- [Architecture Overview](#architecture-overview)  
- [Building and Installation](#building-and-installation)  
  - [Prerequisites](#prerequisites)  
  - [Clone the Repository](#1-clone-the-repository)  
  - [Install Dependencies](#2-install-dependencies)  
- [Building Geomys](#building-geomys)  
  - [Using Task](#using-task)  
  - [Manual Build](#manual-build)  
- [Running Geomys](#running-geomys)  
  - [Start the Server](#start-the-server)  
  - [Start a Client](#start-a-client)  
  - [Using Docker](#with-docker)  
- [Configuration](#configuration)  
- [Testing](#testing)  
- [Directory Structure](#directory-structure)  
- [TODOs & Future Work](#todos--future-work)  

---

## **Features**  
- **Data Structures** – Supports **key-value pairs, counters, and deques (stacks/queues)**.  
- **Flexible Deployment** – Runs in **single-node mode** or **multi-node cluster mode**.  
- **Leader-Follower Replication** – Only the leader handles writes, and followers replicate asynchronously.  
- **Eventual Consistency** – Ensures data synchronization across nodes over time.  
- **gRPC Communication** – High-performance inter-node messaging using Protocol Buffers.  
- **Scalable Clustering** – Distributes data across multiple nodes for horizontal scalability.  
- **Efficient Persistence** – Stores data on disk using a **custom binary format** for fast recovery.  
- **Automatic Failover** – Handles node failures with leader election and recovery mechanisms.  
- **Lightweight & Fast** – Optimized for speed and minimal resource usage.  

---

## **Architecture Overview**  

- **Cluster Management (`internal/cluster`)**  
  - **Leader election** using the highest node ID.  
  - **Heartbeat monitoring** for failure detection.  

- **Replication (`internal/cluster/replication`)**  
  - **Writes go to the leader**, which replicates changes to followers.  
  - **Followers sync on restart** by requesting missing commands from the leader.  

- **Data Storage (`internal/core`)**  
  - **In-memory key-value store** with support for lists and other data types.  
  - **Persistence layer** writes changes to disk for durability.  

- **Networking (`internal/network`)**  
  - Exposes a gRPC API for cluster communication.

For more detailed design overview, see [this](docs/Design.md)  

---

## **Building and Installation**  
Geomys uses [Go Task](https://taskfile.dev/) as the build tool.  

### **Prerequisites**  
- Go 1.23  
- Task 3  
- Proto3 (optional)  

### **1. Clone the repository**  
```sh
git clone https://github.com/vskvj3/geomys.git
cd geomys
```

### **2. Install dependencies**  
```sh
go mod tidy
```

---

## **Building Geomys**  

### **Using Task**  
You can build and run Geomys using Task:  

- **Run the server:**  
  ```sh
  task server
  ```

- **Run the client:**  
  ```sh
  task client
  ```

- **Build the server binary:**  
  ```sh
  task build-server
  ```

- **Build the client binary:**  
  ```sh
  task build-client
  ```

- **Build both server and client binaries:**  
  ```sh
  task build
  ```

- **Clean the build directory:**  
  ```sh
  task clean
  ```

### **Manual Build**  
If you prefer building manually, run:  
```sh
go build -o build/geomys-server.exe ./cmd/server
go build -o build/geomys-client.exe ./cmd/client
```

---

## **Running Geomys**  

### **Start the server**  
```sh
./build/geomys-server.exe
```

### **Start a client**  
```sh
./build/geomys-client.exe
```

Alternatively, run the client using Task:  
```sh
task client
```

### **With Docker** 
- Build Docker image
```sh
task docker-build
```
- Run the Docker container
```sh
task docker-run
```

- Remove Docker images
```sh
task docker-clean
```
---

## **Configuration**  

Geomys loads configuration from `~/.geomys/config.yaml`. Default values can be found in `internal/utils/config.go`.  

Example configuration:  
```yaml
node_id: 1
cluster_mode: true
log_level: "info"
```

---

## **Testing**  

Run unit and integration tests:  
```sh
task test
```

Run specific test files:  
```sh
go test ./tests/integration/integration_test.go
```

---

## **Directory Structure**  

```
geomys/
│── cmd/                  # CLI and server entry points
│   ├── client/           # Client implementation
│   ├── server/           # Server implementation
│
├── docs/                 # Documentation files
│
├── internal/             # Core logic of the project
│   ├── cluster/          # Leader election and cluster management
│   ├── core/             # Key-value store logic
│   ├── network/          # gRPC server
│   ├── persistence/      # Write-ahead logging and storage
│   ├── utils/            # Helper utilities
│
├── tests/                # Unit and integration tests
│
└── docker-compose.yaml   # Docker setup
```

---

## **TODOs & Future Work**  
 
- [ ] Improve **fault tolerance** and automatic recovery  
- [ ] Add **more data structures** like sorted sets  

---