# Auth Service

## Overview

The **Authentication Service** is the core security microservice responsible for user authentication and authorization. It handles user registration, login, JWT token generation, and role-based access control (RBAC). The service ensures secure password hashing using bcrypt and publishes user creation events to other services via Kafka.

**Port:** `8000`  
**Module:** `sports/authservice`

---

## Key Features

### Core Capabilities

- **User Registration**: Secure registration with email validation and bcrypt password hashing
- **User Login**: Email/password authentication with JWT token pair generation
- **JWT Token Management**: Access tokens (24h) and refresh tokens (7d) for stateless authentication
- **Role-Based Access Control**: Assign roles (admin, coach, manager, player) and enforce permissions
- **Event Publishing**: Publishes `UserCreated` events to Kafka for downstream services (user-service, email notifications)
- **CORS Support**: Configurable cross-origin resource sharing for frontend integration

### Architecture

- **Framework**: Go with Gorilla Mux
- **Database**: PostgreSQL with Goose migrations
- **Event Queue**: Apache Kafka/Confluent Kafka
- **Password Hashing**: bcrypt (industry standard, resistant to brute-force attacks)
- **Authentication**: JWT (Header.Payload.Signature structure)

---

## Technologies Used

| Technology | Version | Purpose |
|------------|---------|---------|
| Go | 1.24.5 | Core language |
| PostgreSQL | - | User data persistence |
| Goose | v3.24.3 | Database migrations |
| Gorilla Mux | v1.8.1 | HTTP routing |
| golang-jwt | v5.3.0 | JWT creation/validation |
| bcrypt | crypto/v0.39.0 | Password hashing |
| Confluent Kafka | v2.11.0 | Event publishing |
| Google UUID | v1.6.0 | Unique identifier generation |
| Gorilla Handlers | v1.5.2 | CORS middleware |
| golangci-lint | v1.59.0 | Code linting |
| sqlmock | v1.5.2 | Testing database mocks |
| testify | v1.10.0 | Testing assertions |

---

## API Endpoints

| Method | Endpoint | Description | Auth Required | Request Body |
|--------|----------|-------------|---------------|--------------|
| POST | `/register` | Register new user | None | `firstName`, `lastName`, `email`, `password` |
| POST | `/login` | Authenticate user | None | `email`, `password` |

### Response Examples

**Register (200/417)**
```json
{
  "userId": "550e8400-e29b-41d4-a716-446655440000",
  "firstName": "John",
  "lastName": "Doe",
  "email": "john.doe@example.com",
  "createdat": "2025-01-01T10:00:00Z"
}
```

**Login (200/401)**
```json
{
  "user": {
    "firstName": "John",
    "lastName": "Doe",
    "email": "john.doe@example.com",
    "createdat": "2025-01-01T10:00:00Z"
  },
  "accessToken": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "refreshToken": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
}
```

---

## Authentication Flow

### User Registration
```
Client → POST /register (email, password)
         ↓
Auth Service: Check email uniqueness
         ↓
Hash password with bcrypt
         ↓
Store user in PostgreSQL
         ↓
Publish UserCreated event to Kafka (profiles topic)
         ↓
Response: User profile + 200 OK
```

### User Login
```
Client → POST /login (email, password)
         ↓
Auth Service: Query user by email
         ↓
Compare password hash with bcrypt.Compare()
         ↓
Fetch user roles from user_roles table
         ↓
Generate JWT pair (access + refresh tokens)
         ↓
Response: Tokens + User profile + 200 OK
```

### Protected Route Access
```
Client → GET /api/teams (Authorization: Bearer <accessToken>)
         ↓
Middleware: Extract & validate JWT signature
         ↓
Check token expiration
         ↓
Inject user claims into request context
         ↓
Route handler proceeds or returns 401 Unauthorized
```

---

## Database Schema

