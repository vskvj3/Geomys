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

- Incr only support integer upto 2<sup>63</sup>, updating beyond that point is undefined.
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

### Stack and Queue
- Both stack and queues are implemented inside same structure.
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

#### Internal Implementaion of stack/queue
- Since we considered to develop both stack and queue in a single structure.
- The structure will act similar to dequeue.
- The application should be optimized for:
    - Insert at front: O(1)
    - Insert at rear: O(1)
    - Delete from front: O(1)
    - Delete from rear: O(1)

## Collision behaviour
This part explains the behaviour of basic set command:
- If key already exists
    - returns an error
- If key does not exists
    - creates a new key and saves the value

## Persistance
1. Write Through Disk
2. Buffered Writes
- Persisted databases are by default stored at `HOME:/.geomys/persistance.db`
- Commands are stored after binary encoding.

### Write Through Disk
- Every write operation to the cache is immediately written to the persistent storage.
- It has significate write overhead..
- Highest I/O overhead and slower command execution.
- More reliable than Buffered Writes.
- Commands are stored in an append-only-file when each command is run.
- Those commands are replayed at the time the server restarts to restore the database.
Considerations:
```
SET key 1
SET key 4
INCR key 8
INCR key 10
PUSH list 10
PUSH list 11
PUSH list 12
RPOP list
LPOP list
```
- For ease of handling, instead of storing the opearations as strings, it would be optimal to store the operations as binary objects into the append only file.
- each file could look like this:
```js
{
    timestamp datetime
	Command string 
	Key     string 
	Value   string
}
```
- The operations need to be logged are:
    - SET key value exp
    - INCR key offset
    - EXPR key (When keys expires) 
    - PUSH key value
    - RPOP key
    - LPOP key
- We can store and rebuild this commands similar to how we handle the requests and responses.
### Buffered Writes
- Faster command execution.
- Data is grouped into batches and written on regular intervals.
- Data loss may occure if the data is written after last write.
- Preferable in situations where some data loss is not as inconvicience and latency is more important.
> Both of these persistence methodes currently uses a append only file to store the data.
- It is possible to only one of these persitence mechanism at a time. 
- By default no persistence is enabled, it has to be enabled by using config file[More on this will be clarified later]



### Upcoming Considerations:
- Blocking and non  blocking commands
> In redis there are blocking and non blocking commands
> - for more info [see](https://redis.io/docs/latest/commands/blpop/)\
> In lists BLPOP command block the client connection if the list is empty untill anything is pushed into the list again(Instead of returning an error that the element does not exists)
> - It can take a timeout value, which specifies when the blocking ends.
> 
