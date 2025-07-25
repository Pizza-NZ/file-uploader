## 1. Overview

This document provides a detailed explanation of a file upload web service written in Go. The application is designed to allow users to upload files, which are then stored in an AWS S3 bucket or a local mock storage for development.

The service is built with a modular architecture, separating concerns like logging, middleware, request handling, and file storage logic. It uses Nginx as a reverse proxy to handle incoming requests and serve static content. The entire infrastructure for cloud deployment is provisioned and managed using Terraform on AWS.

## 2. Core Features

*   **File Upload:** Accepts multipart form data for file uploads, storing them in AWS S3 or a mock storage.
*   **Robust File Type Validation:** Implements secure file type validation using magic number detection to prevent malicious uploads.
*   **Externalized Configuration:** Key application parameters like `maxFileSize` and `allowedTypes` are configurable via `config.yml`.
*   **Nginx Reverse Proxy:** Handles incoming HTTP requests and forwards file upload requests to the Go service.
*   **Cloud-Native Storage:** Utilizes AWS S3 for scalable and durable file storage.
*   **Local Mock Storage:** Provides a mock storage implementation for easy local development without AWS dependencies.
*   **Infrastructure as Code (IaC):** All AWS resources are defined and managed using Terraform.
*   **Health Check Endpoint:** Provides a `/health` endpoint to check the service status.
*   **Structured Logging:** Implements structured logging for better observability.
*   **Request ID Middleware:** Adds a unique request ID to each incoming request for tracing.
*   **Robust Error Handling:** Utilizes custom error types for consistent and informative error responses.
*   **CI/CD Pipeline:** Automated testing, Docker image building, and deployment to AWS ECS via GitHub Actions.

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
*   `storage/`: Contains the `FileStorage` interface, its AWS S3 implementation, and a mock implementation.
*   `terraform/`: Contains all Terraform configurations for AWS infrastructure.
*   `types/`: Defines shared data structures and custom error types.
*   `utils/`: Provides utility functions for JSON responses and error handling.

## 4. Key Components & Code

### 4.1. Main Application (`cmd/main.go`)

The `main.go` file sets up the HTTP server, defines routes, and initializes handlers and middleware. It now dynamically selects the storage implementation based on configuration.

```go
func main() {
	// ... (flag parsing and logger setup)

	cfg, err := config.NewConfig(*configPath)
	// ... (config validation and logger setup)

	var fileStorage storage.FileStorage
	switch cfg.StorageType {
	case "s3":
		fileStorage, err = storage.NewS3Storage(context.Background(), cfg.AWS)
		// ... (error handling)
	case "mock":
		fileStorage = storage.NewMockFileStorage()
	default:
		handleStartupError("Invalid storage type", fmt.Errorf("storage type '%s' is not supported", cfg.StorageType))
	}

	fileUploadService := services.NewFileUploadService(fileStorage, cfg.File.AllowedTypes)

	mux := http.NewServeMux()
	handl := handlers.NewFileUploadHandler(cfg.File.MaxSize, fileUploadService)
	mux.HandleFunc("POST /upload", handl.CreateFileUpload)
	mux.HandleFunc("GET /health", handlers.HealthCheck)

	server := http.Server{
		Addr:    cfg.Server.Port,
		Handler: middleware.RequestIDMiddleware(mux),
	}

	// ... (server startup and graceful shutdown)
}
```

### 4.2. Handlers (`handlers/handlers.go`)

The `handlers` package contains the `FileUploadHandler` interface and its implementation. The `CreateFileUpload` method handles the actual file upload process, parsing multipart forms and delegating to the service layer. The `maxFileSize` is now configured externally.

```go
// CreateFileUpload handles the file upload HTTP request.
func (h *FileUploadHandlerImpl) CreateFileUpload(w http.ResponseWriter, r *http.Request) {
	// ... (service initialization check)
	slog.Info("New Put request", "requestID", r.Header.Get("X-Request-ID"))
	r.ParseMultipartForm(h.maxFileSize) // maxFileSize is now from config

	file, handler, err := r.FormFile("uploadFile")
	// ... (error handling)

	fileUploadResponse, err := h.service.CreateFileUpload(r.Context(), file, handler)
	// ... (error handling)

	utils.JSONResponse(w, r, http.StatusCreated, fileUploadResponse)
}

// HealthCheck provides a simple health status for the service.
func HealthCheck(w http.ResponseWriter, r *http.Request) {
	utils.JSONResponse(w, r, http.StatusOK, "OK")
}
```

