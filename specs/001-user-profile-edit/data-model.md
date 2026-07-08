# Data Model: User Profile Edit

**Feature**: 001-user-profile-edit  
**Date**: 2026-07-08

## Entities

### User (existing — no schema changes)

The `users` table already contains all necessary columns. No migration is required.

| Column | Type | Constraints | Notes |
|--------|------|-------------|-------|
| `id` | SERIAL | PRIMARY KEY | Auto-incrementing |
| `email` | VARCHAR(255) | UNIQUE, NOT NULL | Not editable in this feature |
| `username` | VARCHAR(255) | UNIQUE, NOT NULL | Editable; case-insensitive uniqueness enforced |
| `password` | VARCHAR(255) | NOT NULL | Not editable in this feature |
| `first_name` | VARCHAR(255) | NOT NULL | Editable; trimmed, min 3 chars |
| `last_name` | VARCHAR(255) | NOT NULL | Editable; trimmed, min 3 chars |
| `created_at` | TIMESTAMP WITH TIME ZONE | DEFAULT CURRENT_TIMESTAMP | Auto-managed |
| `updated_at` | TIMESTAMP WITH TIME ZONE | DEFAULT CURRENT_TIMESTAMP | Updated on profile edit |

### Editable Fields

| Field | Domain Validation | Notes |
|-------|------------------|-------|
| `first_name` | Required, 3-100 chars, trimmed | Same rules as registration |
| `last_name` | Required, 3-100 chars, trimmed | Same rules as registration |
| `username` | Required, 3-30 chars, alphanumeric + underscores only, case-insensitive unique, trimmed + lowercased | Same rules as registration; must not conflict with another user's username |

### Non-Editable Fields (in this feature)

- `email` — not part of this feature's scope
- `password` — handled by existing `PUT /users/password` endpoint
- `id`, `created_at`, `updated_at` — system-managed

## State Transitions

The profile edit does not involve workflow states. The user profile is always in a single "active" state. The edit operation is a simple UPDATE:

```
[Current Profile] → (user submits edit form) → [Validation] → [Updated Profile]
                                                    ↓ (invalid)
                                              [Profile preserved, errors shown]
```

## Validation Rules

### Domain-level (`User.ValidateProfileUpdate()`)

```
first_name  := required, trimmed, 3-100 characters
last_name   := required, trimmed, 3-100 characters
username    := required, trimmed, lowercased, 3-30 characters,
               matches ^[a-zA-Z0-9_]+$, case-insensitive unique
```

### DTO-level (validator/v10 struct tags)

```go
type UpdateProfileRequest struct {
    FirstName string `json:"first_name" validate:"required,min=3,max=100"`
    LastName  string `json:"last_name" validate:"required,min=3,max=100"`
    Username  string `json:"username" validate:"required,min=3,max=30"`
}
```

### Client-side (Zod schema)

```ts
const editProfileSchema = z.object({
  first_name: z.string().trim().min(3, "First name must be at least 3 characters").max(100, "First name must be at most 100 characters"),
  last_name: z.string().trim().min(3, "Last name must be at least 3 characters").max(100, "Last name must be at most 100 characters"),
  username: z.string().trim().min(3, "Username must be at least 3 characters")
    .max(30, "Username must be at most 30 characters")
    .regex(/^[a-zA-Z0-9_]+$/, "Username may only contain letters, numbers, and underscores"),
});
```

## Relationships

- **User → Follows**: If username changes, existing `follows` records (`follower_username`, `followed_username`) must be updated. The feature MUST cascade the username change to the `follows` table to maintain referential integrity.
- **User → Posts**: Posts use denormalized `username` column. The feature MUST cascade the username change to the `posts` table.
- **User → UserTokens**: Uses denormalized `username` column. The feature MUST cascade the username change to the `user_tokens` table.
- **User → Comments**: Uses denormalized `username`. The feature MUST cascade the username change.
- **User → Notifications**: May reference `actor_username`. The feature MUST cascade the username change.
- **User → PostLikes**: Uses denormalized `username`. The feature MUST cascade the username change.

> **Design Note**: The cascading username update across denormalized columns is the primary implementation complexity of this feature. The `UpdateUser` repository method uses `gorm.Save()` which only updates the `users` table. A dedicated `UpdateProfile` method must be added that updates the `users` row AND propagates username changes to all dependent tables within a transaction.
