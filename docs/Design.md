## Commands
### ECHO
request:
```python
{'command': 'ECHO', 'message': 'hello'}
```
response:
```python
{'status': 'OK', 'message': 'hello'}   
```
### SET
request:
```python
{'command': 'SET', 'key': 'story', 'value': 'quick fox jumbs over a lazy dog'}
```
response:
```python
{'status': 'OK'}
```

### GET
request:
```python
{'command': 'GET', 'key': 'story'}
```
response:
```python
{'status': 'OK', 'value': 'quick fox jumbs over a lazy dog'}
```

### INCR
- only applicable to integer counters

request
```python
{'command': 'INCR', 'key': 'counter', 'offset': '1'}
```
response
```python
{'status': 'OK', 'value': 2}
```
## Data Types:
There are some data types we are planning to itegrate into the key value store
### Strings
- Primary and default data type will be strings
```bash
>> set name john
>> get name
Server: john
```
- Multiword strings can be used with double quotes
> [!WARNING]
> double quotes are not escaped in strings, so be careful.
```bash
>> set name "john doe"
>> get name
Server: john doe
```
- In api specification strings are handled in messagePack as shownn below:
```go
map[command:SET key:name value:john doe]
```
- Implementing word strings is a duty of client side, since server will be able to hanlde any length of strings key and value.
### Counters
- Counters are special type of values that can only increase as time passes
```bash
>> set value_count 1
Server: OK
>> incr value_count 1
Server: 2
>> incr value_count 1
Server: 3
```
incr command is implemented as below
```go
map[command:INCR key:value_count offset:1]
```
server returns a messagepack with `value` set to new value upon completion.

### Stack [Under Construction]
- Stack is the LIFO list supported by geomys
### Queue [Under Construction]
- Queue is the FIFO list supported by geomys
## Collision behaviour
This part explains the behaviour of basic set command:
- If key already exists
    - returns an error
- If key does not exists
    - creates a new key and saves the value

## Persistance
- We plans to use two types of persistance
1. Snapshots
2. WTD: Write Through Disk
### Snapshots
- State of the current in-memory database copied to disk at regular intervals
- intervals can be defined on the config file or by using the `CONFIG` command. [TODO]
- Some data may be lost between snapshots
- Only prefered in cases speed > realiability
### Write Through Disk
- Every write operation to the cache is immediately written to the persistent storage
- It has significate write overhead and should only be used in cases where reliability > latency