### Users Table
```sql
CREATE TABLE Users (
  id SERIAL PRIMARY KEY,
  userid UUID UNIQUE NOT NULL,
  firstname VARCHAR(100),
  lastname VARCHAR(100),
  email VARCHAR(100) UNIQUE NOT NULL,
  password VARCHAR(255) NOT NULL,
  created_at TIMESTAMP DEFAULT NOW(),
  updated_at TIMESTAMP DEFAULT NOW()
);
```

### User Roles Table
```sql
CREATE TABLE user_roles (
  user_id INT REFERENCES Users(id),
  role_id INT REFERENCES roles(id),
  PRIMARY KEY (user_id, role_id)
);

CREATE TABLE roles (
  id SERIAL PRIMARY KEY,
  name VARCHAR(50) UNIQUE NOT NULL
);
```

---

## JWT Token Structure

```
Header: {
  "alg": "HS256",
  "typ": "JWT"
}

Payload (Claims): {
  "id": 1,
  "userid": "550e8400-e29b-41d4-a716-446655440000",
  "email": "john@example.com",
  "roles": ["coach", "manager"],
  "exp": 1735689600,  // Unix timestamp (24h for access)
  "iat": 1735603200,
  "nbf": 1735603200
}

Signature: HMAC-SHA256(
  base64(header) + "." + base64(payload),
  "your-secret-key"
)
```

---

## Environment Variables

```bash
# Database
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=postgres
DB_NAME=auth_db
DB_SSLMODE=disable

# Kafka
KAFKA_BROKER=localhost:9092

# JWT Secrets (set in code, consider externalize)
JWT_SECRET=mydogsnameisrufus
REFRESH_SECRET=myotherdogsnameistommy

# CORS
CORS_ALLOWED_ORIGINS=http://localhost:3000,http://localhost:5173
```

---

## Development & Testing

### Run Tests
```bash
cd auth-service
go test ./...  # All packages
go test -v ./internal/service  # Specific package with verbose output
go test -cover ./...  # With coverage report
```

### Run Linting
```bash
golangci-lint run ./...
```

### Build Locally
```bash
go build -o auth-service ./cmd/main.go
./auth-service
```

### Docker Build & Run
```bash
docker build -t auth-service:latest .
docker run -p 8000:8000 --env-file .env auth-service:latest
```

---

## Error Codes

| Status | Error | Description |
|--------|-------|-------------|
| 200 | OK | Successful operation |
| 400 | Bad Request | Invalid input data |
| 401 | Unauthorized | Invalid credentials or expired token |
| 417 | Expectation Failed | Email already exists or validation failed |
| 500 | Internal Server Error | Database or server error |

---

## Kafka Events

### UserCreated Event
Published to topic: `profiles`

```json
{
  "userid": "550e8400-e29b-41d4-a716-446655440000",
  "firstName": "John",
  "lastName": "Doe",
  "email": "john.doe@example.com"
}
```

Consumed by: `user-service` to create user profiles

---

## Security Considerations

- Passwords hashed with bcrypt (cost factor: 12)
- JWT signed with HS256 algorithm
- Refresh tokens stored server-side (future improvement)
- CORS configured per environment
- TODO: Implement token blacklisting for logout
- TODO: Add rate limiting on login/register endpoints
- TODO: Implement email verification flow

---

## Deployment

See [k8s/](../../k8s/) for Kubernetes manifests. The service is deployed as part of the CI/CD pipeline:

1. Code pushed to `main` branch
2. golangci-lint validates code quality
3. `go test ./...` runs unit tests
4. Docker image built and pushed to Docker Hub
5. `kubectl apply` deploys to cluster

---

## Support & Troubleshooting

- **Port already in use**: Change `PORT` env var or kill process on port 8000
- **Database connection failed**: Verify `DB_HOST`, `DB_PORT`, `DB_NAME` in `.env`
- **Kafka connection failed**: Ensure Kafka broker is running on `KAFKA_BROKER` address
- **JWT validation fails**: Check `JWT_SECRET` matches across services
