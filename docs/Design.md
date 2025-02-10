> [!NOTE]
> This document contains design decisions and implementation considerations.
> This is not a usage document or instruction manual. If someone just wants to know the features, there is no need to waste time with this!

## Commands
### ECHO
**Request:**
```python
{'command': 'ECHO', 'message': 'hello'}
```
**Response:**
```python
{'status': 'OK', 'message': 'hello'}   
```

### SET
**Request:**
```python
{'command': 'SET', 'key': 'story', 'value': 'quick fox jumps over a lazy dog'}
```
**Response:**
```python
{'status': 'OK'}
```

### GET
**Request:**
```python
{'command': 'GET', 'key': 'story'}
```
**Response:**
```python
{'status': 'OK', 'value': 'quick fox jumps over a lazy dog'}
```

### INCR
- Only applicable to integer counters.

**Request:**
```python
{'command': 'INCR', 'key': 'counter', 'offset': '1'}
```
**Response:**
```python
{'status': 'OK', 'value': 2}
```

- INCR only supports integers up to 2<sup>63</sup>. Updating beyond that point is undefined.

## Data Types
There are some data types we are planning to integrate into the key-value store.

### Strings
- The primary and default data type will be strings.
```bash
>> set name john
>> get name
Server: john
```
- Multi-word strings can be used with double quotes.

> [!WARNING]
> Double quotes are not escaped in strings, so be careful.
```bash
>> set name "john doe"
>> get name
Server: john doe
```
- In API specification, strings are handled in MessagePack as shown below:
```go
map[command:SET key:name value:"john doe"]
```
- Implementing multi-word strings is the responsibility of the client-side, as the server will handle strings of any length for both keys and values.

### Counters
- Counters are a special type of value that can only increase over time.
```bash
>> set value_count 1
Server: OK
>> incr value_count 1
Server: 2
>> incr value_count 1
Server: 3
```
**INCR command implementation:**
```go
map[command:INCR key:value_count offset:1]
```
- The server returns a MessagePack object with `value` set to the new value upon completion.

### Stack and Queue
- Both stack and queue functionalities are implemented within the same structure.
- The single structure is called **LIST** and supports the following operations:
    - **PUSH**: Inserts an element into an existing list, or creates a new list if it does not exist.
    - **RPOP**: Removes and returns the last element of the list.
    - **LPOP**: Removes and returns the first element of the list.
- All these commands are non-blocking and return `STATUS: ERROR` upon unsuccessful execution.
```python
req: {'command': 'PUSH', 'key': 'test-stack', 'value': '1'}
res: {'status': 'OK'}

req: {'command': 'PUSH', 'key': 'test-stack', 'value': '2'}
res: {'status': 'OK'}

req: {'command': 'PUSH', 'key': 'test-stack', 'value': '3'}
res: {'status': 'OK'}

req: {'command': 'LPOP', 'key': 'test-stack'}
res: {'status': 'OK', 'value': '1'}

req: {'command': 'RPOP', 'key': 'test-stack'}
res: {'status': 'OK', 'value': '3'}
```

### Internal Implementation of Stack/Queue
- Since both stack and queue are developed within a single structure, it will function similarly to a **deque**.
- The application should be optimized for:
    - Insert at front: **O(1)**
    - Insert at rear: **O(1)**
    - Delete from front: **O(1)**
    - Delete from rear: **O(1)**

## Collision Behavior
This section explains the behavior of the basic **SET** command:
- If the key **already exists**:
    - Returns an error.
- If the key **does not exist**:
    - Creates a new key and stores the value.

## Persistence
### Storage Methods
1. **Write-Through Disk**
2. **Buffered Writes**

- Persisted databases are stored by default at `HOME:/.geomys/persistence.db`.
- Commands are stored after binary encoding.

### Write-Through Disk
- Every write operation to the cache is immediately written to persistent storage.
- It has significant write overhead.
- Highest I/O overhead and slower command execution.
- More reliable than **Buffered Writes**.
- Commands are stored in an **append-only file**, which is replayed upon server restart to restore the database.

### Buffered Writes
- Faster command execution.
- Data is grouped into batches and written at regular intervals.
- Data loss may occur if a failure happens before the last write.
- Preferable when some data loss is acceptable, and lower latency is required.

> Both persistence methods currently use an **append-only file** to store data.
- Only **one persistence mechanism** can be active at a time.
- By default, **no persistence is enabled**; it must be activated via a config file.

## High Availability Architecture
The system follows a **Leader-Follower** architecture.

#### Leader
A leader is responsible for following tasks:
- Writes: 
    - Leader is the only allowed node in the cluster to write.
    - All the other nodes redirects writes to the leader.
- House keeping:
    - Leader keeps track of the existing nodes in the cluster, and checks which of the nodes are alive by using a heartbeat mechanism. 
    - This heartbeat also contains a list of existing nodes in the system and a hash of the list. 
    - The list of the nodes will help the other nodes to keep track of the existing nodes.
    

