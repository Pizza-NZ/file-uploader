# File Uploader Go Application Documentation

## 1. Overview

This document provides a detailed explanation of a file upload web service written in Go. The application is designed to allow users to upload files, which are then stored temporarily on the server.

The service is built with a modular architecture, separating concerns like logging, middleware, request handling, and file storage logic. It uses Nginx as a reverse proxy to handle incoming requests and serve static content.

## 2. Core Features

*   **File Upload:** Accepts multipart form data for file uploads.
*   **Nginx Reverse Proxy:** Handles incoming HTTP requests and forwards file upload requests to the Go service.
*   **Temporary File Storage:** Uploaded files are stored in a temporary directory on the server.
*   **Health Check Endpoint:** Provides a `/health` endpoint to check the service status.
*   **Structured Logging:** Implements structured logging for better observability.
*   **Request ID Middleware:** Adds a unique request ID to each incoming request for tracing.
*   **Robust Error Handling:** Utilizes custom error types for consistent and informative error responses.

## 3. Project Structure

The project is organized into the following directories:

*   `.github/workflows/`: Contains GitHub Actions workflows for CI/CD.
*   `cmd/`: The main application entry point (`main.go`) and integration tests.
*   `handlers/`: Contains HTTP handlers for file upload and health checks.
*   `logging/`: Handles application logging configuration.
*   `middleware/`: Provides HTTP middleware for request processing.
*   `proxy/`: Contains Nginx configuration.
*   `public/`: Stores static web content (e.g., `index.html`).
*   `services/`: Implements the core business logic for file handling.
*   `types/`: Defines shared data structures and custom error types.
*   `utils/`: Provides utility functions for JSON responses and error handling.

## 4. Key Components & Code

### 4.1. Main Application (`cmd/main.go`)

The `main.go` file sets up the HTTP server, defines routes, and initializes handlers and middleware.

```go
func main() {
	// ... (flag parsing and logger setup)

	mux := http.NewServeMux()
	handl := handlers.NewFileUploadHandler(200 << 20) // Max file size 200MB
	mux.HandleFunc("POST /upload", handl.CreateFileUpload)
	mux.HandleFunc("GET /health", handlers.HealthCheck)

	server := http.Server{
		Addr:    *addr,
		Handler: middleware.RequestIDMiddleware(mux),
	}

	// ... (server startup and graceful shutdown)
}
```

### 4.2. Handlers (`handlers/handlers.go`)

The `handlers` package contains the `FileUploadHandler` interface and its implementation. The `CreateFileUpload` method handles the actual file upload process, parsing multipart forms and delegating to the service layer. The `HealthCheck` function provides a simple status endpoint.

```go
// CreateFileUpload handles the file upload HTTP request.
func (h *FileUploadHandlerImpl) CreateFileUpload(w http.ResponseWriter, r *http.Request) {
	// ... (service initialization check)
	slog.Info("New Put request", "requestID", r.Header.Get("X-Request-ID"))
	r.ParseMultipartForm(h.maxFileSize)

	file, handler, err := r.FormFile("uploadFile")
	if err != nil {
		utils.HandleError(w, r, types.NewAppError("Error Reading File", "User file submitted failed to read", http.StatusBadRequest, err))
		return
	}

	fileUploadResponse, err := h.service.CreateFileUpload(file, handler)
	if err != nil {
		utils.HandleError(w, r, err)
		return
	}

	utils.JSONResponse(w, r, http.StatusCreated, fileUploadResponse)
}

// HealthCheck provides a simple health status for the service.
func HealthCheck(w http.ResponseWriter, r *http.Request) {
	utils.JSONResponse(w, r, http.StatusOK, "OK")
}
```

### 4.3. Services (`services/services.go`)

The `services` package contains the `FileUploadService` interface and its `FileUploadServiceImpl` implementation. This layer encapsulates the business logic for file storage.

```go
// CreateFileUpload handles the storage of the uploaded file.
func (s *FileUploadServiceImpl) CreateFileUpload(file multipart.File, handler *multipart.FileHeader) (*types.FileUploadResponse, error) {
	defer file.Close()

	// ... (temporary folder creation and file handling)

	tempFileName := fmt.Sprintf("upload-%s-*%s", utils.FileNameWithoutExtension(handler.Filename), filepath.Ext(handler.Filename))

	tempFile, err := os.CreateTemp(tempFolderPath, tempFileName)
	// ... (error handling and file writing)

	slog.Info("File uploaded successfully", "filename", handler.Filename)
	return &types.FileUploadResponse{FileID: filepath.Base(tempFile.Name()), Size: handler.Size}, nil
}
```

### 4.4. Error Handling (`types/errors.go`, `utils/utils.go`)

The application uses a centralized error handling strategy with a generic `AppError` struct for consistent error responses. The `HandleError` utility function in `utils/utils.go` ensures errors are logged and returned as JSON responses.

