## User Service

- **Core Philosophy**
- Its one and only job is to manage user profile data
- Decoupled from authentication service( does not care about, passwords ,login)
- It is the exclusive owner of the profiles table in users_db, no other service is allowed to touch this table directly
- Event Driven- Creates new profile by listening to events from the auth_service, this makes system resilient in that if user-service it down, events are put in a queue and executed when the service is back

#### service folder structure 

user_service/
├── cmd/
│   └── main.go              # Main application entry point
├── internal/
│   ├── api/
│   │   ├── handlers.go      # HTTP handlers (GET /me, PUT /me)
│   │   └── router.go        # Sets up the HTTP routes and middleware
│   ├── config/
│   │   └── config.go        # Configuration loading (DB strings, etc.)
│   ├── consumers/
│   │   └── user_events.go   # RabbitMQ/Kafka consumer for 'UserCreated' events
│   ├── models/
│   │   └── user.go          # The User Profile struct
│   ├── repository/
│   │   └── postgres.go      # All database query logic (interface and implementation)
│   └── rpc/
│       └── server.go        # The gRPC server implementation for other services
├── db/
│   ├── migrations/
│   │   └── ..._create_profiles_table.sql
│   └── dbconf.yml
├── proto/
│   └── user.proto           # gRPC contract file
├── go.mod
├── go.sum
└── Dockerfile


