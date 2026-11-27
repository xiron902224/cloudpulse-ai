# Plan de sprints — CloudPulse AI

Capacidad: **un desarrollador**, 8 sprints de 5 días laborables (~6 h netas/día).  
Objetivo: un stack Compose que recoja CPU real, la persista, la muestre en UI y dispare alertas. Lo que el README promete (K8s, LSTM, Groq, multi-cloud, ChatOps) **no entra** en estos 40 días.

Cada día cierra con: código formateado, prueba (`test` o `curl`) anotada, y `docs/API.md` / Swagger actualizados si cambió el contrato.

**Fuentes internas (leer antes del sprint 1):**

- [ARCHITECTURE.md](./ARCHITECTURE.md) — componentes, puertos, red Docker
- [API.md](./API.md) — contrato HTTP actual
- [README.md](../README.md) — quick start (hay que reescribirlo en S1; no fiarse de las métricas inventadas)

---

## Cómo usar este documento

Para cada día:

1. **Dominar** los enlaces de *Documentación a dominar* (no solo ojearlos).
2. **Entregar** el criterio de hecho.
3. **Resolver** siguiendo *Cómo*. Si te atascas, usa *Extra*.
4. No pases de sprint si el *Done when* del sprint anterior no se cumple.

Onboarding (4 h, día 0, antes del lunes S1):

