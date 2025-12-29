# Event Service

## Overview

The **Event Service** manages sports events (games, practices, tournaments) associated with teams. It coordinates with team-service and user-service via gRPC to fetch team information and player details. Events are the scheduling and coordination hub where teams plan activities, track attendance, and manage event details.

**Port:** `7000` (HTTP)  
**Module:** `github.com/wycliff-ochieng`

---

## Key Features

### Core Capabilities

- **Event Creation**: Create events (games, practices, tournaments) for teams
- **Event Details Management**: Update event information (name, location, time)
- **Attendance Tracking**: Record and manage player attendance
- **Team Integration**: Fetch team and member info via gRPC
- **User Integration**: Get player profiles via gRPC
- **Event Queries**: Retrieve event details by ID with full context
- **JWT-Protected Routes**: All endpoints require Bearer token authentication

### Architecture

- **Framework**: Go with Gorilla Mux
- **Database**: PostgreSQL with Goose migrations
- **IPC**: gRPC for service-to-service communication
- **Authentication**: JWT validation via middleware
- **Logging**: Structured logging with slog

---

## Technologies Used

| Technology | Version | Purpose |
|------------|---------|---------|
| Go | 1.24.5 | Core language |
| PostgreSQL | v14+ | Event and attendance data |
| Goose | v3.25.0 | Database migrations |
| Gorilla Mux | v1.8.1 | HTTP routing |
| gRPC | v1.75.1 | Service communication |
| Protocol Buffers | v1.36.9 | gRPC contracts |
| golang-jwt | v5.3.0 | JWT validation |
| Google UUID | v1.6.0 | Unique identifiers |
| Gorilla Handlers | v1.5.2 | CORS middleware |

---

## API Endpoints

| Method | Endpoint | Description | Auth Required | Path/Query Params |
|--------|----------|-------------|---------------|-------------------|
| POST | `/api/events/new` | Create new event | Yes | - |
| GET | `/api/events/get/{event_id}` | Get event details | Yes | `event_id` |
| PUT | `/api/event/{event_id}` | Update event details | Yes | `event_id` |

### Request/Response Examples

**Create Event (POST /api/events/new)**
```json
{
  "eventId": "550e8400-e29b-41d4-a716-446655440000",
  "teamId": "660e8400-e29b-41d4-a716-446655440000",
  "name": "Championship Game",
  "eventType": "game",
  "location": "Central Sports Complex",
  "startTime": "2025-01-15T14:00:00Z",
  "endTime": "2025-01-15T16:00:00Z"
}

Response (200):
{
  "eventId": "550e8400-e29b-41d4-a716-446655440000",
  "teamId": "660e8400-e29b-41d4-a716-446655440000",
  "name": "Championship Game",
  "eventType": "game",
  "location": "Central Sports Complex",
  "startTime": "2025-01-15T14:00:00Z",
  "endTime": "2025-01-15T16:00:00Z",
  "createdBy": "770e8400-e29b-41d4-a716-446655440000",
  "createdAt": "2025-01-10T10:00:00Z"
}
```

**Get Event Details (GET /api/events/get/{event_id})**
```json
Response (200):
{
  "eventId": "550e8400-e29b-41d4-a716-446655440000",
  "teamId": "660e8400-e29b-41d4-a716-446655440000",
  "teamName": "Champions United",
  "name": "Championship Game",
  "eventType": "game",
  "location": "Central Sports Complex",
  "startTime": "2025-01-15T14:00:00Z",
  "endTime": "2025-01-15T16:00:00Z",
  "attendees": [
    {
      "userId": "880e8400-e29b-41d4-a716-446655440000",
      "firstName": "John",
      "lastName": "Doe",
      "status": "attending"
    }
  ],
  "createdAt": "2025-01-10T10:00:00Z"
}
```

---

## Database Schema

### Events Table
```sql
CREATE TABLE events (
  id SERIAL PRIMARY KEY,
  event_id UUID UNIQUE NOT NULL,
  team_id UUID NOT NULL,
  name VARCHAR(255),
  event_type VARCHAR(50),  -- game, practice, tournament, etc.
  location VARCHAR(255),
  start_time TIMESTAMP NOT NULL,
  end_time TIMESTAMP,
  created_by UUID,
  created_at TIMESTAMP DEFAULT NOW(),
  updated_at TIMESTAMP DEFAULT NOW()
);

CREATE INDEX idx_event_id ON events(event_id);
CREATE INDEX idx_team_id ON events(team_id);
CREATE INDEX idx_start_time ON events(start_time);
```

