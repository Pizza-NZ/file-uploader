# File Uploader

This project provides a simple file upload service built with Go, Nginx, Docker, and deployed on AWS using Terraform.

## Features

- File upload via HTTP POST request to AWS S3.
- Nginx reverse proxy for handling requests and serving static content.
- Dockerized application for easy deployment.
- Infrastructure as Code (IaC) with Terraform for AWS resource provisioning.
- Structured logging.
- Unit and integration tests.
- CI/CD pipeline with GitHub Actions for automated testing, Docker image building, and ECS deployment.

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
storage/ # New: S3 storage implementation
├── s3.go
├── storage_mock.go
└── storage.go
terraform/ # New: Terraform configurations for AWS infrastructure
├── alb.tf
├── cloudwatch.tf
├── ecr.tf
├── ecs.tf
├── iam.tf
├── main.tf
├── outputs.tf
├── s3.tf
├── secrets.tf
├── variables.tf
└── vpc.tf
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
- AWS Account and configured AWS CLI/credentials
- Terraform

### Installation

1.  Clone the repository:

    ```bash
    git clone https://github.com/pizza-nz/file-uploader.git
    cd file-uploader
    ```

2.  **Deploy Infrastructure with Terraform:**

    Navigate to the `terraform` directory and initialize Terraform, plan, and apply your infrastructure. Ensure your AWS credentials are configured.

    ```bash
    cd terraform
    terraform init
    terraform plan
    terraform apply
    ```

    This will provision AWS resources including an S3 bucket, ECR repository, ECS cluster, VPC, and more.

3.  **Build and Run Locally (for development/testing):**

    For local development and testing, you can still use Docker Compose.

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

-   **POST /upload**: Uploads a file to AWS S3. Expects a multipart form with a field named `uploadFile`.
    -   **Request**: `multipart/form-data`
    -   **Response**: `201 Created` with JSON body `{"fileId": "<uploaded_file_id>", "size": <file_size>}` on success.
-   **GET /health**: Health check endpoint.
    -   **Response**: `200 OK` with JSON body `"OK"`.

## Configuration

-   **`config.yml`**: Application configuration, now including AWS S3 bucket details. This file is updated by the CI/CD pipeline with values from Terraform outputs.
-   **`docker-compose.yml`**: Defines local development services, ports, and volumes.
-   **`proxy/nginx.conf`**: Nginx server configuration, including `client_max_body_size` and proxy pass settings.
-   **`terraform/`**: Contains all Terraform `.tf` files defining the AWS infrastructure.

## CI/CD

The project uses GitHub Actions for continuous integration and continuous deployment. The workflow defined in `.github/workflows/ci.yml` automatically runs tests, builds Docker images, pushes them to ECR, and deploys the updated application to AWS ECS on pushes to the `main` branch.

## Contributing

Feel free to open issues or submit pull requests.
