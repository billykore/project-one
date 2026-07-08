# API Contract: Update User Profile

**Endpoint**: `PUT /users/profile`  
**Feature**: 001-user-profile-edit  
**Auth**: Required (JWT via `access_token` cookie or `Authorization: Bearer` header)

## Request

### Headers

```
Content-Type: application/json
Cookie: access_token=<JWT>
```

### Body

```json
{
  "first_name": "John",
  "last_name": "Doe",
  "username": "johndoe"
}
```

| Field | Type | Required | Constraints |
|-------|------|----------|-------------|
| `first_name` | string | yes | Min 3 chars, trimmed |
| `last_name` | string | yes | Min 3 chars, trimmed |
| `username` | string | yes | 3-30 chars, `[a-zA-Z0-9_]+`, case-insensitive unique, trimmed + lowercased |

## Responses

### 200 OK — Profile updated successfully

```json
{
  "message": "Profile updated successfully",
  "username": "johndoe"
}
```

### 400 Bad Request — Validation error

```json
{
  "error": "Key: 'UpdateProfileRequest.FirstName' Error:Field validation for 'FirstName' failed on the 'min' tag"
}
```

```json
{
  "error": "Username is already taken"
}
```

### 401 Unauthorized — Missing or invalid token

```json
{
  "error": "Unauthorized"
}
```

### 500 Internal Server Error

```json
{
  "error": "Internal server error"
}
```

## Behavior Notes

1. **Ownership**: The handler extracts `username` from the JWT context (set by `middleware.Authorize`). Only the authenticated user's own profile can be updated. There is no URL parameter for the target user.

2. **Partial updates**: All three fields (`first_name`, `last_name`, `username`) are required in the request. The endpoint does not support PATCH-style partial updates. To "not change" a field, send its current value.

3. **Username cascade**: If the username changes, the update MUST cascade to all denormalized username columns in `follows`, `posts`, `user_tokens`, `comments`, `notifications`, and `post_likes` tables. This happens within a database transaction.

4. **Whitespace handling**: Leading/trailing whitespace is trimmed from all fields. Username is lowercased before storage and uniqueness check.

5. **Idempotency**: Submitting the same values multiple times produces the same result (200 OK each time). The `updated_at` timestamp is updated on every successful call, even if field values are unchanged.
