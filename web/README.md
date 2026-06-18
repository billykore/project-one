This is a [Next.js](https://nextjs.org) project bootstrapped with [`create-next-app`](https://nextjs.org/docs/app/api-reference/cli/create-next-app).

## Getting Started

First, run the development server:

```bash
npm run dev
# or
yarn dev
# or
pnpm dev
# or
bun dev
```

Open [http://localhost:3000](http://localhost:3000) with your browser to see the result.

You can start editing the page by modifying `app/page.tsx`. The page auto-updates as you edit the file.

This project uses [`next/font`](https://nextjs.org/docs/app/building-your-application/optimizing/fonts) to automatically optimize and load [Geist](https://vercel.com/font), a new font family for Vercel.

## Real-time Notifications (WebSocket)

The app connects to the backend WebSocket stream to display live notifications without a page refresh.

### Endpoint

| Setting | Default | Description |
|---------|---------|-------------|
| `NEXT_PUBLIC_WS_URL` | `ws://localhost:8080/websocket` | Full WebSocket URL of the backend stream |
| `NEXT_PUBLIC_WS_DEBUG` | _(unset)_ | Set to `"true"` to enable verbose WS debug logging |

Set these in `.env.local` when running locally:

```
NEXT_PUBLIC_WS_URL=ws://localhost:8080/websocket
NEXT_PUBLIC_WS_DEBUG=true
```

### Authentication note

> **⚠ Backend change required for authenticated streams**
>
> The backend `/websocket` handler currently validates the JWT from the
> `Authorization: ****** HTTP header.  Native browser `WebSocket`
> objects cannot set arbitrary headers during the upgrade handshake, so
> authenticated streams will not work until the backend is updated.
>
> **Required backend fix**: update `internal/api/handler/websocket_handler.go`
> to also accept the `access_token` cookie (mirroring the behaviour of
> `internal/api/middleware/authorization.go`).  Browsers automatically forward
> cookies matching the WebSocket host, so no frontend change is required once
> the backend is fixed.
>
> Until that change lands, the panel connection indicator will show **Offline**.

### Reconnection behaviour

The client uses capped exponential back-off (1 s → 30 s, up to 10 attempts).
After a successful reconnect the app fires one extra REST fetch to recover any
notifications that arrived during the downtime.

### Troubleshooting

| Symptom | Likely cause | Fix |
|---------|--------------|-----|
| Panel always shows **Offline** | Backend rejects WS with 401 | See _Authentication note_ above |
| Panel shows **Reconnecting** forever | Backend not running, or wrong URL | Check `NEXT_PUBLIC_WS_URL` and that `localhost:8080` is reachable |
| Duplicate notifications on reconnect | REST refresh races with live stream | Dedup by `id` is handled automatically; no action needed |
| No debug output | Debug logging disabled | Set `NEXT_PUBLIC_WS_DEBUG=true` and check browser console |

## Learn More

To learn more about Next.js, take a look the following resources:

- [Next.js Documentation](https://nextjs.org/docs) - learn about Next.js features and API.
- [Learn Next.js](https://nextjs.org/learn) - an interactive Next.js tutorial.

You can check out [the Next.js GitHub repository](https://github.com/vercel/next.js) - your feedback and contributions are welcome!

## Deploy on Vercel

The easiest way to deploy your Next.js app is to use the [Vercel Platform](https://vercel.com/new?utm_medium=default-template&filter=next.js&utm_source=create-next-app&utm_campaign=create-next-app-readme) from the creators of Next.js.

Check out our [Next.js deployment documentation](https://nextjs.org/docs/app/building-your-application/deploying) for more details.
