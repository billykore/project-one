"use client";

import { useState, useEffect, useRef, useCallback, type FormEvent, type KeyboardEvent } from "react";
import { useRouter } from "next/navigation";
import { handleApiResponse } from "@/lib/errors";
import { Avatar } from "@/components/ui/avatar";
import type { SearchResponse } from "@/lib/types/search.types";

const DEBOUNCE_MS = 300;
const MAX_SUGGESTIONS = 5;

export default function SearchBar() {
  const [query, setQuery] = useState("");
  const [suggestions, setSuggestions] = useState<{ username: string; name: string }[]>([]);
  const [showDropdown, setShowDropdown] = useState(false);
  const [loading, setLoading] = useState(false);
  const [activeIndex, setActiveIndex] = useState(-1);
  const router = useRouter();
  const inputRef = useRef<HTMLInputElement>(null);
  const containerRef = useRef<HTMLDivElement>(null);
  const abortRef = useRef<AbortController | null>(null);

  const handleChange = (value: string) => {
    setQuery(value);
    const trimmed = value.trim();
    if (trimmed.length < 3) {
      setSuggestions([]);
      setShowDropdown(false);
      setActiveIndex(-1);
    }
  };

  // Debounced fetch — only for queries ≥ 3 chars
  useEffect(() => {
    const trimmed = query.trim();
    if (trimmed.length < 3) return;

    // Cancel previous in-flight request
    if (abortRef.current) abortRef.current.abort();
    const controller = new AbortController();
    abortRef.current = controller;

    const timer = setTimeout(async () => {
      setLoading(true);
      try {
        const res = await fetch(`/api/users/search?q=${encodeURIComponent(trimmed)}&limit=${MAX_SUGGESTIONS}`, {
          signal: controller.signal,
        });
        const { data } = await handleApiResponse<SearchResponse>(res);
        setSuggestions(data);
        setShowDropdown(data.length > 0);
        setActiveIndex(-1);
      } catch {
        // ponytail: abort errors and 401 redirects are handled silently
      } finally {
        if (!controller.signal.aborted) setLoading(false);
      }
    }, DEBOUNCE_MS);

    return () => {
      clearTimeout(timer);
      controller.abort();
    };
  }, [query]);

  // Click outside → close dropdown
  useEffect(() => {
    const handleClick = (e: MouseEvent) => {
      if (containerRef.current && !containerRef.current.contains(e.target as Node)) {
        setShowDropdown(false);
      }
    };
    document.addEventListener("mousedown", handleClick);
    return () => document.removeEventListener("mousedown", handleClick);
  }, []);

  const navigateToProfile = useCallback(
    (username: string) => {
      setShowDropdown(false);
      setQuery("");
      router.push(`/${username}`);
    },
    [router],
  );

  const handleSubmit = (e: FormEvent) => {
    e.preventDefault();
    const trimmed = query.trim();
    if (trimmed.length === 0) return;

    // If a suggestion is highlighted, navigate to that profile
    if (activeIndex >= 0 && activeIndex < suggestions.length) {
      navigateToProfile(suggestions[activeIndex].username);
      return;
    }

    setShowDropdown(false);
    router.push(`/search?q=${encodeURIComponent(trimmed)}`);
  };

  const handleKeyDown = (e: KeyboardEvent<HTMLInputElement>) => {
    if (!showDropdown) return;
    switch (e.key) {
      case "ArrowDown":
        e.preventDefault();
        setActiveIndex((prev) => (prev < suggestions.length - 1 ? prev + 1 : 0));
        break;
      case "ArrowUp":
        e.preventDefault();
        setActiveIndex((prev) => (prev > 0 ? prev - 1 : suggestions.length - 1));
        break;
      case "Escape":
        setShowDropdown(false);
        setActiveIndex(-1);
        break;
    }
  };

  return (
    <form onSubmit={handleSubmit} className="flex w-full items-center">
      <div ref={containerRef} className="relative w-full">
        <input
          ref={inputRef}
          type="text"
          value={query}
          onChange={(e) => handleChange(e.target.value)}
          onKeyDown={handleKeyDown}
          onFocus={() => { if (suggestions.length > 0) setShowDropdown(true); }}
          placeholder="Find someone"
          aria-label="Search users"
          autoComplete="off"
          className="field w-full py-1.5 pr-3 pl-9"
        />
        <button
          type="submit"
          aria-label="Search"
          className="absolute top-1/2 left-2 -translate-y-1/2 text-faint transition-colors hover:text-accent"
        >
          <svg className="h-4 w-4" fill="none" viewBox="0 0 24 24" strokeWidth="2" stroke="currentColor">
            <path strokeLinecap="round" strokeLinejoin="round" d="m21 21-5.197-5.197m0 0A7.5 7.5 0 1 0 5.196 5.196a7.5 7.5 0 0 0 10.607 10.607Z" />
          </svg>
        </button>

        {/* Suggestion dropdown */}
        {showDropdown && (
          <div className="pop enter-pop absolute top-full left-0 z-50 mt-1.5 w-full overflow-hidden">
            {loading ? (
              <div className="flex items-center justify-center py-3">
                <div className="h-4 w-4 animate-spin rounded-full border-2 border-accent border-t-transparent" />
              </div>
            ) : suggestions.length === 0 ? (
              <p className="meta px-3 py-2.5">No one by that name</p>
            ) : (
              <ul role="listbox">
                {suggestions.map((user, i) => (
                  <li
                    key={user.username}
                    role="option"
                    aria-selected={i === activeIndex}
                    onClick={() => navigateToProfile(user.username)}
                    className={`flex cursor-pointer items-center gap-3 px-3 py-2 text-sm transition-colors ${
                      i === activeIndex ? "bg-accent-soft" : "hover:bg-sunken"
                    }`}
                  >
                    <Avatar name={user.name} size="sm" />
                    <div className="min-w-0">
                      <p className="truncate font-medium text-ink">{user.name}</p>
                      <p className="meta truncate">@{user.username}</p>
                    </div>
                  </li>
                ))}
              </ul>
            )}
          </div>
        )}
      </div>
    </form>
  );
}
