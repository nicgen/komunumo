# Local infra (Traefik + mkcert)

Provides HTTPS access to backend/frontend on `*.local.hello-there.net`.

## Prerequisites

- `mkcert` installed and `mkcert -install` already run (CA trusted).
- `/etc/hosts` entries:

  ```text
  127.0.0.1 app.local.hello-there.net
  127.0.0.1 api.local.hello-there.net
  ```

## Generate certs

```bash
cd infra/local
mkdir -p certs
mkcert -cert-file certs/local.hello-there.net.pem \
       -key-file  certs/local.hello-there.net-key.pem \
       "local.hello-there.net" "*.local.hello-there.net"
```

## Run

```bash
docker compose up
```

- Frontend: <https://app.local.hello-there.net>
- Backend:  <https://api.local.hello-there.net>
- Traefik dashboard: <http://localhost:8081>
