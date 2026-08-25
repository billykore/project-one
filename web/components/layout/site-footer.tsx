export default function SiteFooter() {
  return (
    <footer className="mt-16 border-t border-rule">
      <div className="mx-auto flex w-full max-w-6xl flex-wrap items-baseline justify-between gap-2 px-4 py-6 sm:px-6">
        <p className="meta">Project One &mdash; a place to write things down</p>
        <p className="meta text-faint">&copy; {new Date().getFullYear()}</p>
      </div>
    </footer>
  );
}
