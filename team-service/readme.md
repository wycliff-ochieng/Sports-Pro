### Team Service

1. GET/teams/me/ - > teams for the current profile
2. Get/teams/id/ -> team details
3. POST/team/ -> Create Team (coaches)
4. PUT/team/id -> Update team (coaches & managers)


**Endpoints for Teams**
### Method	Endpoint	Description	RBAC Middleware

1. POST	**/api/teams**	Creates a new team. The creator is automatically made the 'COACH'.	AuthZ("COACH", "ADMIN")
2. GET	**/api/teams/me**	Gets a list of all teams the current user is a member of.	None (Any role can see their teams)
3. GET	**/api/teams/{teamId}**	Gets the detailed public profile of a single team.	None (Any role can view a team)
4. PUT	**/api/teams/{teamId}**	Updates a team's details (name, description, etc.).	AuthZ("COACH", "ADMIN")
5. DELETE	**/api/teams/{teamId}**	Deletes a team. A highly destructive action.	AuthZ("ADMIN")

**Endpoints for Team Members**

- These endpoints operate on the relationship between users and teams.

### Method	Endpoint	Description	RBAC / Authorization Logic

1. POST	**/api/teams/{teamId}/members**	Adds a user to a team. The request body contains { "user_id": "...", "role": "..." }.	AuthZ("COACH", "ADMIN") + Ownership Check in the service layer.
2. GET	**/api/teams/{teamId}/members**	Gets the full roster (list of members) for a specific team.	Membership Check in the service layer. (Must be a member to see the roster).
3. PUT	**/api/teams/{teamId}/members/{userId}**	Updates a member's role on the team (e.g., promote a player to manager).	AuthZ("COACH", "ADMIN") + Ownership Check.
4. DELETE	**/api/teams/{teamId}/members/{userId}**	Removes a member from a team.	AuthZ("COACH", "ADMIN") + Ownership Check.

### internal service to service(gRPC)
1. gRPC services that team service provides to other services
2. gRPC services  that team service consumes from other services