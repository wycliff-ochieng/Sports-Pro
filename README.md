## Sports Management Platform - Microservices Backend
This repository contains the backend source code for the Sports Management Platform, a high-performance, scalable application built using a microservices architecture in Go. The platform is designed to handle user authentication, team management, event scheduling, and more, with a focus on security, resilience, and maintainability.

### 1. High-Level Architecture
The platform follows a classic microservices pattern where each service is independently deployable, scalable, and owns its own data. Services communicate via a combination of synchronous gRPC calls for immediate requests and an asynchronous Kafka event bus for decoupled communication and data synchronization.

### 2. Core Technologies
- Language: Go (Golang)
- Databases:
- PostgreSQL: Primary relational database for core service data.
- Redis: In-memory cache for sessions and frequently accessed data.
- Communication:
    * REST: For external, client-facing APIs.
    * gRPC: For high-performance, internal service-to-service communication.
    * Kafka: For asynchronous, event-driven communication between services.
- Containerization: Docker & Docker Compose for local development. Docker and Kubernetes for Production
- Database Migrations: Goose for managing SQL schema evolution.
- Authentication: JSON Web Tokens (JWT).

### 3. Microservice Overview
**auth_service**
* Responsibilities: Manages user registration, login, and password hashing. It is the sole creator of JWTs and the source of truth for a user's roles.
* Database: auth_db
* Publishes Events:
    1. UserCreated: Announces a new user has registered, triggering profile creation in user_service.
**user_service**
* Responsibilities: Manages user profile data (name, contact info, etc.). It acts as a central "phone book" for other services.
* Database: user_db
* Consumes Events:
    - UserCreated (from auth_service): To create a new, empty user profile.
* Publishes Events:
    - UserProfileUpdated: To announce changes to a user's name, allowing other services to update denormalized data.
* Provides gRPC API:
    - GetUserProfile(user_id): Allows other services to fetch a user's details.
**team_service**
* Responsibilities: Manages teams, their rosters, and the roles of users within a team (e.g., Coach, Player).
* Database: team_db
* Consumes gRPC from:
    - user_service: To validate users and get their names when adding them to a team.
* Provides gRPC API:
    - IsUserOnTeam(user_id, team_id): Allows other services to perform authorization checks.
**event_service**
* Responsibilities: Manages the creation, scheduling, and attendance for events like games and practices.
* Database: event_db
* Consumes Events:
    - TeamUpdated, TeamMemberAdded (from team_service): To keep its denormalized data and attendance lists in sync.
* Consumes gRPC from:
    - team_service: To authorize actions (e.g., "Is this user a coach of the team for this event?").
    - user_service: To get the latest user details for attendance lists.

### 4. Communication Patterns
***Synchronous: gRPC**
- Used for request/response cycles where a service needs an immediate answer from another service to complete its task.
- Example: When team_service adds a player, it must synchronously call user_service to confirm the user exists before committing the database transaction.
**Asynchronous: Kafka**
- Used for fire-and-forget events to decouple services. The producing service does not wait for a response.
- Example: When auth_service registers a new user, it publishes a UserCreated event. It doesn't know or care if user_service is online to process it immediately. The event will wait in the Kafka topic until the consumer is ready, ensuring resilience