### 4.3. Services (`services/services.go`)

The `services` package contains the `FileUploadService` interface and its `FileUploadServiceImpl` implementation. This layer encapsulates the business logic for file storage and now includes robust file type validation using magic numbers, with allowed types configured externally.

```go
// CreateFileUpload handles the storage of the uploaded file.
func (s *FileUploadServiceImpl) CreateFileUpload(ctx context.Context, file multipart.File, handler *multipart.FileHeader) (*types.FileUploadResponse, error) {
	defer file.Close()

	// Read file header for magic number detection
	head := make([]byte, 261)
	// ... (read and reset file pointer)

	// Use filetype.Match to determine the file type based on magic numbers
	kind, err := filetype.Match(head)
	// ... (error handling)

	// Check if the detected file type is allowed (from configured allowedTypes)
	if kind == filetype.Unknown || !s.allowedTypes[kind.MIME.Value] {
		return nil, types.NewAppError("Invalid File Type", fmt.Sprintf("File type %s is not allowed", kind.MIME.Value), http.StatusBadRequest, nil)
	}

	// Delegate to storage.Upload (either S3 or mock)
	s3ObjectKey, err := s.fileStorage.Upload(ctx, file, handler)
	// ... (error handling)

	slog.Info("File uploaded successfully", "filename", handler.Filename, "s3_key", s3ObjectKey)
	return &types.FileUploadResponse{FileID: s3ObjectKey, Size: handler.Size}, nil
}
```

### 4.4. Storage (`storage/s3.go`, `storage/storage.go`, `storage/storage_mock.go`)

The `storage` package defines the `FileStorage` interface, abstracting file operations. It now includes two implementations:

*   **`s3.go`**: Provides an implementation for AWS S3, handling actual file uploads to the configured S3 bucket.
*   **`storage_mock.go`**: Provides a mock implementation for local development, simulating file uploads without interacting with AWS.

```go
// FileStorage defines the interface for file storage operations.
type FileStorage interface {
	Upload(ctx context.Context, file multipart.File, handler *multipart.FileHeader) (string, error)
}

// S3Storage implements the FileStorage interface for AWS S3.
type S3Storage struct {
	// ...
}

// NewS3Storage creates a new S3Storage instance.
func NewS3Storage(cfg config.AWSConfig) (FileStorage, error) {
	// ...
}

// MockFileStorage is a mock implementation of the FileStorage interface for local development.
type MockFileStorage struct {
	mock.Mock
}

// NewMockFileStorage creates a new MockFileStorage instance.
func NewMockFileStorage() *MockFileStorage {
	return &MockFileStorage{}
}

// Upload simulates a file upload for testing/local development.
func (m *MockFileStorage) Upload(ctx context.Context, file multipart.File, handler *multipart.FileHeader) (string, error) {
	// ... (mock logic)
}
```

### 4.5. Error Handling (`types/errors.go`, `utils/utils.go`)

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

*   **Modular Architecture (Separation of Concerns):** The codebase is divided into distinct packages (`handlers`, `services`, `storage`, `types`, `logging`, `middleware`, `utils`) each responsible for a specific aspect of the application. This separation enhances readability, makes it easier to locate and modify code, and promotes independent development and testing of components.

*   **Interfaces for Abstraction and Testability:** Interfaces like `FileStorage` are used to define contracts for behavior. This allows for loose coupling between components, making it easy to swap out implementations (e.g., for different storage backends like S3 or mock) and, crucially, to create mock implementations for unit testing. This is evident in `handlers/handlers_test.go` and `services/services_test.go` where mock services are used to isolate the component under test.

*   **Dependency Injection:** Dependencies (e.g., `FileStorage` in `FileUploadServiceImpl`) are injected through constructors (`NewS3Storage`, `NewMockFileStorage`). This promotes loose coupling and makes components easier to test by allowing mock dependencies to be provided during testing.

