> [!NOTE]
> This document contains design decisions and implementaion considerations
> This is not a usage documentation or instruction, If someone just want to know the features, there is no need to waste your time with this!

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

Some considerations when implementing stack
```python
    req1: {'command': 'SPUSH', 'key': 'test-stack', 'value': '1'}
    resp: {'status': 'OK'}

    req2: {'command': 'SPUSH', 'key': 'test-stack', 'value': '2'}
    resp: {'status': 'OK'}

    req3: {'command': 'SPOP', 'key': 'test-stack'}
    resp: {'status': 'OK', 'value': '2'}
```
### Queue [Under Construction]
- Queue is the FIFO list supported by geomys
Some considerations when implementing stack
```python
    req1: {'command': 'QPUSH', 'key': 'test-stack', 'value': '1'}
    resp: {'status': 'OK'}

    req2: {'command': 'QPUSH', 'key': 'test-stack', 'value': '2'}
    resp: {'status': 'OK'}

    req3: {'command': 'QPOP', 'key': 'test-stack'}
    resp: {'status': 'OK', 'value': '1'}
```
#### Update:
- Decided to implement both stack and queues inside same structure.
- The single structure will be called LIST and supports two operations:
    - PUSH: Insert an element to an existsing list, create a new list if it does not exists.
    - RPOP: Pop an element from the end of the list.
    - LPOP: Pop an element from the start of the list.
- All of these commands and non-blocking and returns `STATUS:ERROR` upon unsuccessful execution.
```Python
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
- Additional considerations on implementation:
    - Don't use normal lists and list slicing, since it introduces additional memory overhead.
    - Slicing method will also introduce time related overhead when coming to pop operations.
    - Go collections have their own queue and stack implementation, can we modify the internal implementation and use that for our queue/stack?
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

### Upcoming Considerations:
- Blocking and non  blocking commands
> In redis there are blocking and non blocking commands
> - for more info [see](https://redis.io/docs/latest/commands/blpop/)\
> In lists BLPOP command block the client connection if the list is empty untill anything is pushed into the list again(Instead of returning an error that the element does not exists)
> - It can take a timeout value, which specifies when the blocking ends.
> 
