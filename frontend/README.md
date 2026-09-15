# EVA Bharat — Ticket System Frontend

A small, static frontend for the existing EVA Bharat ticket-system backend. Plain HTML/CSS/JavaScript — no framework, no build step, no bundler. It consumes the backend exactly as documented and does not modify it.

## What it does

- Register and log in against the existing backend
- View your tickets on a dashboard (table on desktop, cards on mobile)
- Create a new ticket
- Open a ticket to see its full details
- Move a ticket forward through its status (`open → in_progress → closed`), with the UI itself preventing invalid transitions and a closed ticket clearly marked as unable to reopen
- Log out (clears the stored token)

## Tech stack

- Vanilla HTML, CSS, JavaScript (ES2017+, no transpilation needed)
- Hash-based client-side router (`#/login`, `#/register`, `#/dashboard`, `#/tickets/new`, `#/tickets/:id`)
- No external JS libraries — `fetch` for HTTP, `localStorage` for the JWT
- Google Fonts (IBM Plex Sans / IBM Plex Mono) loaded via `<link>`

Because there's no build step, "building" this app just means serving the three static files (`index.html`, `style.css`, `api.js`, `app.js`) — there's nothing to compile and nothing that can fail to compile.

## Backend API URL

```
https://eva-bharat-p6w2.onrender.com
```

Set once, as `API_BASE_URL` in `api.js`. Nothing else hardcodes the URL.

## How authentication works

1. `POST /auth/login` is called with `{ email, password }`.
2. The JWT from the response is stored in `localStorage`.
3. Every request to a protected endpoint (`/tickets`, `/tickets/:id`, `/tickets/:id/status`) sends `Authorization: Bearer <token>`.
4. If any request comes back `401`, the stored token is cleared and the user is sent back to `/login`.
5. Logging out clears the token directly.

## Run locally

No install step is required. From this folder:

```bash
python3 -m http.server 8080
# or: npx serve .
```

Then open `http://localhost:8080`.

## Deploy

This is a static site, so it deploys as-is to Netlify or Vercel with no build command and no output-directory configuration (or `.` as the output directory if one is required):

- **Netlify**: drag the folder into Netlify's "Deploys" page, or `netlify deploy --prod` from this folder.
- **Vercel**: `vercel --prod` from this folder (framework preset: "Other").

## Known assumptions — please verify against your backend

I was not able to reach `eva-bharat-p6w2.onrender.com` or the Go source from my environment (network access here is restricted to a small package-registry allowlist and does not include Render or a browser). So this frontend was built strictly to the endpoint list and status rules given, with a few defensive assumptions about response shape that you should confirm:

- **Login response**: code looks for `token`, then `access_token`, then `jwt` in the JSON body. If your backend uses a different key, update `renderLogin` in `app.js` (one line).
- **Ticket fields**: code reads `id`/`title`/`description`/`status`/`created_at`, falling back to a couple of common casing variants (`ID`, `createdAt`, etc.). If your JSON uses something else, update `normalizeTicket` in `app.js`.
- **Ticket list response**: code accepts either a bare array or `{ tickets: [...] }` / `{ data: [...] }`.
- **CORS**: I could not inspect the Go backend to confirm whether it already sends `Access-Control-Allow-Origin` for browser requests. If it doesn't, browser requests from wherever you host this frontend will fail with a CORS error in the console (a fetch/network-style failure, not a clean 4xx) even though the API itself works fine from curl or Postman. If you see that, the backend needs a CORS middleware allowing your frontend's origin — that's a small, additive change to the Go server (e.g. one middleware registration), not a rewrite, and it's outside what I could make from here without access to the source.

None of this required guessing at functionality that doesn't exist in your spec — it's only about JSON key names, which weren't specified.

## Testing status (honest account)

I do not have network access to `eva-bharat-p6w2.onrender.com`, a browser, or a deployment CLI from where this was built, so none of the following were actually exercised against your live backend:

- Register / Login / Create ticket / List tickets / Ticket details / Status update / Closed-ticket protection / Logout / Unauthorized-request handling / Page-refresh persistence — **not tested from here**
- Deployment to Vercel/Netlify — **not performed from here**

What was verified locally: the three files parse as valid JavaScript with no syntax errors, and the app serves and loads correctly over a local static server.

**To get real PASS/FAIL results**, run it locally (see above) against the live backend, or deploy it and click through the flow yourself — register, log in, create a ticket, move it through both status transitions, confirm a closed ticket shows no further action, log out, and refresh mid-session to confirm you land back on `/login` only when the token is actually gone. If anything doesn't line up with the backend's real response shape, the two spots noted above under "Known assumptions" are where to look first.
