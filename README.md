# PostHog Proxy

A lightweight reverse proxy for routing PostHog analytics through your own domain. This helps bypass ad blockers and increases event capture rates.

## Why Use a Proxy?

Ad blockers often block requests to `posthog.com`. By proxying through your own domain, analytics requests appear as first-party traffic and are less likely to be blocked.

## How It Works

The proxy routes requests to PostHog's infrastructure:

- `/static/*` → `{region}-assets.i.posthog.com` (JS SDK files)
- `/*` → `{region}.i.posthog.com` (events, feature flags, etc.)

## Deployment

This is configured for [Fly.io](https://fly.io):

```bash
fly launch --no-deploy
fly deploy
```

The app will auto-stop when idle and auto-start on incoming requests to minimize costs.

### CI/CD

Pushes to `main` automatically run tests and deploy to Fly.io via GitHub Actions.

To enable automatic deploys, add `FLY_API_TOKEN` to your repository secrets:

1. Generate a token: `fly tokens create deploy -x 999999h`
2. Add it to GitHub: Settings → Secrets and variables → Actions → New repository secret

## Configuration

### Environment Variables

| Variable         | Default | Description                  |
| ---------------- | ------- | ---------------------------- |
| `POSTHOG_REGION` | `us`    | PostHog region: `us` or `eu` |

Set the region in `fly.toml` or via `fly secrets set POSTHOG_REGION=eu`.

### Client Setup

Configure the PostHog SDK to use your proxy:

```javascript
posthog.init("YOUR_PROJECT_KEY", {
  api_host: "https://your-app-name.fly.dev",
  ui_host: "https://us.posthog.com", // or eu.posthog.com
});
```

## Local Development

Requires Go 1.22. Use [mise](https://mise.jdx.dev) to install:

```bash
mise install
```

### Make Targets

| Command      | Description              |
| ------------ | ------------------------ |
| `make build` | Compile the binary       |
| `make test`  | Run tests                |
| `make run`   | Run the proxy locally    |
| `make clean` | Remove the built binary  |

## Verification

1. Check the health endpoint: `curl https://your-app-name.fly.dev/health`
2. Configure your app's PostHog SDK with the proxy URL
3. Verify events appear in your PostHog dashboard
