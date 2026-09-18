# Data Model: Post CQRS Boundaries

**Date**: 2026-09-17 | **Feature**: `006-post-cqrs`

This is a boundary refactor. It introduces no persisted entities, fields, migrations, indexes, or state transitions.

## Existing persisted entities

| Entity | Relevant fields | Relationship and rules |
|--------|-----------------|------------------------|
| Post | ID, author identity, title, content, tags, like count, created/updated timestamps | Owned by one user; commands create/update/delete it; queries return it. |
| Like | Post ID, username, created timestamp | Identifies one user's engagement with one post; commands create/delete it idempotently; queries check its existence. |
| Notification | recipient, actor, type, post ID, timestamp | Produced only after a successful like of another user's post; query operations never create it. |

## Application boundaries

| Boundary | Operations | Dependencies | State rule |
|----------|------------|--------------|------------|
| Post command | Create, update, delete, like, unlike | Command post repository (including command precondition reads), like repository, user lookup, publisher, logger, feature evaluator | May change post/like state and publish the existing notification. |
| Post query | Get by ID, list by author, get like status | Query post repository, like repository, logger | Must only read existing state. |
| Feed query | Retrieve a followed-user feed | Query post repository, follow repository, user repository, logger | Must only read existing state. |
| Comment command | Create a comment after validating its post | Command post repository, comment repository, user lookup, publisher | May create comment state; post lookup is a command precondition. |

## Validation and transition rules retained

- Post identifiers must be positive; invalid identifiers preserve existing errors.
- Authors may update or delete only their own posts.
- Create requires the existing creation entitlement and valid title/content/tags input.
- Updates trim supplied text and preserve unspecified title or content.
- A duplicate like and a missing like on unlike preserve their existing idempotent count/result.
- Author lists retain the existing cursor, default/max limit, ordering, and continuation behavior.

## Data change impact

None. Two distinct repository adapters and ports use the existing database tables and post model; their SQL behavior and external response models remain unchanged.
