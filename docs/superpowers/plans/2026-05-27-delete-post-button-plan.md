# Post Delete Button Authorization Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Correctly hide the Delete Post button on the frontend if the logged-in user is not the author of the post, and ensure the backend returns the post's author username.

**Architecture:** Update backend post handlers to map `Username` to `Author` in the DTO response. Update the frontend `Post` model, detail page, and `DeletePostButton` to fetch and check this author username against `localStorage` before displaying the button.

**Tech Stack:** Go (Echo), Next.js 16 (React 19, TypeScript)

---

### Task 1: Update Go Backend Handlers

**Files:**
- Modify: `internal/api/handler/post_handler.go`

- [ ] **Step 1: Update `GetPostByID` and `GetPosts` handlers**
  Add mapping for the `Author` field in `PostResponse`.
  - In `GetPostByID` (around line 146):
    ```go
    	return c.JSON(http.StatusOK, dto.PostResponse{
    		ID:        post.ID,
    		Title:     post.Title,
    		Content:   post.Content,
    		Tags:      post.Tags,
    		Author:    post.Username,
    		LikeCount: post.LikeCount,
    		Comments:  commentsResp,
    		CreatedAt: post.CreatedAt,
    		UpdatedAt: post.UpdatedAt,
    	})
    ```
  - In `GetPosts` (around line 189):
    ```go
    	response := make([]dto.PostResponse, 0, len(posts))
    	for _, p := range posts {
    		response = append(response, dto.PostResponse{
    			ID:        p.ID,
    			Title:     p.Title,
    			Content:   p.Content,
    			Tags:      p.Tags,
    			Author:    p.Username,
    			CreatedAt: p.CreatedAt,
    			UpdatedAt: p.UpdatedAt,
    		})
    	}
    ```

- [ ] **Step 2: Run Go backend tests**
  Run: `make test`
  Expected: PASS

- [ ] **Step 3: Commit backend changes**
  ```bash
  git add internal/api/handler/post_handler.go
  git commit -m "feat(api): populate Author field in post responses"
  ```

---

### Task 2: Update Frontend Models and Page Component

**Files:**
- Modify: `web/app/posts/model.ts`
- Modify: `web/app/posts/[id]/page.tsx`

- [ ] **Step 1: Update `Post` interface in model.ts**
  - Add `author?: string;` to `Post` model in `web/app/posts/model.ts`.
  ```typescript
  export interface Post {
    id: number;
    title: string;
    content: string;
    tags?: string[];
    author?: string;
    created_at: string;
    updated_at: string;
    message?: string;
    comments?: Comment[];
  }
  ```

- [ ] **Step 2: Pass `post.author` to `DeletePostButton`**
  - Modify `web/app/posts/[id]/page.tsx` to pass `postAuthor={post.author}`:
  ```typescript
  <DeletePostButton postId={post.id} postAuthor={post.author} />
  ```

- [ ] **Step 3: Commit changes**
  ```bash
  git add web/app/posts/model.ts web/app/posts/[id]/page.tsx
  git commit -m "feat(web): pass post author to DeletePostButton"
  ```

---

### Task 3: Update `DeletePostButton` Component

**Files:**
- Modify: `web/components/posts/DeletePostButton.tsx`

- [ ] **Step 1: Modify DeletePostButton component structure**
  - Update `DeletePostButtonProps` interface to include optional `postAuthor?: string`.
  - Use a `useEffect` hook to read the `localStorage.getItem("username")` client-side to prevent SSR mismatch errors.
  - Return `null` if `postAuthor` is provided and the current user does not match it.
  ```typescript
  interface DeletePostButtonProps {
    postId: number;
    postAuthor?: string;
    redirectPath?: string;
    onSuccess?: () => void;
  }

  export default function DeletePostButton({
    postId,
    postAuthor,
    redirectPath = "/posts",
    onSuccess,
  }: DeletePostButtonProps) {
    const router = useRouter();
    const [isOpen, setIsOpen] = useState(false);
    const [isDeleting, setIsDeleting] = useState(false);
    const [error, setError] = useState<string | null>(null);
    const [currentUser, setCurrentUser] = useState<string | null>(null);

    useEffect(() => {
      setCurrentUser(localStorage.getItem("username"));
    }, []);
  ```
  And in the render return check:
  ```typescript
    if (postAuthor && currentUser !== postAuthor) {
      return null;
    }
  ```

- [ ] **Step 2: Run linter on web**
  Run: `npm run lint` inside `web` directory
  Expected: PASS with no linting errors.

- [ ] **Step 3: Commit changes**
  ```bash
  git add web/components/posts/DeletePostButton.tsx
  git commit -m "feat(web): only show delete post button to post author"
  ```
