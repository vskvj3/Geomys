## **Starting the Software**  
### **Standalone Mode**  
```sh
geomys --node_id=1 --port=1000
```
- The `node_id` is optional.

### **Cluster Mode**  
#### **Bootstrapping the Leader Node**  
> [!NOTE]
> The internal port (used by client software) must be provided. If omitted, the node defaults to port `6973`.  
```sh
geomys --node_id=1 --port=1000 --bootstrap
```

#### **Joining as a Follower**  
- **Note:** When joining, you must use the **external port** of the leader node.  
  - If the leader node uses port `1000`, its **external port** is `2000`.  
  - By default:  
    ```sh
    External port = Internal port + 1000
    ```
##### **Follower 1**  
```sh
geomys --node_id=2 --port=1010 --join="127.0.0.1:2000"
```
##### **Follower 2**  
```sh
geomys --node_id=3 --port=1015 --join="127.0.0.1:2000"
```

---

## **Configurations**  
Basic configurations can be set in a configuration file.  
- By default, configurations are stored in the `.geomys` folder inside the home directory.  

### **Example: `geomys.conf`**  
```json
{
  "internal_port": 6379,
  "external_port": 8080,
  "default_expiry": 60000,
  "persistence": "writethroughdisk",
  "replication_enabled": false,
  "node_id": 1,
  "leader_id": false,
  "sharding_enabled": false,
  "cluster_mode": false
}
```

---

## **Basic Commands**  

### **PING**  
```json
{
  "Command": "PING",
  "Message": "",
  "Key": "",
  "Value": "",
  "Exp": 0,
  "Offset": null
}
```
#### **Response:**  
```json
{
  "message": "PONG",
  "status": "OK"
}
```

---

### **ECHO**  
```json
{
  "Command": "ECHO",
  "Message": "hi",
  "Key": "",
  "Value": "",
  "Exp": 0,
  "Offset": null
}
```
#### **Response:**  
```json
{
  "message": "hi",
  "status": "OK"
}
```

---

### **GET**  
```json
{
  "Command": "GET",
  "Message": "",
  "Key": "key",
  "Value": "",
  "Exp": 0,
  "Offset": null
}
```
#### **Success Response:**  
```json
{
  "status": "OK",
  "value": "value"
}
```
#### **Error Response:**  
```json
{
  "message": "Get failed: key not found",
  "status": "ERROR"
}
```

---

### **SET**  
```json
{
  "Command": "SET",
  "Message": "",
  "Key": "key",
  "Value": "helloworld",
  "Exp": 0,
  "Offset": null
}
```
#### **Success Response:**  
```json
{
  "status": "OK"
}
```

---

### **INCR**  
- `INCR` requires an **offset** value.  
```json
{
  "Command": "INCR",
  "Message": "",
  "Key": "h",
  "Value": "",
  "Exp": 0,
  "Offset": "1"
}
```
#### **Success Response:**  
```json
{
  "status": "OK",
  "value": 2
}
```
#### **Error (Non-Integer Value):**  
```json
{
  "message": "Value is not an integer",
  "status": "ERROR"
}
```

---

### **PUSH**  
- **Adds an element to the start** of a doubly-ended queue (deque).  
- If the list does not exist, it is created.  
```json
{
  "Command": "PUSH",
  "Message": "",
  "Key": "list",
  "Value": "1",
  "Exp": 0,
  "Offset": null
}
```
#### **Response:**  
```json
{
  "status": "OK"
}
```

---

### **RPOP**  
- **Removes and returns** an element from the **right-hand side** of the list.  
- **Non-blocking** (returns an error if the list is empty).  
```json
{
  "Command": "RPOP",
  "Message": "",
  "Key": "list",
  "Value": "",
  "Exp": 0,
  "Offset": null
}
```
#### **Success Response:**  
```json
{
  "status": "OK",
  "value": "2"
}
```

---

### **LPOP**  
- **Removes and returns** an element from the **left-hand side** of the list.  
- **Non-blocking** (returns an error if the list is empty).  
```json
{
  "Command": "LPOP",
  "Message": "",
  "Key": "list",
  "Value": "",
  "Exp": 0,
  "Offset": null
}
```
#### **Success Response:**  
```json
{
  "status": "OK",
  "value": "1"
}
```
#### **Error (List Does Not Exist):**  
```json
{
  "message": "LPOP failed: list is empty",
  "status": "ERROR"
}
```

---

### **FLUSHDB**  
> [!WARNING]
> `FLUSHDB` **clears the entire database**, including persisted disk data.  
```json
{
  "Command": "FLUSHDB",
  "Message": "",
  "Key": "",
  "Value": "",
  "Exp": 0,
  "Offset": null
}
```
#### **Response:**  
```json
{
  "status": "OK"
}
```
