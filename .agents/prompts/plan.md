You are a senior engineer helping me with a new development. Create plan for following development requirements. Write it into a file called ‘ai_plan.md’.

Pre-requisite:

- Work /internal directory.
- Read @AGENTS.md to learn the coding style, rules, and constraints for this codebase.

Development focus:

- Create API for updating a post.

Plan:

- Think step by step and outline a clear plan.
- List the main steps you would take.
- Call out important decisions or tradeoffs.
- Mention edge cases we should keep in mind.


API Specification:

- Method: DELETE 
- Endpoint: /posts/:id
- 200 Success response:
```json
{
    "id": 1,
    "message": "Post deleted successfully",
}
```
- 401 Unauthorized response:
```json
{
    "error": "Unauthorized"
}
```
- 404 Bad Request response:
```json
{
    "error": "Post not found"
}
```
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

- Post ID must be integer and not 0.
- The post must be exists in the database.
- The post must belongs to logged in user.
- Use soft delete using GORM.

Expected plan output:

- Explain the steps clearly so that a junior programmer or an AI model can easily understand.
- Provides code snippets if necessary.

Thank you.
