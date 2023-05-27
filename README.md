# Real-Time Tracking System
This is a real-time tracking system consisting of three main parts: Tracking Service, Pub/Sub System, and CLI Client. The system allows for tracking events, propagating them to subscribed services, and displaying them in real-time.

## Features
Tracking Service: Exposes an endpoint to receive events and validates the account before propagating the event.  
Pub/Sub System: Accepts events from the Tracking Service and propagates them to subscribed services over the network.  
CLI Client: Fault-tolerant CLI client that subscribes to the Pub/Sub System and displays messages as they arrive. Supports filtering on multiple account IDs.
## Technologies Used
Go programming language  
Redis for pub/sub messaging  
Third-party libraries:  
[go-redis/redis](https://github.com/redis/go-redis): Redis client library for Go
## Prerequisites
Go 1.16 or higher  
Redis server
## Setup
1. Clone the repository:
git clone https://github.com/simepavlic/tracking-service.git

2. Install the project dependencies:
```
$ go mod tidy
```
3. Configure tracking service port in cmd/tracking-service/tracking-service.go file:
```
const servicePort = "8080"
```
4. Configure the Redis connection in the cmd/global/global.go file:
```
// Redis configuration
const RedisAddress = "localhost:6379"
const RedisChannel = "tracking-events"
```
5. Start the Tracking Service:
```
$ go run cmd/tracking-service/tracking-service.go
```
6. Start the CLI Client:
```
$ go run cmd/cli-client/client.go
```
## Usage
1. Sending Events:

To send an event, make a GET request to the endpoint localhost:{servicePort}/{accountId}?data={data}. Replace {servicePort} with tracking service port, {accountId} with the account ID and {data} with the data for the event.  

2. CLI Client:

The CLI client will display the events in real-time as they are received from the Pub/Sub System.

To filter by account IDs specify wanted IDs after go run CLI command. For example:
```
$ go run cmd/cli-client/client.go <accountID1> <accountID2> ...
```
Only events with the specified account IDs will be displayed.

3. Termination:

Press Ctrl+C to stop the Tracking Service or CLI Client.
## Testing
To run the tests, execute the following command:
```
$ go test ./...
```
## License
This project is licensed under the MIT License.