*   **Centralized Error Handling:** The `AppError` custom type and the `HandleError` utility function provide a consistent and centralized mechanism for handling errors across the application. This ensures that errors are logged uniformly, and meaningful, structured JSON responses are returned to the client, improving both debugging and user experience.

## 6. Testing Strategy

This project utilizes a multi-faceted testing strategy to ensure the reliability and correctness of the application:

*   **Unit Tests:**
    *   **Location:** `handlers/handlers_test.go`, `services/services_test.go`, `storage/storage_test.go` (if applicable)
    *   **Purpose:** These tests focus on individual components (functions, methods) in isolation. They verify the correctness of the business logic within the `services` package and the request/response handling within the `handlers` package.
    *   **Methodology:** Mock implementations (e.g., `MockFileUploadService`, `MockFileStorage`) are used to isolate the component under test from its dependencies, ensuring that only the logic of the unit itself is being validated. This provides fast feedback during development and helps pinpoint bugs precisely.

*   **Integration Tests:**
    *   **Location:** `cmd/integration_test.go`
    *   **Purpose:** These tests verify the interactions between different components and services, including the Go application, Nginx, and the Docker environment. They simulate real-world scenarios to ensure that the entire system functions correctly as a whole.
    *   **Methodology:** Docker Compose is used to spin up the entire application stack (Go service, Nginx) in a controlled environment. Tests then make HTTP requests to the Nginx proxy, mimicking client behavior, and assert on the responses. This ensures that network configurations, service discovery, and inter-service communication are working as expected.

## 7. Rationale and Reasoning

*   **AWS S3 for File Storage:** The application now uses AWS S3 for file storage, providing a robust, scalable, and highly available solution for storing uploaded files. This is a significant improvement over temporary local filesystem storage, making the application suitable for production environments.

*   **Local Mock Storage:** The introduction of a `MockFileStorage` implementation significantly improves the local development experience. Developers can now run and test the application without needing to configure AWS credentials or deploy any cloud infrastructure, leading to faster iteration cycles and reduced development costs.

*   **Secure File Type Validation:** The transition from `http.DetectContentType` to magic number-based detection using the `filetype` library drastically enhances the security of file uploads. This prevents malicious actors from disguising harmful files as legitimate ones, mitigating a critical vulnerability.

*   **Externalized Configuration:** Moving `maxFileSize` and `allowedTypes` to `config.yml` improves the application's flexibility and maintainability. These parameters can now be easily adjusted without recompiling the application, simplifying updates and deployments.

*   **Nginx as a Reverse Proxy:** Nginx is used as a reverse proxy for several reasons:
    *   **Load Balancing:** While not explicitly configured for load balancing in this simple setup, Nginx is highly capable of distributing incoming traffic across multiple instances of the Go service, improving scalability and reliability.
    *   **Static File Serving:** Nginx efficiently serves static assets (like `index.html`), offloading this task from the Go application and improving performance.
    *   **Security:** Nginx can act as a first line of defense, handling SSL termination, request filtering, and other security measures before requests reach the application.
    *   **API Gateway Functionality:** It provides a single entry point for clients, simplifying client-side configuration and allowing for easy routing of requests to different backend services.

*   **Structured Logging (slog):** The `log/slog` package is used for structured logging. This makes logs easier to parse, filter, and analyze with log management tools, improving observability and debugging capabilities, especially in production environments.

*   **Request ID Middleware:** The `RequestIDMiddleware` assigns a unique ID to each incoming HTTP request. This ID is propagated through logs, allowing for end-to-end tracing of a request's journey through different components of the system, which is invaluable for debugging and performance monitoring.

*   **Infrastructure as Code (Terraform):** The adoption of Terraform for infrastructure provisioning ensures that the entire AWS environment is defined, versioned, and managed as code. This provides consistency, repeatability, and reduces the risk of manual configuration errors, making deployments more reliable and efficient.

*   **CI/CD with GitHub Actions:** The comprehensive CI/CD pipeline automates the entire software delivery process, from code changes to deployment. This enables rapid iteration, ensures code quality through automated testing, and provides a streamlined, reliable deployment mechanism to AWS ECS.
