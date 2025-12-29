# Workout Service

## Overview

The **Workout Service** manages workout programs, exercises, and related media (images, videos). It is responsible for storing workout definitions, exercises, and handling file uploads to MinIO object storage. Coaches and managers create workouts with ordered exercises, while the service manages presigned URLs for secure media uploads and maintains metadata about workout media.

**Port:** `3000` (HTTP)  
**Module:** `github.com/wycliff-ochieng`

---

## Key Features

### Core Capabilities

- **Workout Management**: Create and retrieve structured workout programs
- **Exercise Library**: Define exercises with instructions and descriptions
- **Workout Composition**: Link exercises to workouts with sets, reps, and order
- **Media Upload**: Generate presigned URLs for secure image/video uploads to MinIO
- **Media Management**: Track media metadata and associate with workouts/exercises
- **File Storage**: Integration with MinIO S3-compatible object storage
- **User Integration**: Fetch user profiles via gRPC for ownership/permissions
- **JWT Protection**: All endpoints require Bearer token authentication
- **Pagination**: List workouts with cursor-based pagination and search

### Architecture

- **Framework**: Go with Gorilla Mux
- **Database**: PostgreSQL with Goose migrations
- **File Storage**: MinIO S3-compatible object storage
- **IPC**: gRPC for user-service communication
- **Authentication**: JWT validation via middleware
- **Logging**: Structured logging with slog

---

## Technologies Used

| Technology | Version | Purpose |
|------------|---------|---------|
| Go | 1.24.5 | Core language |
| PostgreSQL | v14+ | Workout/exercise/media metadata |
| Goose | v3.26.0 | Database migrations |
| Gorilla Mux | v1.8.1 | HTTP routing |
| MinIO Go | v7.0.97 | S3-compatible object storage |
| gRPC | v1.75.1 | User-service communication |
| Protocol Buffers | v1.36.9 | gRPC contracts |
| golang-jwt | v5.3.0 | JWT validation |
| Google UUID | v1.6.0 | Unique identifiers |
| Gorilla Handlers | v1.5.2 | CORS middleware |

---

## API Endpoints

| Method | Endpoint | Description | Auth Required | Query Params |
|--------|----------|-------------|---------------|-------------|
| POST | `/api/` | Create workout | Yes | - |
| GET | `/api/workout` | List workouts | Yes | `limit`, `cursor`, `search` |
| POST | `/api/exercise` | Create exercise | Yes | - |
| GET | `/api/exercise/fetch` | List exercises | Yes | - |
| POST | `/api/media/presigned-url` | Get upload URL | Yes | - |
| POST | `/api/media/upload-complete` | Complete upload | Yes | - |

### Request/Response Examples

**Create Workout (POST /api/)**
```json
{
  "name": "Full Body Strength - Phase 1, Day A",
  "description": "Beginner-focused workout targeting major muscle groups",
  "category": "Strength Training",
  "exercises": [
    {
      "exerciseid": "d1029346-3033-4e7b-8860-598f0d32beaa",
      "order": 1,
      "sets": 5,
      "reps": 5
    },
    {
      "exerciseid": "5bf3f4bc-9df8-4092-bdac-9f1da8c789af",
      "order": 2,
      "sets": 3,
      "reps": 8
    }
  ]
}

Response (200):
{
  "workoutid": "550e8400-e29b-41d4-a716-446655440000",
  "name": "Full Body Strength - Phase 1, Day A",
  "description": "Beginner-focused workout...",
  "category": "Strength Training",
  "created_by": "660e8400-e29b-41d4-a716-446655440000",
  "created_at": "2025-01-10T10:00:00Z",
  "exercises": [...]
}
```

**Create Exercise (POST /api/exercise)**
```json
{
  "name": "Barbell Squat",
  "description": "Compound lower body exercise",
  "instructions": "Stand with feet shoulder-width apart, lower body until thighs are parallel..."
}

Response (200):
{
  "exerciseid": "d1029346-3033-4e7b-8860-598f0d32beaa",
  "name": "Barbell Squat",
  "description": "Compound lower body exercise",
  "instructions": "...",
  "created_by": "660e8400-e29b-41d4-a716-446655440000",
  "created_at": "2025-01-10T10:00:00Z"
}
```

