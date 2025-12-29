# Team Service

## Overview

The **Team Service** manages team creation, membership, and team-related operations. Teams are the core organizational units where coaches, managers, and players collaborate. The service uses gRPC to fetch user and team information from other services, enforces role-based access control, and publishes team events to Kafka for downstream processing.

**Port:** `4000` (HTTP)  
**gRPC Port:** `50052`  
**Module:** `github/wycliff-ochieng`

---

## Key Features

### Core Capabilities

- **Team Creation**: Create teams with sport type, description, and metadata
- **Team Management**: Update team details (name, description)
- **Member Management**: Add/remove team members with role-based access
- **Team Roster**: Retrieve full team membership and member details
- **RBAC Enforcement**: Only coaches/managers can modify teams
- **gRPC Integration**: Calls user-service to validate members and fetch user info
- **Event Publishing**: Publishes team events to Kafka (team_events topic)
- **Service Discovery**: Communicates with user-service via gRPC for member validation

### Architecture

- **Framework**: Go with Gorilla Mux + gRPC
- **Database**: PostgreSQL with Goose migrations
- **Event Queue**: Apache Kafka/Confluent Kafka (Producer)
- **IPC**: gRPC for user-service communication
- **Authentication**: JWT validation + Role-based access control
- **Logging**: Structured logging with slog

---

## Technologies Used

| Technology | Version | Purpose |
|------------|---------|---------|
| Go | 1.24.5 | Core language |
| PostgreSQL | - | Team and membership data |
| Goose | v3.24.3 | Database migrations |
| Gorilla Mux | v1.8.1 | HTTP routing |
| gRPC | v1.75.1 | User-service communication |
| Protocol Buffers | v1.36.9 | gRPC contracts |
| Confluent Kafka | v2.11.1 | Event publishing |
| golang-jwt | v5.3.0 | JWT validation |
| Google UUID | v1.6.0 | Unique identifiers |
| Gorilla Handlers | v1.5.2 | CORS middleware |

---

## API Endpoints

| Method | Endpoint | Description | Auth Required | Roles Required | Path/Query Params |
|--------|----------|-------------|---------------|----------------|------------------|
| POST | `/api/teams` | Create team | Yes | coach, manager, player | - |
| GET | `/api/get/teams` | List user's teams | Yes | - | - |
| GET | `/api/team/{team_id}` | Get team details | Yes | - | `team_id` |
| PUT | `/api/team/{team_id}/update` | Update team | Yes | coach, manager, player | `team_id` |
| POST | `/api/team/{team_id}/add` | Add team member | Yes | coach, manager, player | `team_id` |
| GET | `/api/team/{team_id}/members` | Get team roster | Yes | - | `team_id` |
| PUT | `/api/team/{teamid}/members/{user_id}/update` | Update member | Yes | coach, manager | `teamid`, `user_id` |
| DELETE | `/api/team/{teamid}/member/{user_id}/delete` | Remove member | Yes | coach, manager | `teamid`, `user_id` |

### Request/Response Examples

**Create Team (POST /api/teams)**
```json
{
  "teamid": "550e8400-e29b-41d4-a716-446655440000",
  "name": "Champions United",
  "sport": "football",
  "description": "Our championship-winning team",
  "createdat": "2025-01-01T10:00:00Z",
  "updatedat": "2025-01-01T10:00:00Z"
}

Response (200):
{
  "teamid": "550e8400-e29b-41d4-a716-446655440000",
  "name": "Champions United",
  "sport": "football",
  "description": "Our championship-winning team",
  "creatorid": "660e8400-e29b-41d4-a716-446655440000",
  "created_at": "2025-01-01T10:00:00Z",
  "updated_at": "2025-01-01T10:00:00Z"
}
```

**Add Team Member**
```json
{
  "user_id": "770e8400-e29b-41d4-a716-446655440000",
  "role": "player"
}
```

---

## gRPC Service Definition

### Team Service RPC

```protobuf
service TeamRPC {
  rpc CreateTeam(CreateTeamRequest) returns (CreateTeamResponse);
  rpc GetTeamMembers(GetTeamMembersRequest) returns (GetTeamMembersResponse);
  rpc AddTeamMember(AddTeamMemberRequest) returns (AddTeamMemberResponse);
}

message CreateTeamRequest {
  string team_id = 1;
  string name = 2;
  string sport = 3;
  string description = 4;
}

message CreateTeamResponse {
  string team_id = 1;
  string name = 2;
  string creator_id = 3;
  string created_at = 4;
}

message GetTeamMembersRequest {
  string team_id = 1;
}

message GetTeamMembersResponse {
  repeated TeamMember members = 1;
}

message TeamMember {
  string user_id = 1;
  string first_name = 2;
  string last_name = 3;
  string role = 4;
  string joined_at = 5;
}
```

---

## Database Schema

