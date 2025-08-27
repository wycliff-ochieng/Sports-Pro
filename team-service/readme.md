### Team Service

1. GET/teams/me/ - > teams for the current profile
2. Get/teams/id/ -> team details
3. POST/team/ -> Create Team (coaches)
4. PUT/team/id -> Update team (coaches & managers)


**Endpoints for Teams**
### Method	Endpoint	Description	RBAC Middleware

1. POST	**/api/teams**	Creates a new team. The creator is automatically made the 'COACH'.	AuthZ("COACH", "ADMIN") -> Done
2. GET	**/api/teams/me**	Gets a list of all teams the current user is a member of.	None (Any role can see their teams) -> Done
3. GET	**/api/teams/{teamId}**	Gets the detailed public profile of a single team.	None (Any role can view a team) -> Done
4. PUT	**/api/teams/{teamId}**	Updates a team's details (name, description, etc.).	AuthZ("COACH", "ADMIN") -> Done
5. DELETE	**/api/teams/{teamId}**	Deletes a team. A highly destructive action.	AuthZ("ADMIN") -> In Progrss -> (Thinking oof implementin status( Active and Inactive) ) instead of deleting || or both ->Can delete or set status

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


The **make** built-in function allocates and initializes an object of type slice, map, or chan (only). Like new, the first argument is a type, not a value. Unlike new, make's return type is the same as the type of its argument, not a pointer to it. The specification of the result depends on the type:

Slice: The size specifies the length. The capacity of the slice is equal to its length. A second integer argument may be provided to specify a different capacity; it must be no smaller than the length. For example, make([]int, 0, 10) allocates an underlying array of size 10 and returns a slice of length 0 and capacity 10 that is backed by this underlying array.
Map: An empty map is allocated with enough space to hold the specified number of elements. The size may be omitted, in which case a small starting size is allocated.
Channel: The channel's buffer is initialized with the specified buffer capacity. If zero, or the size is omitted, the channel is unbuffered.