**Get Presigned URL (POST /api/media/presigned-url)**
```json
{
  "parentId": "550e8400-e29b-41d4-a716-446655440000",
  "parentType": "workout",
  "filename": "squat_demo.mp4",
  "mimeType": "video/mp4"
}

Response (200):
{
  "uploadUrl": "https://minio.example.com/presigned-url...",
  "expiresIn": 3600
}
```

**List Workouts (GET /api/workout?limit=25&cursor=abc123)**
```json
Response (200):
{
  "workouts": [
    {
      "workoutid": "550e8400-e29b-41d4-a716-446655440000",
      "name": "Full Body Strength",
      "description": "...",
      "created_at": "2025-01-10T10:00:00Z",
      "exerciseCount": 5
    }
  ],
  "nextCursor": "xyz789"
}
```

---

## Database Schema

### Workouts Table
```sql
CREATE TABLE workouts (
  id SERIAL PRIMARY KEY,
  workout_id UUID UNIQUE NOT NULL,
  created_by UUID NOT NULL,
  name VARCHAR(255),
  description TEXT,
  category VARCHAR(100),
  created_at TIMESTAMP DEFAULT NOW(),
  updated_at TIMESTAMP DEFAULT NOW()
);

CREATE INDEX idx_workout_id ON workouts(workout_id);
CREATE INDEX idx_created_by ON workouts(created_by);
```

### Exercises Table
```sql
CREATE TABLE exercises (
  id SERIAL PRIMARY KEY,
  exercise_id UUID UNIQUE NOT NULL,
  created_by UUID NOT NULL,
  name VARCHAR(255),
  description TEXT,
  instructions TEXT,
  created_at TIMESTAMP DEFAULT NOW(),
  updated_at TIMESTAMP DEFAULT NOW()
);

CREATE INDEX idx_exercise_id ON exercises(exercise_id);
CREATE INDEX idx_created_by ON exercises(created_by);
```

### Workout-Exercise Mapping Table
```sql
CREATE TABLE workout_exercises (
  id SERIAL PRIMARY KEY,
  workout_id UUID REFERENCES workouts(workout_id),
  exercise_id UUID REFERENCES exercises(exercise_id),
  "order" INT NOT NULL,
  sets INT,
  reps INT,
  created_at TIMESTAMP DEFAULT NOW()
);

CREATE UNIQUE INDEX idx_workout_exercise_unique 
  ON workout_exercises(workout_id, exercise_id);
CREATE INDEX idx_workout_exercises_order 
  ON workout_exercises(workout_id, "order");
```

### Media Table
```sql
CREATE TABLE media (
  id SERIAL PRIMARY KEY,
  media_id UUID UNIQUE NOT NULL,
  parent_id UUID NOT NULL,
  parent_type VARCHAR(50),  -- workout, exercise
  filename VARCHAR(255),
  mime_type VARCHAR(100),
  storage_path VARCHAR(255),
  size_bytes BIGINT,
  upload_status VARCHAR(50),  -- pending, completed
  created_at TIMESTAMP DEFAULT NOW(),
  updated_at TIMESTAMP DEFAULT NOW()
);

CREATE INDEX idx_parent_id ON media(parent_id);
CREATE INDEX idx_upload_status ON media(upload_status);
```

---

## Media Upload Flow

### Step 1: Request Presigned URL
```
Client → POST /api/media/presigned-url
{
  "parentId": "workout-uuid",
  "parentType": "workout",
  "filename": "demo.mp4",
  "mimeType": "video/mp4"
}
       ↓
Service validates mime type (video/mp4, image/jpeg, image/png only)
       ↓
Generates presigned URL from MinIO (valid 1 hour)
       ↓
Returns URL to client
```

### Step 2: Client Uploads File
```
Client → PUT to presigned URL (direct to MinIO)
       ↓
MinIO receives and stores file
```

### Step 3: Notification (Optional)
```
MinIO bucket notification → Event queue
       ↓
Webhook triggers workout-service
```

