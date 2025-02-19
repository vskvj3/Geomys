It is an unstable version intended for early testing, and anything may change at any time.  

#### **Key Features**  
- Basic key-value store with **SET, GET, INCR, and ECHO** commands  
- **List (Stack & Queue)** with PUSH, LPOP, and RPOP operations  
- **Write-through disk persistence** with binary log storage  
- **Leader-Follower replication** with automatic leader election  
- **Cluster management** with manual node joining and failover  

#### **Known Limitations**  
- Unstable API, subject to change  
- No authentication or security  
- Performance optimizations pending  
- Write-heavy workloads may experience I/O slowdowns  