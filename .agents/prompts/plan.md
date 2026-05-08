You are a senior engineer helping me with a new development. Create plan for following development requirements. Write it into a file called ‘ai_plan.md’.

Pre-requisite:

- Work /internal directory.
- Read @README.md to understand the project scope, architecture, and constraints.
- Read @AGENTS.md to learn the coding style, rules, and constraints for this codebase.
- Use available skills for Go programming language.

Development focus:

- Create API for updating a post.

Plan:

- Think step by step and outline a clear plan.
- List the main steps you would take.
- Call out important decisions or tradeoffs.
- Mention edge cases we should keep in mind.


API Specification:

- Method: PUT 
- Endpoint: /posts/:id
- Request body:
```json
{
    "title": "Updated title",
    "content": "Updated content"
}
```
- 200 Success response:
```json
{
    "id": 1,
    "message": "Post updated successfully",
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
- Title and content must not be empty.
- If title or content are empty, Use current title or content when updated into the database.

Explain the steps clearly so that a junior programmer or an AI model can easily understand.