| Tema | Qué leer |
| --- | --- |
| Mapa del repo | Este archivo + Architecture + API |
| Go API | [Fiber routing](https://docs.gofiber.io/api/app), [Ctx](https://docs.gofiber.io/api/ctx), [testing](https://docs.gofiber.io/guide/testing) |
| Postgres | [pgx pool](https://pkg.go.dev/github.com/jackc/pgx/v5/pgxpool) |
| Observabilidad | [Prometheus metric types](https://prometheus.io/docs/concepts/metric_types/) |

---

## Sprint 1 — Hacer que arranque

**Meta:** un clon + `.env` + `docker compose up postgres redis api` responde. Cero features nuevas.

**Done when:** `GET /health` y `GET /ready` correctos; `gofmt` limpio; la API arranca en Compose **sin** bind-mount que pise el binario.

### Lunes S1 — Alinear repo y docs honestas

**Entregar:** `go.mod` / imports / `gofmt` coherentes; README que describa lo que el repo hace.

**Cómo:**

1. Fijar Go a **1.23** (o la versión real del [Dockerfile](../services/api-go/Dockerfile)); no dejar `go 1.27.0` si el toolchain del equipo no existe.
2. Un solo module path: el de `go.mod` (`github.com/xiron902224/cloudpulse-ai/api`) o el de GitHub (`xiron9090`). Cambiar imports y documentar el elegido.
3. `gofmt -w .` en `services/api-go`. Quitar tags `gorm` de `models.Metric` (no hay GORM).
4. Reescribir Overview del README: API Go + Postgres + Redis + Prometheus. Quitar cifras inventadas (P99 50 ms, 10k rps) y badges que no existen (Codecov, K8s) o marcarlos *planned*.
5. `make install` solo `go mod download`. Rust/Node se anotan para sprints 2–3.

**Documentación a dominar:**

- [Go modules](https://go.dev/doc/modules)
- [gofmt](https://pkg.go.dev/cmd/gofmt)
- [ARCHITECTURE.md](./ARCHITECTURE.md)
- [API.md](./API.md)

**Extra:** [Managing dependencies](https://go.dev/doc/modules/managing-dependencies) · [Semantic Versioning](https://semver.org/)

---

### Martes S1 — Bug `UpdateAlert` (B01)

**Entregar:** `PUT /api/v1/alerts/:id` actualiza la fila correcta; test que falle si `$1` vuelve a ser el timestamp.

**Cómo:**

El handler arma `UPDATE alerts SET updated_at = $1` y al final añade `WHERE id = $1`. El primer placeholder ya es `updated_at`, así que el `id` nunca filtra bien.

1. Escribir el test **antes** (Fiber `app.Test` + Postgres de Compose, o sqlmock).
2. Reconstruir: `SET status=$2, severity=$3, message=$4, updated_at=$5 WHERE id=$1`.
3. Validar `status` / `severity` igual que en `CreateAlert`.
4. 404 si no existe; 200 con la fila completa.

**Documentación a dominar:**

- Código: [`alerts.go`](../services/api-go/internal/handlers/alerts.go)
- [Fiber Params](https://docs.gofiber.io/api/ctx#params)
- [pgx Query / Exec](https://pkg.go.dev/github.com/jackc/pgx/v5#hdr-Query)

**Extra:** [sqlc named params](https://docs.sqlc.dev/en/latest/howto/named_parameters.html) (patrón, aunque no uses sqlc) · [table-driven tests](https://go.dev/wiki/TableDrivenTests)

---

### Miércoles S1 — Redis nil-safe, cache, `total` (B05–B07)

**Entregar:** sin Redis la API sigue; crear/borrar métrica invalida listas; `total` es `COUNT(*)`.

**Cómo:**

1. En `Connect`, si `Ping` falla: `RedisClient{Client: nil}` (hoy se devuelve un cliente vivo y `Get`/`Set` pegan a Redis muerto).
2. Invalidar `metric:{id}` y `metrics:list:*` (SCAN + UNLINK) o un entero `cache_gen`. No basta con borrar `metrics:latest`.
3. `ListMetrics`: `SELECT COUNT(*) FROM metrics` + `LIMIT/OFFSET`.
4. Test de cache con [miniredis](https://github.com/alicebob/miniredis).

**Documentación a dominar:**

- Código: [`redis.go`](../services/api-go/internal/cache/redis.go), [`metrics.go`](../services/api-go/internal/handlers/metrics.go)
- [SCAN](https://redis.io/docs/latest/commands/scan/)
- [UNLINK](https://redis.io/docs/latest/commands/unlink/)
- [go-redis](https://github.com/redis/go-redis)

**Extra:** [Cache stampede](https://en.wikipedia.org/wiki/Cache_stampede) · [Redis keys](https://redis.io/docs/latest/commands/keys/) (por qué no usar `KEYS` en prod)

---

### Jueves S1 — Ready, rate limit, cardinalidad Prometheus (B08–B10)

**Entregar:** `/health` liveness; `/ready` pinea DB y Redis (503 si Postgres cae); rate limit 100/min; series Prometheus por **ruta**, no por UUID.

**Cómo:**

1. `HealthCheck` no toca dependencias.
2. `ReadyHandler`: `pool.Ping` + Redis (si `Client != nil`).
3. En el middleware, `c.Route().Path` en vez de `c.Path()`.
4. Limiter: `Expiration: 1 * time.Minute`; mensaje alineado.
5. En `prometheus.yml`, job `prometheus`: `targets: ['localhost:9090']` (dentro del contenedor el puerto es 9090, no 9091).

**Documentación a dominar:**

- [Probes liveness vs readiness](https://kubernetes.io/docs/concepts/workloads/pods/pod-lifecycle/#container-probes) (concepto; no hace falta K8s)
- [Prometheus naming](https://prometheus.io/docs/practices/naming/)
- [Fiber Route](https://docs.gofiber.io/api/ctx#route)
- [Fiber limiter](https://docs.gofiber.io/api/middleware/limiter)
- Config: [`prometheus.yml`](../deployment/prometheus/prometheus.yml)

**Extra:** [Cardinality](https://grafana.com/blog/2022/02/15/what-are-cardinality-spikes-and-why-do-they-matter/) · [Histogram buckets](https://prometheus.io/docs/practices/histograms/)

---

### Viernes S1 — Compose API + migraciones

**Entregar:** demo `curl`; tablas vía SQL de migración, no solo `CREATE TABLE` al boot; Compose API sin volume que pise `/api`.

**Cómo:**

1. `services/api-go/migrations/0001_init.sql` con el DDL actual.
2. Ejecutar migraciones con [golang-migrate](https://github.com/golang-migrate/migrate) o `psql -f` documentado en Makefile.
3. Quitar `volumes: ./services/api-go:/app` del servicio `api` en [docker-compose.yml](../docker-compose.yml).
4. `docker compose up -d postgres redis api` → `curl /health` y `/ready`.
5. [scripts/seed-data.sh](../scripts/seed-data.sh) contra API viva.

**Documentación a dominar:**

- [golang-migrate](https://github.com/golang-migrate/migrate)
- [Compose depends_on + healthcheck](https://docs.docker.com/compose/how-tos/startup-order/)
- [Dockerfile de la API](../services/api-go/Dockerfile)

**Extra:** [Multi-stage builds](https://docs.docker.com/build/building/multi-stage/) · [12-factor config](https://12factor.net/config)

---

## Sprint 2 — Collector Rust

**Meta:** daemon que muestre el host, haga POST al contrato `Metric`, exponga `/health` y `/metrics` en `:9090`.

**Done when:** `GET /api/v1/metrics` muestra muestras nuevas cada ~10 s con Compose.

### Lunes S2 — Crate compilable

**Entregar:** `edition = "2021"`, módulos, `cargo test` + `clippy -D warnings`. `Cargo.lock` versionado (quitarlo de `.gitignore`).

**Cómo:**

1. `main.rs` deja de ser `Hello, world!`. Módulos `collector/{mod,cpu,memory,disk,network}.rs`.
2. Verificar API de [sysinfo 0.30](https://docs.rs/sysinfo/0.30.13/sysinfo/): `refresh_cpu_usage` vs `refresh_cpu`; `global_cpu_info` vs `global_cpu_usage`. El `cpu.rs` actual puede estar desfasado.
3. CI: [dtolnay/rust-toolchain](https://github.com/dtolnay/rust-toolchain) en vez de `actions-rs` (deprecado).

**Documentación a dominar:**

- [The Cargo Book — TOML vs lock](https://doc.rust-lang.org/cargo/guide/cargo-toml-vs-cargo-lock.html)
- [Packages, crates, modules](https://doc.rust-lang.org/book/ch07-00-managing-growing-projects.html)
- [sysinfo 0.30](https://docs.rs/sysinfo/0.30.13/sysinfo/)
- Código: [`cpu.rs`](../services/collector-rust/src/collector/cpu.rs), [`Cargo.toml`](../services/collector-rust/Cargo.toml)

**Extra:** [Clippy lints](https://doc.rust-lang.org/clippy/) · [edition guide](https://doc.rust-lang.org/edition-guide/)

---

### Martes S2 — Memoria, disco, red

**Entregar:** `HostSample` serde; porcentajes 0–100; red en **KiB/s** (delta / intervalo), documentado en API.md.

**Cómo:**

1. Memoria: `used/total * 100`. Disco: partición `/` o env `DISK_PATH`.
2. Red: no enviar contadores acumulados; `delta_bytes / dt`.
3. Tests de rango para porcentajes.

**Documentación a dominar:**

- [API.md POST /metrics](./API.md)
- [Metric JSON](../services/api-go/internal/models/metric.go)
- [sysinfo Disks](https://docs.rs/sysinfo/0.30.13/sysinfo/struct.Disks.html) · [Networks](https://docs.rs/sysinfo/0.30.13/sysinfo/struct.Networks.html)
- [serde](https://serde.rs/)

**Extra:** [chrono serde](https://docs.rs/chrono) · unidades IEC vs SI

---

### Miércoles S2 — Cliente HTTP

**Entregar:** loop Tokio cada `COLLECT_INTERVAL_SECS` (default 10); retry 3× backoff en 5xx/timeout; `tracing`; sin `unwrap` en el path de prod.

**Env:** `API_URL`, `COLLECT_INTERVAL_SECS`, `HOST_ID` (hostname).

**Documentación a dominar:**

- [Tokio interval](https://docs.rs/tokio/latest/tokio/time/fn.interval.html)
- [reqwest JSON](https://docs.rs/reqwest)
- [tracing](https://docs.rs/tracing)
- [anyhow](https://docs.rs/anyhow)

**Extra:** [wiremock](https://docs.rs/wiremock) para tests · [exponential backoff](https://aws.amazon.com/blogs/architecture/exponential-backoff-and-jitter/)

---

### Jueves S2 — Axum `:9090` + Dockerfile

**Entregar:** `/health`, `/metrics` Prometheus; gauges `last_cpu`, counters `post_success` / `post_fail`; imagen multi-stage non-root + HEALTHCHECK.

**Documentación a dominar:**

- [Axum 0.7](https://docs.rs/axum/0.7.9/axum/)
- [prometheus crate](https://docs.rs/prometheus) o [metrics-exporter-prometheus](https://docs.rs/metrics-exporter-prometheus)
- [Multi-stage Docker](https://docs.docker.com/build/building/multi-stage/)
- Job `collector-rust` en [prometheus.yml](../deployment/prometheus/prometheus.yml) (ya apunta a `collector:9090`)

**Extra:** [distroless](https://github.com/GoogleContainerTools/distroless) · [HEALTHCHECK](https://docs.docker.com/reference/dockerfile/#healthcheck)

---

### Viernes S2 — Compose collector

**Entregar:** servicio sin bind-mount de código; `depends_on` API healthy; `make dev-rust` con `API_URL=http://localhost:8080`.

**Documentación a dominar:**

- [docker-compose.yml collector](../docker-compose.yml)
- [ARCHITECTURE.md puerto 9090](./ARCHITECTURE.md)
- [dtolnay/rust-toolchain](https://github.com/dtolnay/rust-toolchain)

**Extra:** [Compose build](https://docs.docker.com/compose/compose-file/build/)

---

## Sprint 3 — Frontend dashboard

**Meta:** Vite + React 18 + TypeScript. Polling 10 s. Sin WebSocket ni ChatOps.

**Done when:** UI en `:3000` contra API `:8080` muestra charts que se mueven con el collector.

### Lunes S3 — Scaffold que pase CI

**Entregar:** `frontend/` con `npm ci`, `lint`, `test` (vitest), `build`; Dockerfile.

**Cómo:**

1. `npm create vite@latest frontend -- --template react-ts`.
2. Node 20 Alpine; stage nginx o `vite preview`.
3. Proxy Vite `/api` → `:8080` (evita CORS en dev).
4. El job [CI test-frontend](../.github/workflows/ci.yml) deja de romper.

**Documentación a dominar:**

- [Vite guide](https://vitejs.dev/guide/)
- [Vite env](https://vitejs.dev/guide/env-and-mode.html)
- [Vitest](https://vitest.dev/guide/)
- [ESLint](https://eslint.org/docs/latest/use/getting-started)

**Extra:** [TypeScript handbook](https://www.typescriptlang.org/docs/handbook/intro.html) · [nginx SPA try_files](https://nginx.org/en/docs/http/ngx_http_core_module.html#try_files)

---

### Martes S3 — Cliente API + Health

**Entregar:** tipos `Metric` / `Alert` alineados al JSON; `fetchMetrics` / `fetchAlerts`; errores 4xx/5xx visibles; loading y empty.

**Documentación a dominar:**

- [API.md](./API.md)
- [Fetch API](https://developer.mozilla.org/en-US/docs/Web/API/Fetch_API)
- [TypeScript narrowing](https://www.typescriptlang.org/docs/handbook/2/narrowing.html)

**Extra:** [AbortController](https://developer.mozilla.org/en-US/docs/Web/API/AbortController)

---

### Miércoles S3 — Dashboard (4 series + tabla)

**Entregar:** CPU / mem / disk / net; eje X = `timestamp * 1000`; poll 10 s abortado en unmount; layout sidebar alertas + main charts.

`timestamp` del backend es **unix seconds**, no ms.

**Documentación a dominar:**

- [Recharts LineChart](https://recharts.org/en-US/api/LineChart)
- [React useEffect cleanup](https://react.dev/learn/synchronizing-with-effects)

**Extra:** [uPlot](https://github.com/leeoniya/uPlot) si Recharts se queda corto · [date-fns](https://date-fns.org/)

---

### Jueves S3 — Vista alertas

**Entregar:** lista, filtro severity/status, acknowledge con `PUT` (depende de B01).

**Documentación a dominar:**

- [API.md Alerts](./API.md)
- [alert.go constantes](../services/api-go/internal/models/alert.go)

**Extra:** [HTTP PUT semantics](https://developer.mozilla.org/en-US/docs/Web/HTTP/Reference/Methods/PUT)

---

### Viernes S3 — Compose frontend

**Entregar:** servicio frontend; **quitar `VITE_WS_URL`** hasta que exista WS; bind-mount `node_modules` solo en override de dev.

**Verificar en el navegador** el flujo collector → API → chart (no basta un screenshot).

**Documentación a dominar:**

- [Compose frontend](../docker-compose.yml)
- [Vite proxy](https://vitejs.dev/config/server-options.html#server-proxy)

**Extra:** [CORS](https://developer.mozilla.org/en-US/docs/Web/HTTP/CORS)

---

## Sprint 4 — Timescale, queries y motor de alertas

**Meta:** time-series de verdad + umbrales. Schema change: el riesgo más alto del plan.

**Done when:** gráfica de 1 h; alerta automática si CPU > 80 durante 60 s.

### Lunes S4 — Migración hypertable + `host_id`

**Entregar:** `0002_timescale.sql`; `host_id TEXT NOT NULL`; índice `(host_id, created_at DESC)`; collector envía `host_id`.

**Cómo:**

1. Columna `timestamptz` de tiempo (o `created_at`) y `create_hypertable(..., migrate_data => true)`.
2. Alertas: `host_id`, `source`.
3. Extender el JSON de `Metric` (hoy el API ignora campos extra).

**Documentación a dominar:**

- [Hypertables](https://docs.timescale.com/use-timescale/latest/hypertables/)
- [Schema management](https://docs.timescale.com/use-timescale/latest/schema-management/)
- [create_hypertable](https://docs.timescale.com/api/latest/hypertable/create_hypertable/)

**Extra:** [PostgreSQL ALTER TABLE](https://www.postgresql.org/docs/current/sql-altertable.html) · [UUID vs composite PK en hypertables](https://docs.timescale.com/use-timescale/latest/hypertables/about-hypertables/#primary-keys)

---

### Martes S4 — `GET /api/v1/metrics/query`

**Query:** `from`, `to`, `host_id`, `step`. Validar `from < to`, rango máx. 7 días. Agregar con `time_bucket` si hay `step`. Frontend del chart 1 h usa **este** endpoint, no el list paginado.

**Documentación a dominar:**

- [time_bucket](https://docs.timescale.com/api/latest/hyperfunctions/time_bucket/)
- [Fiber query](https://docs.gofiber.io/api/ctx#query)
- [API.md](./API.md) (actualizar)

**Extra:** [time_bucket_gapfill](https://docs.timescale.com/api/latest/hyperfunctions/gapfilling/time_bucket_gapfill/)

---

### Miércoles S4 — `alert_rules` + evaluator

**Entregar:** tabla de reglas; goroutine cada 15 s; dedup `host + rule` si ya hay alerta `active`.

Regla ejemplo: `cpu_usage > 80` durante 60 s, severity `warning`. Config en SQL seed o YAML. **Sin ML.**

**Documentación a dominar:**

- [Prometheus alerting (`for:`)](https://prometheus.io/docs/prometheus/latest/configuration/alerting_rules/) (modelo mental)
- [Go time.Ticker](https://pkg.go.dev/time#Ticker)

**Extra:** [SELECT FOR UPDATE skip locked](https://www.postgresql.org/docs/current/sql-select.html#SQL-FOR-UPDATE-SHARE) si más adelante hay varios evaluators

---

### Jueves S4 — CRUD rules + UI umbral

**Entregar:** `POST/GET /api/v1/alert-rules`; form de threshold; `GET /alerts/:id`. Probar con carga de CPU (`yes >/dev/null`) y ver la alerta en UI.

**Documentación a dominar:** [API.md](./API.md) · vista alertas del frontend

**Extra:** [stress-ng](https://wiki.debian.org/StressTesting) para carga controlada

---

### Viernes S4 — Retention 14 días

**Entregar:** `add_retention_policy` (o cron SQL); documentado en Architecture.

**Documentación a dominar:**

- [Data retention](https://docs.timescale.com/use-timescale/latest/data-retention/)
- [add_retention_policy](https://docs.timescale.com/api/latest/data-retention/add_retention_policy/)

---

## Sprint 5 — Auth y aislamiento

Alcance: **API keys por tenant**, no OAuth/SSO.

**Done when:** sin key → 401; tenant A no ve filas de B.

### Lunes S5 — `tenants`, `api_keys`, Bearer

**Entregar:** hash SHA-256 de la key; prefix visible `cp_live_xxxx`; header `Authorization: Bearer`. Públicas: `/health`, `/ready`, `/metrics`, `/swagger`.

**Documentación a dominar:**

- [OWASP REST Security](https://cheatsheetseries.owasp.org/cheatsheets/REST_Security_Cheat_Sheet.html)
- [OWASP API Security Top 10 (2023)](https://owasp.org/API-Security/editions/2023/en/0x11-t10/)
- [crypto/sha256](https://pkg.go.dev/crypto/sha256)

**Extra:** [CWE-798 hard-coded credentials](https://cwe.mitre.org/data/definitions/798.html) · [Timing-safe compare](https://pkg.go.dev/crypto/subtle#ConstantTimeCompare)

---

### Martes S5 — `tenant_id` en todas las queries

**Entregar:** FK en metrics, alerts, rules; test key A vs B.

**Documentación a dominar:** [postgres.go](../services/api-go/internal/database/postgres.go) · [Row-level security](https://www.postgresql.org/docs/current/ddl-rowsecurity.html) (opcional, extra)

---

### Miércoles S5 — CORS allowlist + rate limit por key

**Entregar:** `FRONTEND_ORIGIN` en env; limiter `KeyGenerator` = hash de API key (fallback IP).

**Documentación a dominar:**

- [Fiber CORS](https://docs.gofiber.io/api/middleware/cors)
- [X-Forwarded-For pitfalls](https://owasp.org/www-community/attacks/HTTP_Response_Splitting) (no confiar IP de proxy sin config)

**Extra:** [RFC 9110 Authorization](https://www.rfc-editor.org/rfc/rfc9110.html#name-authorization)

---

### Jueves S5 — UI: pegar API key (localStorage, demo)

**Entregar:** pantalla settings; nunca commitear keys; `.env.example` con valor fake.

**Documentación a dominar:** [Vite env](https://vitejs.dev/guide/env-and-mode.html)

**Extra:** [XSS y localStorage](https://owasp.org/www-community/attacks/xss/) — aceptable para demo, no para prod real

---

### Viernes S5 — Collector y seed con `Authorization`

**Entregar:** scripts y collector envían Bearer. Decidir: scrape Prometheus sin auth, o [basic auth](https://prometheus.io/docs/guides/basic-auth/).

**No tocar Terraform.**

---

## Sprint 6 — Tests, CI, Grafana

**Done when:** CI en `main` verde; dashboard Grafana muestra RPS de la API.

### Lunes S6 — Tests de handlers reales

**Entregar:** testcontainers Postgres + Redis (o service containers en CI); **borrar** tests que reimplementan validación en closures. Cobertura > 60 % en `handlers`, `cache`, `database`.

**Documentación a dominar:**

- [testcontainers-go](https://golang.testcontainers.org/)
- [Fiber testing](https://docs.gofiber.io/guide/testing)
- Tests actuales (a reemplazar): [`metrics_test.go`](../services/api-go/internal/handlers/metrics_test.go)

**Extra:** [testing.Short](https://pkg.go.dev/testing#Short) · [GitHub Actions services](https://docs.github.com/en/actions/using-containerized-services/about-service-containers)

---

### Martes S6 — CI

**Entregar:** rust-toolchain; frontend existe; `docker compose build` solo servicios con Dockerfile. Codecov puede seguir con `fail_ci_if_error: false`.

**Documentación a dominar:** [ci.yml](../.github/workflows/ci.yml) · [codecov-action](https://github.com/codecov/codecov-action)

**Extra:** [actions/setup-go](https://github.com/actions/setup-go) cache

---

### Miércoles S6 — Grafana provisioning

**Entregar:** datasource Prometheus `http://prometheus:9090`; provider de dashboards; volumes en Compose.

Hoy el JSON está suelto en `deployment/grafana/dashboards` **sin** provisioning: Grafana no lo carga solo.

**Documentación a dominar:**

- [Grafana provisioning](https://grafana.com/docs/grafana/latest/administration/provisioning/)
- [Dashboard JSON](../deployment/grafana/dashboards/api-dashboard.json)
- [PromQL basics](https://prometheus.io/docs/prometheus/latest/querying/basics/)

**Extra:** [Grafana HTTP API](https://grafana.com/docs/grafana/latest/developers/http_api/)

---

### Jueves S6 — Makefile

**Entregar:** `test`, `test-integration`, `coverage`, `bootstrap`, `fmt`. Quitar `test-e2e` del README o añadir un smoke Playwright mínimo.

**Documentación a dominar:** [Makefile](../Makefile) · [GNU make](https://www.gnu.org/software/make/manual/make.html#Introduction)

---

### Viernes S6 — Swagger = código = API.md

**Entregar:** `swag init`; Architecture mermaid al día; `ENABLE_AI` en `.env.example` marcado unused o eliminado.

**Documentación a dominar:** [swag](https://github.com/swaggo/swag) · [fiber-swagger](https://github.com/swaggo/fiber-swagger)

---

## Sprint 7 — IA thin slice (sin LSTM)

Un z-score/IQR sobre ventana de 1 h. **No** Groq, ONNX ni servicio ML completo.

**Done when:** alerta `info` cuando CPU sale de banda y el chart pinta los puntos.

### Lunes S7 — Especificación

**Entregar:** `docs/ANOMALY.md`: anomalía = `|x − mean| > 3σ` en 60 puntos por `host_id` + métrica. Seed con un spike artificial.

**Documentación a dominar:**

- [Standard score](https://en.wikipedia.org/wiki/Standard_score)
- [Robust statistics / IQR](https://en.wikipedia.org/wiki/Interquartile_range) (alternativa si hay outliers en el baseline)

**Extra:** [Numenta NAB](https://github.com/numenta/NAB) (concepto; no copiar código)

---

### Martes S7 — `GET /api/v1/anomalies`

**Entregar:** SQL `avg`/`stddev` (Timescale/Postgres); lista de puntos + score.

**Documentación a dominar:** [Aggregate functions](https://www.postgresql.org/docs/current/functions-aggregate.html)

**Extra:** [Window functions](https://www.postgresql.org/docs/current/tutorial-window.html)

---

### Miércoles S7 — Integrar evaluator

**Entregar:** alerta `info` `anomaly detected`, dedup 15 min, reutilizando el motor del sprint 4.

---

### Jueves S7 — UI puntos anómalos

**Entregar:** dots en el chart + lista.

**Documentación a dominar:** [Recharts ReferenceDot](https://recharts.org/en-US/api/ReferenceDot)

---

### Viernes S7 — Opcional FastAPI

**Entregar:** `ENABLE_ML=true` llama a `services/ml-python`; si no hay tiempo, interfaz Go + stub. **No HuggingFace.**

**Documentación a dominar:** [FastAPI](https://fastapi.tiangolo.com/) · [httpx](https://www.python-httpx.org/)

---

## Sprint 8 — Release local

**Done when:** `git clone` + `make bootstrap` deja una demo de 15 minutos.

### Lunes S8 — `make bootstrap`

**Entregar:** [setup.sh](../scripts/setup.sh) no interactivo; Compose + seed + healthgate. [deploy.sh](../scripts/deploy.sh): quitar prompts `s/n` o mover a `docs/DEPLOY.md` como **no soportado** (Vercel/Railway/Fly con URLs inventadas).

---

### Martes S8 — README y roadmap honestos

**Entregar:** features = checkboxes de lo shipped. `docs/ROADMAP.md`: K8s, Groq, LSTM, multi-cloud, ChatOps = later.

---

### Miércoles S8 — Hardening Compose

**Entregar:** passwords por env; Grafana admin por secret. Terraform: borrar o marcar experimental y **quitar** `publicly_accessible` + SSH `0.0.0.0/0` de [`main.tf`](../deployment/terraform/main.tf).

**Documentación a dominar:** [Compose secrets/env](https://docs.docker.com/compose/how-tos/use-secrets/) · [CIS Docker](https://www.cisecurity.org/benchmark/docker) (ojeo)

---

### Jueves S8 — Runbook

**Entregar:** `docs/RUNBOOK.md` (una página): API unhealthy, Redis nil, collector 401, Prometheus target down.

**Documentación a dominar:** [ARCHITECTURE.md puertos](./ARCHITECTURE.md)

---

### Viernes S8 — QA 15 min y tag `v0.2.0`

Flujo: collector → API → UI → alerta umbral → anomalía spike → Grafana RPS (`:3001`).

---

## Fuera de alcance (no planificar días)

| Tema | Por qué no |
| --- | --- |
| Manifests Kubernetes | No hay imágenes/probes estables; Compose cubre a un dev |
| Terraform AWS RDS público | Inseguro e irrelevante al MVP |
| LSTM / Groq / ONNX / ChatOps | Sin datos etiquetados; el z-score basta |
| WebSocket live | Polling 10 s hasta que el collector sea estable |
| Agentes AWS/GCP/Azure | El collector es `sysinfo` local |
| Deploy Vercel + Railway + Fly | `deploy.sh` asume CLIs y URLs que no existen |

---

## Bugs de referencia (sprint 1–2)

| ID | Sev | Dónde | Qué |
| --- | --- | --- | --- |
| B01 | P0 | `alerts.go` `UpdateAlert` | `WHERE id = $1` es `updated_at` |
| B02 | P0 | `collector` `main.rs` | No recolecta ni envía |
| B03 | P0 | Compose | Dockerfiles FE/collector faltan; bind-mount pisa binario |
| B04 | P0 | CI frontend | `frontend/` no existe |
| B05 | P1 | `redis.go` | Ping fail → cliente vivo |
| B06 | P1 | cache métricas | Invalida keys que nadie usa |
| B07 | P1 | `ListMetrics` | `total` = len(página) |
| B08 | P1 | `prometheus.yml` | Self-scrape `:9091` |
| B09 | P1 | `/health` | Siempre healthy |
| B10 | P2 | rate limit | 100 req / 60 min; el mensaje dice 1 min |
| B11 | P2 | `go.mod` / `main.go` | Module vs GitHub; `go 1.27`; tags gorm |
| B12 | P2 | tests handlers | No ejecutan el handler real |

---

## Definición de hecho (cada día)

- `gofmt` / `rustfmt` / `eslint` limpios en lo tocado
- Test o `curl` documentado en el mensaje de commit/PR
- Swagger o [API.md](./API.md) si cambia el contrato
- No merge con el job de CI del componente en rojo
