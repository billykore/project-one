# Design Spec: Post Delete Button Enhancements

## 1. Problem Statement
In the current implementation of the post details page at `web/app/posts/[id]/page.tsx`, a `DeletePostButton` is rendered for every user.
However, only the author of the post is authorized to delete the post according to the backend Go API. If a non-author attempts to delete the post, the API call will fail, but the button is still fully visible and clickable, leading to a poor user experience.
Additionally, the backend `PostResponse` DTO includes an `Author` field, but it is currently not populated by the `GetPostByID` or `GetPosts` handlers, meaning the frontend lacks information about the post's author.

## 2. Proposed Changes

### Backend (Go)
1. **Update Handlers**: Update `GetPostByID` and `GetPosts` in `internal/api/handler/post_handler.go` to set the `Author` field of the `dto.PostResponse` to the post's author username (`post.Username`).

### Frontend (Next.js)
1. **Update Model**: Update `web/app/posts/model.ts` to include the `author` field in the `Post` interface.
2. **Update Page Component**: Update `web/app/posts/[id]/page.tsx` to read `post.author`.
3. **Update Button Component**:
   - Update `DeletePostButton` to accept `postAuthor?: string` as a prop.
   - On the client-side, retrieve the current logged-in username from `localStorage.getItem("username")`.
   - Only render the button if the current user matches `postAuthor`.
   - If not matching (or if the current user is not logged in), render `null` so the delete action is completely hidden.

## 3. Alternative Approaches Considered
- **Approach B**: Keep the button visible but disable it.
  - *Trade-off*: Clutters the UI on posts the user doesn't own.
- **Approach C**: Rely only on backend API error responses.
  - *Trade-off*: Bad user experience since the user can trigger an action that is guaranteed to fail.

## 4. Verification Plan
1. **Backend Tests**: Run `make test` to ensure all tests pass.
2. **Linter**: Run `npm run lint` on the Next.js app to ensure no TypeScript or styling issues.
3. **Manual Verification**: Run the development server and test with two different logged-in users to verify that the Delete button is only visible to the author of the post.
