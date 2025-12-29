# User Service

## Overview

The **User Service** is responsible for managing user profile data independently of authentication. It follows the principle of domain-driven design where this service owns the `profiles` table exclusively. The service is **event-driven**, consuming `UserCreated` events from the auth-service via Kafka and creating user profiles asynchronously. It also exposes a gRPC interface for other services to query user information.

**Port:** `9000` (HTTP)  
**gRPC Port:** `50051`  
**Module:** `github.com/wycliff-ochieng`

---

## Key Features

### Core Capabilities

- **User Profile Management**: Create, retrieve, and update user profiles
- **Event-Driven Architecture**: Listens to `UserCreated` events from auth-service (Kafka)
- **gRPC Service**: Exposes user data via gRPC for internal service-to-service communication
- **JWT-Protected Routes**: All endpoints require valid Bearer token
- **Profile Metadata**: Store and manage user location, preferences, and other metadata
- **Resilient Design**: If service is down, Kafka events are queued and processed on recovery

### Architecture

- **Framework**: Go with Gorilla Mux + gRPC
- **Database**: PostgreSQL with Goose migrations
- **Event Queue**: Apache Kafka/Confluent Kafka (Consumer)
- **IPC**: gRPC (Protocol Buffers)
- **Authentication**: JWT validation via middleware
- **Logging**: Structured logging with slog

---

## Technologies Used

| Technology | Version | Purpose |
|------------|---------|---------|
| Go | 1.24.5 | Core language |
| PostgreSQL | - | Profile data persistence |
| Goose | v3.24.3 | Database migrations |
| Gorilla Mux | v1.8.1 | HTTP routing |
| gRPC | v1.75.1 | RPC communication |
| Protocol Buffers | v1.36.9 | Service contracts |
| Confluent Kafka | v2.11.0 | Event consumption |
| golang-jwt | v5.3.0 | JWT validation |
| Google UUID | v1.6.0 | Unique identifiers |
| Gorilla Handlers | v1.5.2 | CORS middleware |

---

## API Endpoints

### HTTP REST API

| Method | Endpoint | Description | Auth Required | Path Params |
|--------|----------|-------------|---------------|------------|
| GET | `/profile/get` | Get user profile by UUID | Yes | - |
| PUT | `/update` | Update user profile | Yes | - |

### Response Examples

**Get Profile (200)**
```json
{
  "id": 42,
  "userId": "550e8400-e29b-41d4-a716-446655440000",
  "firstName": "John",
  "lastName": "Doe",
  "email": "john@example.com",
  "createdat": "2025-01-01T10:00:00Z"
}
```

**Update Profile (200)**
```json
{
  "id": 42,
  "userId": "550e8400-e29b-41d4-a716-446655440000",
  "firstName": "Jane",
  "lastName": "Smith",
  "email": "jane@example.com",
  "updatedat": "2025-01-02T15:30:00Z"
}
```

---

## gRPC Service Definition

### User Service RPC

```protobuf
service UserServiceRPC {
  rpc GetUser(GetUserRequest) returns (GetUserResponse);
  rpc UpdateUser(UpdateUserRequest) returns (UpdateUserResponse);
}

message GetUserRequest {
  string user_id = 1;
}

message GetUserResponse {
  int32 id = 1;
  string user_id = 2;
  string first_name = 3;
  string last_name = 4;
  string email = 5;
  string created_at = 6;
}
```

---

## Database Schema

### Profiles Table
```sql
CREATE TABLE profiles (
  id SERIAL PRIMARY KEY,
  user_id UUID UNIQUE NOT NULL,
  firstname VARCHAR(100),
  lastname VARCHAR(100),
  email VARCHAR(100) UNIQUE NOT NULL,
  metadata JSONB DEFAULT '{}',
  created_at TIMESTAMP DEFAULT NOW(),
  updated_at TIMESTAMP DEFAULT NOW()
);

CREATE INDEX idx_user_id ON profiles(user_id);
CREATE INDEX idx_email ON profiles(email);
```

---

## Event-Driven Flow

