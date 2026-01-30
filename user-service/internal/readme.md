sequenceDiagram
    participant C as Client
    participant A as auth_service
    participant K as Kafka (Topic: user.events)
    participant U as user_service

    C->>A: POST /register (email, password)
    A->>A: Create user in auth_db
    A->>K: Publish UserCreated Event {user_id, email}
    K-->>U: Consumer receives event
    U->>U: Create profile in user_db
    A-->>C: 201 Created