```go
// AppError is a generic error type for the application.
type AppError struct {
	Underlying      error `json:"-"`
	HTTPStatus      int    `json:"-"`
	Message         string `json:"message"`
	InternalMessage string `json:"-"`
}

// HandleError is a utility function to handle errors in HTTP handlers.
func HandleError(w http.ResponseWriter, r *http.Request, err error) {
	var appErr *types.AppError
	if errors.As(err, &appErr) {
		// This is our custom error type, we can trust its fields.
		slog.Error("Handle Error", "error", appErr.Error(), "requestID", r.Header.Get("X-Request-ID"))

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(appErr.HTTPStatus)
		json.NewEncoder(w).Encode(appErr)
		return
	}

	// For any other error, return a generic 500.
	slog.Error("An unexpected error occurred", "error", err.Error(), "requestID", r.Header.Get("X-Request-ID"))
	http.Error(w, `{"message":"An internal server error occurred."}`, http.StatusInternalServerError)
}
```

## 5. Design Patterns and Architectural Choices

This project employs several design patterns and architectural choices to ensure maintainability, testability, and scalability:

*   **Modular Architecture (Separation of Concerns):** The codebase is divided into distinct packages (`handlers`, `services`, `types`, `logging`, `middleware`, `utils`) each responsible for a specific aspect of the application. This separation enhances readability, makes it easier to locate and modify code, and promotes independent development and testing of components.

*   **Interfaces for Abstraction and Testability:** Interfaces like `FileUploadHandler` and `FileUploadService` are used to define contracts for behavior. This allows for loose coupling between components, making it easy to swap out implementations (e.g., for different storage backends) and, crucially, to create mock implementations for unit testing. This is evident in `handlers/handlers_test.go` and `services/services_test.go` where mock services are used to isolate the component under test.

*   **Dependency Injection:** Dependencies (e.g., `FileUploadService` in `FileUploadHandlerImpl`) are injected through constructors (`NewFileUploadHandler`). This promotes loose coupling and makes components easier to test by allowing mock dependencies to be provided during testing.

*   **Centralized Error Handling:** The `AppError` custom type and the `HandleError` utility function provide a consistent and centralized mechanism for handling errors across the application. This ensures that errors are logged uniformly, and meaningful, structured JSON responses are returned to the client, improving both debugging and user experience.

## 6. Testing Strategy

This project utilizes a multi-faceted testing strategy to ensure the reliability and correctness of the application:

*   **Unit Tests:**
    *   **Location:** `handlers/handlers_test.go`, `services/services_test.go`
    *   **Purpose:** These tests focus on individual components (functions, methods) in isolation. They verify the correctness of the business logic within the `services` package and the request/response handling within the `handlers` package.
    *   **Methodology:** Mock implementations (e.g., `MockFileUploadService`) are used to isolate the component under test from its dependencies, ensuring that only the logic of the unit itself is being validated. This provides fast feedback during development and helps pinpoint bugs precisely.

*   **Integration Tests:**
    *   **Location:** `cmd/integration_test.go`
    *   **Purpose:** These tests verify the interactions between different components and services, including the Go application, Nginx, and the Docker environment. They simulate real-world scenarios to ensure that the entire system functions correctly as a whole.
    *   **Methodology:** Docker Compose is used to spin up the entire application stack (Go service, Nginx) in a controlled environment. Tests then make HTTP requests to the Nginx proxy, mimicking client behavior, and assert on the responses. This ensures that network configurations, service discovery, and inter-service communication are working as expected.

## 7. Rationale and Reasoning

*   **Temporary File Storage:** For this project, files are stored temporarily on the server's local filesystem for simplicity. In a production environment, a more robust and scalable solution would be required, such as persistent storage volumes (e.g., Docker volumes, Kubernetes PVCs), cloud storage services (e.g., AWS S3, Google Cloud Storage), or a dedicated file storage system.

*   **Nginx as a Reverse Proxy:** Nginx is used as a reverse proxy for several reasons:
    *   **Load Balancing:** While not explicitly configured for load balancing in this simple setup, Nginx is highly capable of distributing incoming traffic across multiple instances of the Go service, improving scalability and reliability.
    *   **Static File Serving:** Nginx efficiently serves static assets (like `index.html`), offloading this task from the Go application and improving performance.
    *   **Security:** Nginx can act as a first line of defense, handling SSL termination, request filtering, and other security measures before requests reach the application.
    *   **API Gateway Functionality:** It provides a single entry point for clients, simplifying client-side configuration and allowing for easy routing of requests to different backend services.

*   **Structured Logging (slog):** The `log/slog` package is used for structured logging. This makes logs easier to parse, filter, and analyze with log management tools, improving observability and debugging capabilities, especially in production environments.

*   **Request ID Middleware:** The `RequestIDMiddleware` assigns a unique ID to each incoming HTTP request. This ID is propagated through logs, allowing for end-to-end tracing of a request's journey through different components of the system, which is invaluable for debugging and performance monitoring.