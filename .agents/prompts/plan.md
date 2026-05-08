You are a senior engineer helping me with a new development. Create plan for following development requirements. Write it into a file called ‘ai_plan.md’.

Pre-requisite:

- Read @AGENTS.md to apply skills, learn the coding style, rules, and constraints for this codebase.
- Work /internal directory.

Development focus:

- Create API for get user's a posts.

Plan:

- Think step by step and outline a clear plan.
- List the main steps you would take.
- Call out important decisions or tradeoffs.
- Mention edge cases we should keep in mind.


API Specification:

- Method: GET 
- Endpoint: /posts
- 200 Success response:
```json
[
    {
        "id": 1,
        "title": "My first post",
        "content": "Lorem ipsum dolor sit amet consectetur adipiscing elit. Sit amet consectetur adipiscing elit quisque faucibus ex. Adipiscing elit quisque faucibus ex sapien vitae pellentesque.",
        "tags": [
            "first-post",
            "lorem-ipsum"
        ],
        "created_at": "2026-05-06T21:04:34.377004+07:00",
        "updated_at": "2026-05-06T21:04:34.377004+07:00"
    },
    {
        "id": 2,
        "title": "My second post",
        "content": "Lorem ipsum dolor sit amet consectetur adipiscing elit. Sit amet consectetur adipiscing elit quisque faucibus ex. Adipiscing elit quisque faucibus ex sapien vitae pellentesque.",
        "tags": [
            "second-post",
            "lorem-ipsum"
        ],
        "created_at": "2026-05-06T21:04:34.377004+07:00",
        "updated_at": "2026-05-06T21:04:34.377004+07:00"
    },
    ...(more posts)
]
```
- Success but empty:
```json
[]
```
- 401 Unauthorized response:
```json
{
    "error": "Unauthorized"
}
- 500 Internal Server Error response:
```json
{
    "error": "Something went wrong"
}
```

Dependencies:

- Echo for routing.
- JWT for authentication.
- GORM for ORM.
- PostgreSQL for database.
- Validator for input validation.

Rules:

- The posts must be exists in the database.
- The posts must belongs to logged in user.

The plan may have:

- Explain the steps clearly so that a junior programmer or an AI model can easily understand.
- Provides code snippets (if necessary).

Thank you.