### Teams Table
```sql
CREATE TABLE teams (
  id SERIAL PRIMARY KEY,
  team_id UUID UNIQUE NOT NULL,
  creator_id UUID NOT NULL,
  name VARCHAR(200),
  sport VARCHAR(100),
  description TEXT,
  created_at TIMESTAMP DEFAULT NOW(),
  updated_at TIMESTAMP DEFAULT NOW()
);

CREATE INDEX idx_team_id ON teams(team_id);
CREATE INDEX idx_creator_id ON teams(creator_id);
```

### Team Members Table
```sql
CREATE TABLE team_members (
  id SERIAL PRIMARY KEY,
  team_id UUID REFERENCES teams(team_id),
  user_id UUID NOT NULL,
  role VARCHAR(50),  -- player, coach, manager
  joined_at TIMESTAMP DEFAULT NOW(),
  updated_at TIMESTAMP DEFAULT NOW()
);

CREATE INDEX idx_team_members_team_id ON team_members(team_id);
CREATE INDEX idx_team_members_user_id ON team_members(user_id);
CREATE UNIQUE INDEX idx_team_members_unique ON team_members(team_id, user_id);
```

---

## Event-Driven Integration

### Team Events Published (Kafka)
Topic: `team_events`

```json
{
  "event_type": "TEAM_CREATED",
  "team_id": "550e8400-e29b-41d4-a716-446655440000",
  "team_name": "Champions United",
  "creator_id": "660e8400-e29b-41d4-a716-446655440000",
  "timestamp": "2025-01-01T10:00:00Z"
}
```

### Service Communication Flow
```
Team-Service
├── (gRPC Call) → User-Service:50051
│   └── Validate user exists before adding to team
├── (Kafka Publish) → Kafka Broker
│   └── Publish team events for event-service consumption
└── PostgreSQL (teams, team_members tables)
```

---

## Role-Based Access Control (RBAC)

| Endpoint | Coach | Manager | Player | Regular User |
|----------|-------|---------|--------|--------------|
| Create Team | Yes | Yes | Yes | None |
| Update Team | Yes | Yes | None | None |
| Add Member | Yes | Yes | None | None |
| Remove Member | Yes | Yes | None | None |
| View Team | Yes | Yes | Yes | Yes |
| View Roster | Yes | Yes | Yes | Yes |

---

## Configuration

### Environment Variables

```bash
# Database
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=postgres
DB_NAME=teams_db
DB_SSLMODE=disable

# Kafka
KAFKA_BROKER=localhost:9092

# JWT
JWT_SECRET=your-secret-key

# CORS
CORS_ALLOWED_ORIGINS=http://localhost:3000,http://localhost:5173

# gRPC
USER_SERVICE_GRPC_ADDR=localhost:50051  # user-service endpoint
GRPC_ADDR=0.0.0.0:50052  # this service's gRPC port
```

---

## Development & Testing

### Run Tests
```bash
cd team-service
go test ./...
go test -v ./internal/service
go test -cover ./...
```

### Run Linting
```bash
golangci-lint run ./...
```

### Build & Run
```bash
go build -o team-service ./cmd/main.go
./team-service
```

### Docker Build & Run
```bash
docker build -t team-service:latest .
docker run -p 4000:4000 -p 50052:50052 --env-file .env team-service:latest
```

---

## Error Codes

| Status | Error | Description |
|--------|-------|-------------|
| 200 | OK | Successful operation |
| 400 | Bad Request | Invalid input or UUID parsing error |
| 401 | Unauthorized | Missing or invalid JWT |
| 403 | Forbidden | User lacks required role |
| 404 | Not Found | Team or member not found |
| 424 | Failed Dependency | User-service unavailable or user validation failed |
| 500 | Internal Server Error | Database or server error |

---

## Service Dependencies

```
team-service
├── (gRPC) → user-service:50051
│   └── Validate users exist before adding
├── PostgreSQL (teams, team_members tables)
├── Kafka (publish team_events)
└── auth-service (JWT validation via middleware)

Used by:
├── event-service (gRPC: GetTeamMembers)
└── Frontend (HTTP REST API)
```

---

## Deployment

Kubernetes manifests in [k8s/team-service/](../../k8s/):

1. ConfigMap for environment variables
2. PostgreSQL persistent volume
3. Service exposing HTTP:4000 and gRPC:50052
4. Health checks (readiness/liveness probes)
5. Resource limits and requests

---

## Security Considerations

- JWT validation on all endpoints
- Role-based authorization per endpoint
- User validation via gRPC before team membership
- CORS configured per environment
- TODO: Implement audit logging for team operations
- TODO: Add rate limiting for team creation
- TODO: Implement team invitation system (instead of direct add)

---

## Common Issues & Troubleshooting

- **gRPC user-service connection refused**: Check `USER_SERVICE_GRPC_ADDR` environment variable
- **Cannot add member - user validation failed**: Verify user exists in user-service
- **UUID parsing error**: Ensure path parameters are valid UUIDs
- **Kafka publish fails**: Verify `KAFKA_BROKER` is reachable
- **Role validation fails**: Check JWT contains valid roles in claims