### Attendance Table
```sql
CREATE TABLE attendance (
  id SERIAL PRIMARY KEY,
  event_id UUID REFERENCES events(event_id),
  user_id UUID NOT NULL,
  status VARCHAR(50),  -- attending, absent, maybe
  check_in_time TIMESTAMP,
  created_at TIMESTAMP DEFAULT NOW(),
  updated_at TIMESTAMP DEFAULT NOW()
);

CREATE INDEX idx_attendance_event_id ON attendance(event_id);
CREATE INDEX idx_attendance_user_id ON attendance(user_id);
CREATE UNIQUE INDEX idx_attendance_unique ON attendance(event_id, user_id);
```

---

## Service Integration Flow

```
Event-Service
├── (gRPC Call) → Team-Service:50052
│   ├── GetTeamMembers (fetch team roster)
│   └── GetTeamDetails (validate team exists)
├── (gRPC Call) → User-Service:50051
│   └── GetUser (fetch player profile for attendance)
└── PostgreSQL
    ├── events table
    └── attendance table
```

---

## Event Types

| Type | Description | Example |
|------|-------------|---------|
| `game` | Competitive match | Championship, League Match |
| `practice` | Training session | Weekly Practice, Drills |
| `tournament` | Multi-team event | Cup Competition, League Finals |
| `training` | Skill-focused session | Shooting Drill, Fitness Test |
| `friendly` | Non-competitive match | Scrimmage, Exhibition |

---

## Configuration

### Environment Variables

```bash
# Database
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=postgres
DB_NAME=events_db
DB_SSLMODE=disable

# JWT
JWT_SECRET=your-secret-key

# CORS
CORS_ALLOWED_ORIGINS=http://localhost:3000,http://localhost:5173

# gRPC Endpoints
TEAM_SERVICE_GRPC_ADDR=localhost:50052
USER_SERVICE_GRPC_ADDR=localhost:50051
```

---

## Development & Testing

### Run Tests
```bash
cd event-service
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
go build -o event-service ./cmd/main.go
./event-service
```

### Docker Build & Run
```bash
docker build -t event-service:latest .
docker run -p 7000:7000 --env-file .env event-service:latest
```

---

## Error Codes

| Status | Error | Description |
|--------|-------|-------------|
| 200 | OK | Successful operation |
| 400 | Bad Request | Invalid input or UUID parsing error |
| 401 | Unauthorized | Missing or invalid JWT |
| 417 | Expectation Failed | Missing required fields (teamId, eventType) |
| 424 | Failed Dependency | Team-service or user-service unavailable |
| 500 | Internal Server Error | Database or server error |

---

## Service Dependencies

```
event-service
├── (gRPC) → team-service:50052
│   ├── GetTeamMembers
│   └── ValidateTeam
├── (gRPC) → user-service:50051
│   └── GetUser (for attendance player info)
├── PostgreSQL (events, attendance tables)
└── auth-service (JWT validation via middleware)

Used by:
└── Frontend (HTTP REST API)
```

---

## Deployment

Kubernetes manifests in [k8s/event-service/](../../k8s/):

1. ConfigMap for environment variables and gRPC addresses
2. PostgreSQL persistent volume with migrations
3. Service exposing HTTP:7000
4. Health checks (readiness/liveness probes)
5. Resource limits and requests

---

## Attendance Management

### Mark Attendance
```
1. Event occurs on scheduled date
2. Team members check in via mobile/web app
3. Service records attendance with timestamp
4. Coach/manager can view attendance report
```

### Attendance Status
- `attending`: Player confirmed attendance
- `absent`: Player did not attend
- `maybe`: Player tentative status

---

## Security Considerations

- JWT validation on all endpoints
- User can only see/modify events for their teams
- gRPC calls to services validate team membership
- CORS configured per environment
- TODO: Implement event permissions (who can modify)
- TODO: Add audit logging for event changes
- TODO: Implement event cancellation with notifications

---

## Common Issues & Troubleshooting

- **gRPC service connection refused**: Verify `TEAM_SERVICE_GRPC_ADDR` and `USER_SERVICE_GRPC_ADDR`
- **Cannot create event - team not found**: Ensure team exists in team-service
- **UUID parsing error in path**: Validate path parameters are valid UUIDs
- **Attendance records missing**: Check attendance table for data integrity
- **JWT validation fails**: Verify `JWT_SECRET` matches auth-service