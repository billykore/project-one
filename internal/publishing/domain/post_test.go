package domain

import "testing"

func TestPostMutations(t *testing.T) {
	post := &Post{Title: "old", Content: "old", LikeCount: 0}

	post.Update(" new title ", " new content ")
	post.RemoveLike()
	post.AddLike()

	if post.Title != "new title" || post.Content != "new content" || post.LikeCount != 1 {
		t.Fatalf("unexpected post state: %#v", post)
	}
}
