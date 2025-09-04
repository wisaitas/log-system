# Distributed Logging System

A custom-built distributed logging system for microservices architecture using Go and Fiber framework.

## Architecture Overview

This system consists of two main services:
- **Server** (Port 8081) - Entry point service that receives external requests
- **Processor** (Port 8082) - Processing service that handles business logic

## System Components

### 1. Custom Logging Middleware (`pkg/middleware.go`)
- **Distributed Tracing**: Automatic trace ID generation and propagation
- **Request/Response Logging**: Complete HTTP request and response capture
- **Source Tracking**: Tracks microservice communication chains
- **Performance Monitoring**: Request duration tracking
- **File Location Tracking**: Automatic source file and line number capture

### 2. API Caller (`pkg/apicaller.go`)
- **Service-to-Service Communication**: Standardized HTTP client wrapper
- **Header Propagation**: Automatic trace ID and internal call header management
- **Error Handling**: Integrated error response handling
- **Context Preservation**: Maintains request context across service calls

### 3. Response Handler (`pkg/response.go`)
- **Standardized Responses**: Consistent API response format
- **Error Code System**: Structured error codes (E40000, E50000, etc.)
- **Success Responses**: Standardized success response format
- **Pagination Support**: Built-in pagination structure

## Key Benefits of Custom Logging System

### 1. **Distributed Tracing Capabilities**
```json
{
  "trace_id": "080756bb07ef25ebf08ad51d020678b9",
  "timestamp": "2025-09-04T02:25:39.1585767+07:00",
  "current": { "service": "svc-a", ... },
  "source": { "service": "svc-b", ... }
}
```
- **End-to-End Visibility**: Track requests across multiple services
- **Request Correlation**: Link related requests in distributed systems
- **Performance Analysis**: Identify bottlenecks in service chains

### 2. **Comprehensive Request/Response Logging**
```json
{
  "request": {
    "headers": { "Content-Type": "application/json", ... },
    "body": { "first_name": "John", "last_name": "Doe" }
  },
  "response": {
    "headers": { "Content-Type": "application/json" },
    "body": { "error": "b is not allowed" }
  }
}
```
- **Complete Context**: Full request and response data for debugging
- **Header Tracking**: All HTTP headers captured for analysis
- **Body Inspection**: Request/response payload logging
- **Error Details**: Comprehensive error information

### 3. **Source Code Tracking**
```json
{
  "file": "/path/to/service/hello.go:26"
}
```
- **Precise Error Location**: Exact file and line number where errors occur
- **Debugging Efficiency**: Quick identification of problem sources
- **Code Maintenance**: Easy tracking of error-prone code sections

### 4. **Performance Monitoring**
```json
{
  "duration_ms": "0",
  "timestamp": "2025-09-05T01:29:02+07:00"
}
```
- **Response Time Tracking**: Millisecond-precision duration measurement
- **Performance Baselines**: Historical performance data
- **Bottleneck Identification**: Slow service detection

### 5. **Standardized Error Handling**
```json
{
  "status_code": "400",
  "code": "E40000",
  "message": "error message"
}
```
- **Consistent Error Format**: Uniform error response structure
- **Error Classification**: Structured error codes for categorization
- **Client Integration**: Easy error handling in frontend applications

### 6. **Microservice Communication Tracking**
- **Service Chain Visibility**: Track requests through multiple services
- **Internal Call Detection**: Distinguish between external and internal calls
- **Dependency Mapping**: Understand service relationships

## Advantages Over Standard Logging

### 1. **Business Context Integration**
- Logs include business-relevant data (user input, processing results)
- Not just technical logs, but business process logs
- Better correlation between technical issues and business impact

### 2. **Distributed System Optimization**
- Built specifically for microservices architecture
- Automatic service-to-service call tracking
- Better than generic logging solutions for distributed systems

### 3. **Development Efficiency**
- Automatic file location tracking reduces debugging time
- Structured error codes enable quick issue identification
- Complete request/response context eliminates guesswork

### 4. **Production Monitoring**
- Real-time performance monitoring
- Service health tracking
- Error pattern analysis

### 5. **Compliance and Auditing**
- Complete request/response audit trail
- Traceable user actions across services
- Regulatory compliance support

## Usage

### Running the System
```bash
# Terminal 1 - Start Server
make run

# Terminal 2 - Start Processor  
make processor
```

### Testing
```bash
# Test Server
curl -X POST http://localhost:8081/do \
  -H "Content-Type: application/json" \
  -d '{"first_name": "John", "last_name": "Doe"}'

# Test Processor
curl -X POST http://localhost:8082/do/a \
  -H "Content-Type: application/json" \
  -d '{"first_name": "John", "last_name": "Doe"}'
```

## Dependencies

```bash
go get github.com/gofiber/fiber/v2
go get github.com/google/uuid
```

## Log Output Example

The system produces structured JSON logs like:

```json
{
  "timestamp": "2025-09-05T01:29:02+07:00",
  "duration_ms": "0",
  "current": {
    "service": "processor",
    "method": "POST",
    "path": "localhost:8082/do/b",
    "status_code": "400",
    "request": {
      "headers": { "Content-Type": "application/json" },
      "body": { "first_name": "John", "last_name": "Doe" }
    },
    "response": {
      "headers": { "Content-Type": "application/json" },
      "body": { "error": "b is not allowed" }
    }
  },
  "source": {
    "service": "processor",
    "method": "POST",
    "path": "localhost:8082/do/b",
    "status_code": "400",
    "request": { ... },
    "response": { ... }
  }
}
```

## Conclusion

This custom logging system provides significant advantages for microservices architectures:

1. **Better Debugging**: Complete context and source location tracking
2. **Performance Insights**: Detailed timing and bottleneck identification  
3. **Distributed Visibility**: End-to-end request tracking across services
4. **Business Integration**: Logs that include business context and data
5. **Production Readiness**: Built-in monitoring and error handling
6. **Developer Experience**: Automatic file tracking and structured errors

The system is specifically designed for modern microservices environments and provides better observability than generic logging solutions.