### Leader Election
- Initially, the **node started in bootstrap mode** becomes the **default leader** (highest ID).
- Followers send **heartbeats every 5 seconds** to the leader.
- If the leader does not respond to **3 consecutive pings**, a new leader is elected.
- The **node with the highest ID** becomes the new leader, and replication resumes.

### Request Handling
- **Write Requests**: Routed to the leader → Replicated to followers.
- **Read Requests**: Can be handled by any node (leader or follower).

### Node Failures
- If a **replica fails**, it recovers by fetching the latest state from the leader.
- Heartbeat mechanism detects node failures (every 5 seconds).
- If a node is unresponsive for **15 seconds**, it is marked as failed.
#### Leader Failover
- If leader fails, nodes will detect the leader failure by delay in heartbeat.
- Leader election begins and a follower will become leader.
- 

### Cluster Management
- **Bootstrap Mode**: Start as a leader if no existing cluster if bootstrap flag is given when starting the server.
- **Join Mode**: Connect to an existing leader.
- **Standalone Mode**: Operate independently without clustering.

#### How cluster mode should look like?
- Stariting leader:
```bash
geomys -node_id=6767 -port=6767 -bootstrap
```
port 6767 will be used for the main server to run(which is used for clients to connect to run queries), main port + 1000 (7767) will be used for the grpc server, which will be used to distribute data and replications
- if no port flag, and node_id flag is given, they will be assigned automatically

- Starting follower and joining leader
```bash
geomys -node_id=6767 -port=6797 -join=127.0.0.1:7767
```

The follower will connect to leader on port 7767(grpc port), if leader is found working on this port, it must return an error

- Once we add multiple followers to join the cluster, then the cluster management will become more automatic(if the existing leader fails, next available follower will become leader after an election)

- If a follower fails, it fails. A maintainer can manually start up a new node to join an existing leader and it will work. The follower will not try to start up again, there is not point in that.
- Same goes for the leader node, if it completely fails, there is no point in starting it again. since it is gone, a new follower will take its job, if we want to scale again, create a new follower to join the existsing cluster.

## Leader election:
### Case 1:
- Each node will have a list of nodes:
- node will send heartbeat to followers in a specific time interval
- if the leader does not respond, the node will know the leader is down, and starts an election
- each node wil check the smallest node in the existing node list, and checks for the in request vote section, and grants vote to the smallest node id in its list. if the leader election fails, each node updates its nodes list by sending vote request to each node in nodes list. If any node does not respond, the node will be removed from the list
leader elction algorithm client:
- checks for the smallest node id in nodes list
- if it is other node:
    - sends canYouLeader request to the other node
    - the node sends grants request if it is alive
    - if the node is not alive, removes the node from the nodes list, and restarts the election
- if is current node:
    - starts node in leader mode


### case 3 quorum based:
- upon leader election starts, each node sends a voteRequest to all other nodes on the list.
- If the node is not alive, it is remove from the nodes list.
- If the node is alive:
    - Reciever node: checks the term and node id, if it is smaller than its own node id, it grant vote, otherwise refuses.
- If a node gets majority vote, it becomes leader, and sends a newLeader message to all other nodes.
- If it does not gets majority vote, if listens for newleader message. if newLeader message is recieved, it updates its own leader, and if it does not recieve newLeader message, another election starts.

### Case 3.2 
- Same as above, it sends vote request to other nodes, if the vote it recieved is from the smallest node in its nodes list, it will send granted, or does not grants vote. when a node gets majority vote. it will sent new leader request to all other nodes on the list. all other nodes will agree on the list will update its nodes list

### Case 4: Leader initiative:
- Leader sends heartbeat to each of the followers, followers responds to the heartbeat.
- Along with heart beat, the leader also sends existing nodes list to the follower, so the follower can update its nodes list.
- Leader keeps the node list with the nodes that responds to the heartbeat with address and nodes id.
- Upon the reciever does not recieve new heartbeat for 15 seconds, an election begins.
- Each node sends an VoteRequest to each of the node in its nodes list. 
- Upon recieving VoteRequest, the node wil grant the vote if its node id is the smallest in the nodes list.
- If all the nodes agrees on a single node, it will elect that node as the smallest, and the election will conclude.

### Final
- Each follower will send heartbeat to the leader.
- Leader will send the updated nodes list.
- When leader goes down, a new election begins.
- Each node will send a VoteRequest to all the nodes in its nodes list.
- Upon recieving this request, the nodes will check the smalllest node id in its list and send it back. 
- If a node receives same node id from all the votes, it will be selected as the new leader. 
- A canYouLeades message is send to this node by all the nodes, if the node grants the message, it will be new leader, else a new leader will election will start again(after removing that node from the list)
## Upcoming Considerations
### Blocking & Non-Blocking Commands
> In Redis, some commands block execution until a condition is met.
> For example, **BLPOP** blocks the client until an element is available in the list.
> - It can take a **timeout** value specifying when blocking ends.
> - See [Redis BLPOP documentation](https://redis.io/docs/latest/commands/blpop/) for reference.

