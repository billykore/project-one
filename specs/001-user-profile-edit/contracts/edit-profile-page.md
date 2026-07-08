# Frontend Contract: Edit Profile Page

**Route**: `/settings/profile/edit`  
**Feature**: 001-user-profile-edit  
**Auth**: Required (enforced by `proxy.ts` middleware + server-side cookie check)

## Page Behavior

### Loading State (Server Component)

The Server Component at `app/settings/profile/edit/page.tsx`:
1. Reads the `username` cookie to identify the current user
2. Calls `serverFetch(\`/api/users/${username}\`)` to fetch current profile data
3. Displays a `PageSkeletonLayout` while loading
4. On 401 → redirects to `/login`
5. On other errors → shows error state with retry

### Normal State

After data loads, renders `<EditProfileForm>` client component with pre-populated data:

```tsx
<EditProfileForm
  initialFirstName="John"
  initialLastName="Doe"
  initialUsername="johndoe"
/>
```

### Form Interaction

The `EditProfileForm` client component:
- Displays three `InputField` components (First Name, Last Name, Username)
- Shows current values as pre-filled defaults
- Validates on submit via Zod schema
- Shows field-level error messages on validation failure
- Submits `PUT /api/users/profile` with JSON body
- Shows a submit button with loading spinner during request
- On network error: shows error banner with retry
- On 400 (validation/duplicate): shows inline error message from API
- On 200 success: calls `router.push(\`/${username}?updated=1\`)`

### Success Redirect

The profile page at `app/[username]/page.tsx` detects `?updated=1` search param and displays a success banner:

```
✅ Profile updated successfully
```

### Error States

| Condition | UX |
|-----------|-----|
| Empty first_name | Field-level: "First name is required" |
| Empty last_name | Field-level: "Last name is required" |
| Short name (<3 chars) | Field-level: "Must be at least 3 characters" |
| Short/long username | Field-level: "Username must be 3-30 characters" |
| Invalid username chars | Field-level: "Username may only contain letters, numbers, and underscores" |
| Username taken | Banner below form: "This username is already taken. Please choose another." |
| Auth expired | Redirect to /login (with return URL) |
| Network error | Banner: "Unable to connect. Please try again." with Retry button |

## Component Props Contract

```ts
interface EditProfileFormProps {
  initialFirstName: string;
  initialLastName: string;
  initialUsername: string;
}
```

## Navigation Flow

```
Profile Page ([username])
  └── "Edit Profile" button (visible only if isOwner === true)
        └── Link → /settings/profile/edit
              └── EditProfileForm (pre-populated)
                    ├── Success → router.push(`/[username]?updated=1`)
                    └── Cancel → router.back() or Link to /[username]
```

## Design Constraints

- Follows glassmorphism patterns: `rounded-2xl`, `bg-white/80 backdrop-blur`, border-zinc-200/dark:border-zinc-800
- Responsive: single column on mobile, centered max-width card on desktop
- Uses existing `InputField` from `web/components/ui/input.tsx`
- Dark mode support via Tailwind `dark:` variants
- Accessibility: labels, `aria-invalid`, `aria-describedby` on error fields