### Step 4: Mark Upload Complete
```
Client → POST /api/media/upload-complete
{
  "mediaId": "media-uuid",
  "filesize": 1024000
}
       ↓
Service updates media.upload_status = "completed"
       ↓
Confirms to client
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
DB_NAME=workouts_db
DB_SSLMODE=disable

# MinIO Storage
MINIO_ENDPOINT=localhost:9000
MINIO_ACCESS_KEY=minioadmin
MINIO_SECRET_KEY=minioadmin
MINIO_BUCKET=workouts
MINIO_USE_SSL=false

# JWT
JWT_SECRET=your-secret-key

# CORS
CORS_ALLOWED_ORIGINS=http://localhost:3000,http://localhost:5173

# gRPC
USER_SERVICE_GRPC_ADDR=localhost:50051
```

---

## Supported Media Types

| Type | MIME Types | Max Size |
|------|-----------|----------|
| Video | `video/mp4`, `video/quicktime` | 500 MB |
| Image | `image/jpeg`, `image/png`, `image/webp` | 10 MB |

---

## Development & Testing

### Run Tests
```bash
cd workout-service
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
go build -o workout-service ./cmd/main.go
./workout-service
```

### Docker Build & Run
```bash
docker build -t workout-service:latest .
docker run -p 3000:3000 --env-file .env workout-service:latest
```

### Test Media Upload (curl)
```bash
# 1. Get presigned URL
curl -X POST http://localhost:3000/api/media/presigned-url \
  -H "Authorization: Bearer <token>" \
  -H "Content-Type: application/json" \
  -d '{"parentId":"...", "parentType":"workout", "filename":"demo.mp4", "mimeType":"video/mp4"}'

# 2. Upload file using returned URL
curl -X PUT "<presigned-url>" -T demo.mp4

# 3. Mark complete
curl -X POST http://localhost:3000/api/media/upload-complete \
  -H "Authorization: Bearer <token>" \
  -d '{"mediaId":"...", "filesize":1024000}'
```

---

## Error Codes

| Status | Error | Description |
|--------|-------|-------------|
| 200 | OK | Successful operation |
| 400 | Bad Request | Invalid input or validation error |
| 401 | Unauthorized | Missing or invalid JWT |
| 417 | Expectation Failed | Invalid mime type or missing required fields |
| 424 | Failed Dependency | User-service unavailable |
| 500 | Internal Server Error | Database or server error |

---

## Service Dependencies

```
workout-service
├── (gRPC) → user-service:50051
│   └── Validate user permissions
├── PostgreSQL (workouts, exercises, media tables)
├── MinIO (object storage for media)
└── auth-service (JWT validation via middleware)

Used by:
└── Frontend (HTTP REST API)
```

---

## Deployment

Kubernetes manifests in [k8s/workout-service/](../../k8s/):

1. ConfigMap for MinIO and database config
2. PostgreSQL persistent volume
3. MinIO service reference (externally managed)
4. Service exposing HTTP:3000
5. Health checks (readiness/liveness probes)
6. Resource limits and requests

---

## Performance Optimization

### Pagination Strategy
- **Cursor-based**: Uses `created_at` timestamp for efficient pagination
- **Limits**: Default 25, max 100 results per page
- **Search**: Full-text search on workout name and description

### Caching Opportunities
- Exercise list (rarely changes, ~5 min TTL)
- User permissions (per-request validation)
- MinIO presigned URLs (not cached)

---

## Security Considerations

- JWT validation on all endpoints
- MIME type validation for uploads
- Presigned URL expiration (1 hour)
- User ownership verification
- CORS configured per environment
- TODO: Implement virus scanning for uploads
- TODO: Add rate limiting for presigned URL generation
- TODO: Audit logging for media deletions

---

## Common Issues & Troubleshooting

- **MinIO connection refused**: Verify `MINIO_ENDPOINT`, `MINIO_ACCESS_KEY`, `MINIO_SECRET_KEY`
- **Presigned URL invalid**: Check MinIO bucket exists and credentials are correct
- **Invalid mime type error**: Verify uploaded file MIME type is in allowed list
- **Upload timeout**: Increase presigned URL expiration for large files
- **gRPC user-service unavailable**: Check `USER_SERVICE_GRPC_ADDR`
- **Pagination cursor invalid**: Use cursor from previous response, don't guess

        }
    ]
}