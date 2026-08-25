import { notFound } from "next/navigation";
import { serverFetch } from "@/lib/server-fetch";
import { ApiError, handleApiResponse } from "@/lib/errors";
import { Post } from "@/lib/types/post.types";
import PostInteractionSection from "@/components/posts/post-interaction-section";
import { cookies } from "next/headers";
import Navbar from "@/components/layout/navbar";
import SiteFooter from "@/components/layout/site-footer";

async function getPost(id: string) {
  try {
    const res = await serverFetch(`/api/posts/${id}`);
    const post = await handleApiResponse<Post>(res);
    if (!post) {
      notFound();
    }
    return post;
  } catch (err) {
    if (err instanceof ApiError) {
      if (err.status === 404) {
        notFound();
      }
    }
    throw err;
  }
}

interface PageProps {
  params: Promise<{ id: string }>;
}

export default async function PostDetailPage({ params }: PageProps) {
  const { id } = await params;
  const post = await getPost(id);
  const cookieStore = await cookies();
  const isAuthenticated = cookieStore.has("access_token");

  const published = new Date(post.created_at);
  const wasUpdated = post.updated_at !== post.created_at;

  return (
    <div className="flex min-h-screen flex-col bg-paper">
      <Navbar pageTitle="Post" />

      <main className="mx-auto w-full max-w-4xl flex-1 px-4 py-8 sm:px-6 sm:py-12">
        <article>
          <header className="border-b border-rule pb-8">
            <h1 className="text-3xl font-semibold text-ink sm:text-[2.75rem] sm:leading-[1.1]">
              {post.title}
            </h1>
          </header>

          <div className="grid gap-8 pt-8 lg:grid-cols-[9rem_minmax(0,1fr)] lg:gap-12">
            {/* The margin: attribution, dates, filing marks. */}
            <aside className="flex flex-wrap items-baseline gap-x-5 gap-y-3 lg:sticky lg:top-24 lg:block lg:self-start lg:space-y-4">
              {post.author && (
                <div>
                  <p className="meta">Written by</p>
                  <a href={`/${post.author}`} className="meta mt-0.5 block font-medium text-ink hover:text-accent">
                    {post.author}
                  </a>
                </div>
              )}
              <div>
                <p className="meta">Published</p>
                <time dateTime={post.created_at} className="meta mt-0.5 block text-ink">
                  {published.toLocaleDateString(undefined, {
                    day: "numeric",
                    month: "long",
                    year: "numeric",
                  })}
                </time>
              </div>
              {wasUpdated && (
                <div>
                  <p className="meta">Revised</p>
                  <time dateTime={post.updated_at} className="meta mt-0.5 block text-ink">
                    {new Date(post.updated_at).toLocaleDateString(undefined, {
                      day: "numeric",
                      month: "long",
                      year: "numeric",
                    })}
                  </time>
                </div>
              )}
              {post.tags && post.tags.length > 0 && (
                <div>
                  <p className="meta">Filed under</p>
                  <div className="mt-1 flex flex-wrap gap-x-2.5 gap-y-1">
                    {post.tags.map((tag) => (
                      <span key={tag} className="tag">{tag}</span>
                    ))}
                  </div>
                </div>
              )}
            </aside>

            <div className="min-w-0">
              <div className="prose-body max-w-[38rem] whitespace-pre-wrap">{post.content}</div>

              <PostInteractionSection
                postId={post.id}
                postAuthor={post.author}
                initialComments={post.comments}
                isGuest={!isAuthenticated}
                initialLikeCount={post.like_count}
              />
            </div>
          </div>
        </article>
      </main>

      <SiteFooter />
    </div>
  );
}
