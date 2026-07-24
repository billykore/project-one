# Generate Test Case and Postman Collection

You are an experienced QA Engineer tasked with turning an API specification into:

1. A structured list of test cases
2. A Postman Collection (v2.1 format) containing the automated implementation of those test cases
3. Save the test cases and collection in the `/test/api` folder

## OUTPUT RULES (MUST BE FOLLOWED, DO NOT DEVIATE)

### Part 1: Test Cases

For EVERY endpoint, create test cases using the table format below. Do not add or remove columns.

| ID | Endpoint | Method | Scenario | Type | Precondition | Request Body/Params | Expected Status | Expected Response |
|----|----------|--------|----------|------|---------------|----------------------|------------------|---------------------|

ID convention: `TC-{SHORT_ENDPOINT_NAME}-{3_DIGIT_SEQUENCE}`
Example: TC-LOGIN-001, TC-LOGIN-002, TC-PAYMENT-001

"Type" convention (use ONLY one of the following labels):

- Positive (happy path)
- Negative (invalid input)
- Edge Case (boundary values, empty data, etc.)
- Security (auth/authorization, injection, etc.)
- Contract (response schema/data type validation)

For EVERY endpoint, there must be at least:

- 1 Positive test case
- 2 Negative test cases (missing required field, wrong data type)
- 1 Edge Case test case
- 1 Security test case (if the endpoint requires authentication/authorization)
- 1 Contract test case (validating the response structure against the spec)

### Part 2: Postman Collection

After the test case table, generate ONE Postman Collection v2.1 file (JSON format) that:

- Represents each test case above as a single Postman request
- Names each request EXACTLY after its test case ID, format: "{ID} - {Short scenario}"
- Groups requests into one folder per endpoint (Postman folder = group per endpoint)
- Includes a test script (`event` type "test") in every request using pm.test() to:
  - Validate the status code against the Expected Status
  - Validate the response structure/fields against the Expected Response (using pm.expect / pm.response.to.have)
  - Validate response time < 1000ms (default, unless stated otherwise in the spec)
- Uses collection variables for base_url and token/auth — DO NOT hardcode sensitive values
- For test cases requiring chaining (e.g., getting a token from login and using it in another request), use pm.collectionVariables.set() in the preceding request's test script, and reference it as {{variable_name}} in the following request
- Includes a pre-request script ONLY when needed (e.g., generating a timestamp/signature)

### Response Format

Respond ONLY with the following two sections, with no additional explanation outside of them:

1. Heading "## Test Case" followed by the complete test case table
2. Heading "## Postman Collection" followed by ONE valid JSON code block containing the entire collection

Do not abbreviate or write "etc." — every endpoint and every test case from Part 1 must have a full representation in the collection JSON in Part 2.

## API SPEC

@api/swagger/swagger.yaml
