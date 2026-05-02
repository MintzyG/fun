# **Go Response Library**

A flexible and powerful Go library for creating standardized and consistent JSON HTTP responses. This library simplifies response handling with a fluent, chainable API, built-in support for common HTTP status codes, and a powerful interceptor system for middleware-like functionality.  
It is designed to be lightweight, and easily configurable to fit the needs of your application.

## **✨ Features**

* **Fluent API**: Chain methods together to build complex responses in a readable and intuitive way.  
* **Standardized JSON Structure**: Enforce a consistent JSON output format (message, data, trace, etc.) across your entire API.  
* **Common HTTP Status Helpers**: Pre-defined builder functions for common HTTP statuses (OK, BadRequest, NotFound, etc.).  
* **Detailed Trace Information**: Easily append debugging and trace information to your responses.  
* **Global Configuration**: Customize default settings like content type, response size limits, and more.  
* **Interceptor System**: Add custom logic (e.g., logging, metrics) that runs just before a response is sent.  
* **Custom Error Types**: A set of specific error types for handling issues like configuration, validation, and encoding.  
* **Thread-Safe**: Designed for concurrent use in high-performance web servers.

## **💾 Installation**

To install the library, use go get:  
```bash
go get github.com/MintzyG/fun
```

## **🚀 Quick Start**

Here's a simple example of how to use the library within a standard http.HandlerFunc.  
package main

```go
package main

import (
	"log"
	"net/http"

	"github.com/MintzyG/fun"
)

type User struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

func GetUserHandler(w http.ResponseWriter, r *http.Request) {
	user := User{ID: 1, Name: "John Doe"}

	// Create a 200 OK response with a message and data
	response.OK().
		WithMessage("User found successfully").
		WithData(user).
		Send(w) // Sends the response to the client
}

func main() {
	http.HandleFunc("/user", GetUserHandler)
	log.Println("Server running on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
```
**Test Command:**
```bash
curl -X GET localhost:8080/user
```
**Example JSON Output:**  
```json
{  
    "message": "User found successfully",  
    "data": {  
        "id": 1,  
        "name": "John Doe"  
    },  
    "timestamp": "2023-10-27T10:00:00Z",  
    "code": 200  
}
```

## **📖 Core Concepts**

### **Response Builders**

The library provides convenient builder functions for most standard HTTP status codes. You can then chain methods to customize the response.

#### **Successful Responses**

Use builders like OK(), Created(), or Accepted() for successful responses.  
```go
// 201 Created response  
response.Created().  
    WithMessage("Resource created").  
    WithData(map[string]int{"id": 123}).  
    Send(w)
```

#### **Error Responses**

Use builders like BadRequest(), NotFound(), or InternalServerError() for error responses.  
```go
// 404 Not Found response  
response.NotFound().  
    WithMessage("The requested resource was not found").  
    AppendTrace("database: record not found for id: 42").  
    Send(w)
```

### **Chaining Methods**

Most methods on the Response struct return the response object, allowing you to chain calls fluently.

* `.WithMessage(string)`: Sets the main message for the response.  
* `.WithData(any)`: Attaches a payload (struct, map, etc.) to be marshaled as JSON.  
* `.WithModule(string)`: Specifies the application module this response originated from.  
* `.WithContentType(string)`: Overrides the default application/json content type.  
* `.AppendTrace(...any)`: Adds one or more strings or errors to the trace array for debugging.

### **Handling Validation Errors**

The WithValidationErrors function provides a standardized way to return detailed validation failures.  
```go
// You can create your own validation logic  
validationErrors := []response.ValidationErr{  
    {Field: "email", Message: "Email address is not valid", Value: "invalid-email"},  
    {Field: "password", Message: "Password must be at least 8 characters long"},  
}

// This automatically creates a 400 Bad Request response  
response.WithValidationErrors(validationErrors).Send(w)
```

**Example JSON Output:**  
```json
{  
    "message": "Validation failed",  
    "trace": [  
        "(email) Email address is not valid: invalid-email",  
        "(password) Password must be at least 8 characters long"  
    ],  
    "timestamp": "2023-10-27T10:00:00Z",  
    "code": 400  
}
```

## **🛠️ Advanced Usage**

### **Global Configuration**

You can customize the library's default behavior by setting a global configuration. This is best done once when your application starts.  
```go
func init() {  
	newConfig := response.Config{  
		MaxTraceSize:         100,                  // Allow more trace entries  
		ResponseSizeLimit:    20 * 1024 * 1024,     // 20MB response limit  
		MaxInterceptorAmount: 10,  
		DefaultContentType:   "application/vnd.api+json",  
		EnableSizeValidation: true,                 // Validate response size before sending  
	}  
	response.SetConfig(newConfig)  
}
```

### **Interceptors**

Interceptors allow you to execute custom logic just before a response is sent. This is useful for cross-cutting concerns like logging, metrics, or injecting headers.  
An interceptor must implement the ResponseInterceptor interface.

#### **Example: A Logging Interceptor**

Here is an example of an interceptor that logs every response.  
package main

```go
import (  
	"context"  
	"log"  
	"net/http"  
	"your/package/path/response"  
)

// SimpleLogger is a custom interceptor  
type SimpleLogger struct{}

// Intercept is called for responses sent with a context  
func (l *SimpleLogger) Intercept(ctx context.Context, resp *response.Response, statusCode int) {  
	// You could extract a request ID or user info from the context  
	requestID := ctx.Value("requestID")  
	log.Printf("RequestID: %v | Sending Response -> Status: %d, Message: %s", requestID, statusCode, resp.Message)  
}

// InterceptSimple is called for responses sent without a context  
func (l *SimpleLogger) InterceptSimple(resp *response.Response, statusCode int) {  
	log.Printf("Sending Response -> Status: %d, Message: %s", statusCode, resp.Message)  
}

func init() {  
	// Register the interceptor during application startup  
	err := response.AddInterceptor(&SimpleLogger{})  
	if err != nil {  
		log.Fatalf("Failed to add interceptor: %v", err)  
	}  
}

// Example handler that uses the context  
func MyHandler(w http.ResponseWriter, r *http.Request) {  
	// Add a request ID to the context  
	ctx := context.WithValue(r.Context(), "requestID", "xyz-123")

	response.OK().  
		WithMessage("Success!").  
		SendWithContext(ctx, w) // Use SendWithContext to pass the context  
}

func main() {  
	http.HandleFunc("/", MyHandler)  
	log.Println("Server starting on :8080")  
	http.ListenAndServe(":8080", nil)  
}
```

When a request is made to /, the logger will print a line like this before the response is sent:  
> 2023/10/27 10:30:00 RequestID: xyz-123 | Sending Response \-\> Status: 200, Message: Success\!  
