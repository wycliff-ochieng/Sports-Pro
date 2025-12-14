package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"github.com/wycliff-ochieng/internal/filestore"
	"github.com/wycliff-ochieng/internal/models"
	"github.com/wycliff-ochieng/internal/service"
	auth "github.com/wycliff-ochieng/sports-common-package/middleware"
)

func TestMediaPresignedURL(t *testing.T) {
	// Setup Logger
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	// Setup Mock Server to prevent MinIO client from hanging on connection attempts
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	// Extract host:port from server.URL
	endpoint := strings.TrimPrefix(server.URL, "http://")

	// Setup MinIO Client
	minioClient, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4("access", "secret", ""),
		Secure: false, // httptest server is HTTP
		Region: "us-east-1",
	})
	if err != nil {
		t.Fatalf("Failed to create minio client: %v", err)
	}

	fs := &filestore.FileStore{
		Client: minioClient,
		Bucket: "test-bucket",
	}

	// Setup Service (Concrete)
	// We pass nil for DB and UserClient as they are not used in GeneratePresignedURL
	ws := service.NewWorkoutService(nil, nil, fs)

	// Setup Handler
	handler := NewWorkoutHandler(logger, ws)

	// Define test cases
	tests := []struct {
		name           string
		reqBody        models.PresignedURLReq
		expectedStatus int
	}{
		{
			name: "Success",
			reqBody: models.PresignedURLReq{
				ParentID:   uuid.New(),
				ParentType: "workout",
				Filename:   "test.jpg",
				MimeType:   "image/jpeg",
			},
			expectedStatus: http.StatusOK,
		},
		{
			name: "Invalid Parent Type",
			reqBody: models.PresignedURLReq{
				ParentID:   uuid.New(),
				ParentType: "invalid",
				Filename:   "test.jpg",
				MimeType:   "image/jpeg",
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "Invalid Mime Type",
			reqBody: models.PresignedURLReq{
				ParentID:   uuid.New(),
				ParentType: "workout",
				Filename:   "test.jpg",
				MimeType:   "application/pdf",
			},
			expectedStatus: http.StatusExpectationFailed,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create Request
			body, _ := json.Marshal(tt.reqBody)
			req := httptest.NewRequest(http.MethodPost, "/api/media/presigned-url", bytes.NewReader(body))

			// Add User ID to Context (Simulating Auth Middleware)
			// The middleware uses a string for the UUID value in the context
			ctx := context.WithValue(req.Context(), auth.UserUUIDKey, uuid.New().String())
			req = req.WithContext(ctx)

			w := httptest.NewRecorder()

			// Execute
			handler.MediaPresignedURL(w, req)

			// Assert
			if w.Code != tt.expectedStatus {
				t.Errorf("expected status %d, got %d. Body: %s", tt.expectedStatus, w.Code, w.Body.String())
			}
		})
	}
}