### UserCreated Event (from auth-service)
```
Kafka Topic: profiles

Event Structure:
{
  "userid": "550e8400-e29b-41d4-a716-446655440000",
  "firstName": "John",
  "lastName": "Doe",
  "email": "john.doe@example.com"
}

Processing:
1. User-service consumer reads event
2. Inserts new profile into profiles table
3. Creates metadata entry (empty JSON initially)
4. Acknowledges message to Kafka
```

### Data Flow
```
Auth-Service (publishes)
  ↓
Kafka Topic: profiles
  ↓
User-Service Consumer (listens)
  ↓
PostgreSQL profiles table
  ↓
gRPC Server (for other services)
```

---

## Configuration

### Environment Variables

```bash
# Database
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=postgres
DB_NAME=users_db
DB_SSLMODE=disable

# Kafka
KAFKA_BROKER=localhost:9092
KAFKA_GROUP_ID=user-service-group
KAFKA_TOPIC=profiles

# JWT
JWT_SECRET=your-secret-key

# CORS
CORS_ALLOWED_ORIGINS=http://localhost:3000,http://localhost:5173

# gRPC
GRPC_ADDR=0.0.0.0:50051
```

---

## Development & Testing

### Run Tests
```bash
cd user-service
go test ./...  # All packages
go test -v ./internal/service  # Specific package
go test -cover ./...  # With coverage
```

### Run Linting
```bash
golangci-lint run ./...
```

### Build & Run Locally
```bash
go build -o user-service ./cmd/main.go
./user-service
```

### Docker Build & Run
```bash
docker build -t user-service:latest .
docker run -p 9000:9000 -p 50051:50051 --env-file .env user-service:latest
```

### Test gRPC Endpoint (using grpcurl)
```bash
grpcurl -plaintext -d '{"user_id":"550e8400-e29b-41d4-a716-446655440000"}' \
  localhost:50051 UserServiceRPC.GetUser
```

---

## Error Codes

| Status | Error | Description |
|--------|-------|-------------|
| 200 | OK | Successful operation |
| 400 | Bad Request | Invalid input data |
| 401 | Unauthorized | Invalid or missing JWT |
| 404 | Not Found | User profile not found |
| 417 | Expectation Failed | Validation error |
| 500 | Internal Server Error | Database or server error |

---

## Service Dependencies

```
user-service
├── auth-service (consumes UserCreated events)
├── PostgreSQL (profiles table)
└── Kafka (event subscription)

Used by:
├── team-service (gRPC: GetUser)
├── event-service (gRPC: GetUser)
└── workout-service (gRPC: GetUser)
```

---

## Deployment

Kubernetes deployment automatically:

1. Creates ConfigMap for environment variables
2. Provisions PostgreSQL service
3. Sets up Kafka consumer group
4. Exposes both HTTP (9000) and gRPC (50051) ports
5. Implements health checks (readiness/liveness probes)

See [k8s/user-service/](../../k8s/) for manifests.

---

## Important Design Patterns

### 1. **Event-Driven Creation**
Unlike REST where you POST to create, this service creates profiles by consuming events. This decouples auth from profiles.

### 2. **Read Replicas via gRPC**
Other services don't call REST endpoints; they use gRPC for low-latency user lookups.

### 3. **Eventual Consistency**
User registered → Kafka event → Profile created (slight delay acceptable)

### 4. **Database Ownership**
Only user-service can write to profiles table. Other services query via gRPC or REST.

---

## Security Considerations

- All REST endpoints require JWT authentication
- gRPC calls use mTLS (TODO: implement)
- PostgreSQL user has minimal privileges (SELECT, INSERT, UPDATE only)
- Kafka consumer group isolation prevents duplicate processing
- TODO: Add audit logging for profile updates
- TODO: Implement GDPR data deletion endpoint

---

## Troubleshooting

- **Kafka consumer not consuming**: Check `KAFKA_BROKER`, `KAFKA_TOPIC`, `KAFKA_GROUP_ID`
- **Profile not created after registration**: Verify auth-service publishes to correct topic
- **gRPC connection refused**: Ensure service is listening on port 50051
- **JWT validation fails**: Check `JWT_SECRET` matches auth-service
- **Database locked**: Check for long-running migrations or transactions
