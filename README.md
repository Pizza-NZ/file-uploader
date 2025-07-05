# File Uploader

This project provides a simple file upload service built with Go, Nginx, and Docker.

## Features

- File upload via HTTP POST request.
- Nginx reverse proxy for handling requests and serving static content.
- Dockerized application for easy deployment.
- Structured logging.
- Unit and integration tests.
- CI/CD pipeline with GitHub Actions.

## Project Structure

```
.github/
├── workflows/
│   └── ci.yml
cmd/
├── integration_test.go
└── main.go
docker-compose.yml
dockerfile
go.mod
go.sum
handlers/
├── handlers.go
└── handlers_test.go
logging/
└── logging.go
makefile
middleware/
└── middleware.go
proxy/
└── nginx.conf
public/
└── index.html
services/
├── services.go
└── services_test.go
types/
├── errors.go
└── types.go
utils/
└── utils.go
README.md
TODO.md
```

## Getting Started

### Prerequisites

- Docker
- Docker Compose
- Go (for local development and running tests outside Docker)

### Installation

1.  Clone the repository:

    ```bash
    git clone https://github.com/pizza-nz/file-uploader.git
    cd file-uploader
    ```

2.  Build and run the Docker containers:

    ```bash
    docker-compose up --build
    ```

    The Go service will be running on port `2131` (inside Docker) and Nginx will be accessible on `http://localhost:8080`.

### Running Tests

#### Unit Tests

To run unit tests for all packages:

```bash
go test -v ./...
```

To run unit tests for a specific package (e.g., `handlers`):

```bash
go test -v ./handlers
```

#### Integration Tests

To run integration tests (requires Docker and Docker Compose):

```bash
docker-compose up --build --abort-on-container-exit integration-tests
```

## API Endpoints

-   **POST /upload**: Uploads a file. Expects a multipart form with a field named `uploadFile`.
    -   **Request**: `multipart/form-data`
    -   **Response**: `201 Created` with JSON body `{"fileId": "<uploaded_file_id>", "size": <file_size>}` on success.
-   **GET /health**: Health check endpoint.
    -   **Response**: `200 OK` with JSON body `"OK"`.

## Configuration

-   **`docker-compose.yml`**: Defines the services, ports, and volumes.
-   **`proxy/nginx.conf`**: Nginx server configuration, including `client_max_body_size` and proxy pass settings.

## CI/CD

The project uses GitHub Actions for continuous integration. The workflow defined in `.github/workflows/ci.yml` automatically runs tests and builds Docker images on pushes to the `main` branch and pull requests.

## Contributing

Feel free to open issues or submit pull requests.
