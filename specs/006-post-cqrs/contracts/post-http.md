# HTTP Contract: Post CQRS Refactor

**Date**: 2026-09-17 | **Feature**: `006-post-cqrs`

The refactor is internal. Every route, HTTP method, authentication rule, request/response DTO, status code, and RFC 9457 error remains unchanged. The table records which internal capability owns each established route after the split.

| Method | Path | Capability | Authentication | Successful response |
|--------|------|------------|----------------|---------------------|
| POST | `/posts` | Post command: create | Required; existing creation feature gate | `201` create-post response |
| PUT | `/posts/{id}` | Post command: update | Required | `200` post response |
| DELETE | `/posts/{id}` | Post command: delete | Required | `200` post response |
| POST | `/posts/{id}/likes` | Post command: like | Required | `200` like response |
| DELETE | `/posts/{id}/likes` | Post command: unlike | Required | `200` like response |
| GET | `/posts/{id}` | Post query: get by ID | Public | `200` post response including comments |
| GET | `/posts` | Post query: current user's list | Required | `200` cursor-paginated post list |
| GET | `/users/{username}/posts` | Post query: author's list | Public | `200` cursor-paginated post list |
| GET | `/posts/{id}/likes` | Post query: like status | Required | `200` like response |

## Unchanged behaviors

- List `limit` remains 1–100 with a default of 10; the returned cursor remains opaque.
- The detail response continues to include comments from the existing comment query.
- The existing feed remains read-only and continues to use the same cursor behavior through the query repository.
- Like and unlike responses continue to expose the current `liked` state and count, including idempotent repeats.
- Invalid IDs, invalid cursors, missing authentication, not-found results, ownership failures, validation failures, and internal failures retain their current problem response categories.
- `POST /posts/{id}/comments` is unchanged and remains a command route delegated to the existing comment use case; it is not one of the post use-case operations being split.

No Swagger contract revision is expected. Run documentation generation to prove annotations and generated artifacts remain synchronized.
