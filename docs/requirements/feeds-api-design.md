# Development Requirement

Act as Senior Backend Engineer to make a plan from this development requirements.

## Prerequiset

- Create new working branch with name  `<category:feature|fix|refactor|test|chore|docs>/<short-description`.
- Read @AGENTS.md.
- Read @README.md.

## Planning

- Think step by step and outline a clear plan.
- List the main steps you would take.
- Call out important decisions or tradeoffs.
- Mention edge cases we should keep in mind.

## Development Objective

- Create new API for gets user feeds.
- API specification:
  - Method: GET
  - Path: /feeds
  - Request body: (none)
  - Query params (for pagination):
    - page: integer
    - limit: integer
    - example: `?page=1&limit=10`
  - Response body:

    ```json
    [
        {
            "author": "string",
            "comments": [
                {
                "content": "string",
                "created_at": "string",
                "id": 0,
                "username": "string"
                }
            ],
            "content": "string",
            "created_at": "string",
            "id": 0,
            "like_count": 0,
            "message": "string",
            "tags": [
                "string"
            ],
            "title": "string",
            "updated_at": "string"
        },
        {
            "author": "string",
            "comments": [
                {
                "content": "string",
                "created_at": "string",
                "id": 1,
                "username": "string"
                }
            ],
            "content": "string",
            "created_at": "string",
            "id": 1,
            "like_count": 0,
            "message": "string",
            "tags": [
                "string"
            ],
            "title": "string",
            "updated_at": "string"
        },
        {
            "author": "string",
            "comments": [
                {
                "content": "string",
                "created_at": "string",
                "id": 2,
                "username": "string"
                }
            ],
            "content": "string",
            "created_at": "string",
            "id": 2,
            "like_count": 0,
            "message": "string",
            "tags": [
                "string"
            ],
            "title": "string",
            "updated_at": "string"
        }
    ]
    ```

- The posts in feed is from the users that followed by logged in user and the logged in user posts.
- Use cursor based pagination for database optimization.
