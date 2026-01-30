#  Auth-Service

## Authentication && Authorization 

- **Authentication** : Who are you? This is done through username/email and password
- **Authorization** : What are you allowed to do. Process of checkingn verified users permisssions(**Their roles**)

## JSON Web Tokens
- JWT : Compact self contained string that proves a users is who they say thay are. Has three parts :(payload,signature,haeder)
    (a)Payload(claims) : Contain data about a user their ID,roles and expiry time. NB : payload is url Base64 encoded not encrypted. Never put sensitive data in the payload
    (b)Siganture : This is the security . Created by hashing the header & payload with a secret known to only you and the server.If someone chnages the payload or header the signature will no longer match and the token will be rejected
    (c)Header : 

## Password Hashing(bcrypt)
- Though SHALL NOT store plain passwords in a database. We will always store hashes, in that when a user logins we compare the password provided with hash
- **why bcrypt** : industry standard due to it being slow and provides a "salt" which is resistant to brute-force attacks and rainbow attacks

## Middleware 
- Go web server, middleware is a function that wraps an HTTP handler. It sits between the incoming request and your main logic. Our authentication middleware will act as a gatekeeper: it will inspect every incoming request for a valid JWT. If the token is valid, it passes the request on; if not, it rejects it with a **401 Unauthorized error.**


## Authentication Data Flow
`User Registration Flow`
-Client (Frontend) sends a POST /register request to the auth_service with the user's email and password.->

The auth_service hashes the password securely using bcrypt.->

It then stores the new user in the auth_db: INSERT INTO users (email, password_hash).->

Once the user is successfully created, auth_service publishes a "UserCreated" event — useful for notifying other services like user_service or email_service.

The auth_service responds to the

`User Login flow`
Client sends a POST /login request with email and password.

auth_service queries the database:
SELECT password_hash FROM users WHERE email=?

If a matching record is found, it compares the submitted password with the stored hash using bcrypt.Compare.

If the password matches:
4. The service retrieves the user's roles:
SELECT roles FROM user_roles WHERE user_id=?
5. It creates a JWT containing the user's ID, roles, and expiration time.
6. Returns 200 OK to the client along with the JWT.

If the password doesn't match:
4. Returns 401 Unauthorized to the client.

`protected Routes`(viewing Teams)
Client tries to access a protected route (e.g., GET /api/teams) and includes the JWT in the Authorization header.

The API Gateway receives the request and delegates JWT validation to the auth_service, or does it inline via middleware.

auth_service verifies the JWT’s signature and expiration.

If the JWT is valid:
4. The API Gateway forwards the request to the appropriate microservice (e.g., team_service).
5. The client receives a successful response — the request proceeds.

If the JWT is invalid or expired:
4. API Gateway responds with 401 Unauthorized to the client.

✅ Summary
Flow	Status Code	Notes
Register	201 Created	User stored securely with hashed password
Login (Success)	200 OK	Returns JWT for use in protected routes
Login (Fail)	401 Unauthorized	Wrong credentials
Protected Access (Valid Token)	Proceeds to service	JWT is accepted
Protected Access (Invalid Token)	401 Unauthorized	JWT rejected



## Schema Migration 

- we will be using golang goose
# Install the goose CLI tool
`go install github.com/pressly/goose/v3/cmd/goose@latest`

here we define the database connection

