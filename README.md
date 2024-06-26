# TykHealthcheck

## Overview

TykHealthcheck is a flexible health checking framework designed for Tyk ecosystem applications. It offers a structured approach to implement and manage both liveness and readiness checks, ensuring your services are healthy and ready to handle requests. With TykHealthcheck, you can define custom health checks tailored to your application's specific needs and easily expose them over HTTP.

## Key Features

- **Custom Health Checks**: Easily define your own liveness and readiness checks specific to your application's requirements.
- **Health Check Types**: Distinguish between essential (required) and optional checks.
- **HTTP Handler Support**: TykHealthcheck provides HTTP handlers to expose your health checks over HTTP, making it easy to integrate with your application's routing and middleware.
- **Caching**: The library features built-in caching that performs periodic updates of the cache at a configurable interval. Once set, the library will automatically execute the checker every specified number of seconds, ensuring that an up-to-date value is always readily available for retrieval.
- 
## Getting Started

### Prerequisites

Before you begin, ensure you have a working Go environment. TykHealthcheck is built using Go and requires Go installed to compile and run the applications that use it.

### Installing TykHealthcheck

To start using TykHealthcheck in your project, you need to add it as a dependency:

```sh
go get github.com/TykTechnologies/TykHealthcheck
```
### Implementing Health Checks
Implementing health checks with TykHealthcheck involves creating health checkers, registering checks, and exposing them via HTTP. Here's a quick start guide:

#### 1- Create Health Checkers:
Initialize health checkers for liveness and readiness checks.

```go
readinessHealthChecker := hc.NewHealthChecker()
livenessHealthChecker := hc.NewHealthChecker()
```

#### 2- Register checks
Define and register your custom health checks.

```go
livenessHealthChecker.RegisterCheck("PingCheck", hc.Required, func() (hc.HealthStatus, error) {
    return hc.StatusPass, nil
})

readinessHealthChecker.RegisterCheck("Database", hc.Required, func() (hc.HealthStatus, error) {
    // Implement your check logic here
    return hc.StatusPass, nil
})
```

Optionally you can enable the caching mechanism by calling the method `WithCache(seconds)`, Eg: 

```go
readinessHealthChecker.RegisterCheck("Database", hc.Required, func() (hc.HealthStatus, error) {
    // Implement your check logic here
    return hc.StatusPass, nil
}).WithCache(100)
```

#### 3- Expose Health Checks:
Add HTTP handlers to expose your health checks.

```go
http.HandleFunc("/health/live", livenessHealthChecker.HTTPHandler())
http.HandleFunc("/health/ready", readinessHealthChecker.HTTPHandler())
```

#### 4- Start the HTTP Server:
Listen on a port to serve the health check endpoints.

```go
http.ListenAndServe(":9000", nil)
```

