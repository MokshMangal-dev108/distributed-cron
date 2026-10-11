# Distributed Cron

A fault-tolerant, horizontally scalable cron job scheduler written in Go.

Distributed Cron lets you register jobs through a REST API, schedules them using standard cron expressions, and executes each one **exactly once per tick**, even when multiple scheduler instances are running, instances crash mid-run, or downstream calls fail.

---

## Table of Contents

- [Features](#features)
- [Architecture](#architecture)
- [How It Works](#how-it-works)
- [Tech Stack](#tech-stack)
- [Project Structure](#project-structure)
- [Getting Started](#getting-started)
- [Configuration](#configuration)
- [API Reference](#api-reference)
- [Reliability Guarantees](#reliability-guarantees)
- [Observability](#observability)
- [Testing](#testing)
- [Deployment](#deployment)

---

## Features

- **Cron expressions**: standard 5-field syntax (`*/5 * * * *`) for defining schedules
- **REST API**: create, list, update, pause, resume and delete jobs over HTTP
- **Persistent storage**: all jobs and execution history are stored in PostgreSQL, so nothing is lost on restart
- **Multiple scheduler instances**: run as many schedulers as you like for high availability
- **Distributed locking**: Redis-based locks prevent duplicate execution across schedulers
- **Worker pool**: scheduling and execution are decoupled; workers scale independently
- **Retries with exponential backoff**: transient failures are retried automatically
- **Dead-letter handling**: jobs that exhaust their retries are parked for inspection
- **Idempotent execution**: a crash between "work done" and "marked complete" never causes double processing
- **Crash recovery**: expired locks and missed runs are picked up automatically
- **Graceful shutdown**: in-flight jobs are allowed to finish before a process exits
- **Job status and history**: full audit trail of every execution attempt
- **Metrics and logging**: structured logs and Prometheus metrics with Grafana dashboards
- **Containerized**: Docker Compose for local development, Kubernetes manifests for production

---

## Architecture

```
                    ┌──────────────────┐
                    │      Client      │
                    └────────┬─────────┘
                             │ HTTP
                             ▼
                    ┌──────────────────┐
                    │     REST API     │
                    └────────┬─────────┘
                             │
                             ▼
                    ┌──────────────────┐
                    │    PostgreSQL    │  jobs, executions, history
                    └────────┬─────────┘
                             │
              ┌──────────────┼──────────────┐
              ▼              ▼              ▼
        Scheduler 1    Scheduler 2    Scheduler 3
              │              │              │
              └──────────────┼──────────────┘
                             │  SET NX EX (claim job)
                             ▼
                      ┌─────────────┐
                      │    Redis    │  distributed locks
                      └──────┬──────┘
                             │
                             ▼
                        ┌─────────┐
                        │  Queue  │
                        └────┬────┘
                             │
              ┌──────────────┼──────────────┐
              ▼              ▼              ▼
          Worker 1       Worker 2       Worker 3
              │              │              │
              └──────────────┼──────────────┘
                             ▼
                      Execute Job ──► Update DB ──► Metrics ──► Grafana
```

### Components

| Component | Responsibility |
|-----------|----------------|
| **API** | Accepts job definitions, validates cron expressions, exposes job status and history |
| **Scheduler** | Finds due jobs, claims them via a distributed lock, and enqueues them. It does not run the job itself |
| **Redis** | Provides distributed locks with TTLs so only one scheduler claims a given job run |
| **Queue** | Buffers due jobs between schedulers and workers |
| **Worker** | Pulls jobs from the queue, executes them, handles retries, and records the result |
| **PostgreSQL** | Source of truth for jobs, schedules, execution records, and dead-lettered jobs |

Keeping the scheduler and worker roles separate means a slow or failing job never delays the scheduling loop, and each side can be scaled on its own.

---

## How It Works

### 1. Scheduling

Every job has a cron expression and a computed `next_run_at`. Each scheduler runs a loop that polls for due jobs:

```sql
SELECT * FROM jobs
WHERE next_run_at <= NOW()
  AND status = 'active';
```

Once a job is claimed and enqueued, its `next_run_at` is advanced to the next tick according to its cron expression.

### 2. Claiming a job (distributed locking)

When several schedulers see the same due job, only one may enqueue it. Each scheduler tries to take a lock in Redis:

```
SET lock:job:<job_id> <scheduler_id> NX EX 30
```

- The scheduler that gets `OK` owns the run and enqueues the job.
- Every other scheduler gets a failure response and skips it.

### 3. Surviving crashes

Locks are created with a **30 second TTL**. If a scheduler dies after claiming a job, the lock expires on its own and another scheduler can pick the job up. There are no permanent locks and no manual cleanup.

### 4. Idempotent execution

A crash can still happen *after* a job did its work but *before* it was marked complete. To make re-delivery safe, every execution carries an idempotency key:

```
idempotency_key = job_id + execution_id
```

Before running, a worker checks whether that key has already been processed. If it has, the work is skipped and the execution is simply marked complete. Jobs that move money, send emails, or call external APIs are never executed twice for the same run.

### 5. Retries and dead letters

If a job fails, the worker retries it with exponential backoff instead of giving up immediately:

```
Attempt 1  ──fail──►  wait 5s
Attempt 2  ──fail──►  wait 15s
Attempt 3  ──fail──►  wait 60s
                       │
                       ▼
                  Dead Letter
```

Jobs that exhaust their attempts are moved to the dead-letter store with their last error, so they can be inspected and replayed.

### 6. Graceful shutdown

On `SIGINT` / `SIGTERM`, schedulers stop claiming new work and workers finish their in-flight jobs (up to a configurable timeout) before exiting. Anything left unfinished is recovered through lock expiry.

---

## Tech Stack

- **Language:** Go
- **Database:** PostgreSQL
- **Locking and queueing:** Redis
- **Metrics:** Prometheus and Grafana
- **Packaging:** Docker, Docker Compose
- **Orchestration:** Kubernetes

---

## Project Structure

```
distributed-cron/
├── cmd/
│   ├── api/              # REST API entrypoint
│   ├── scheduler/        # Scheduler entrypoint
│   └── worker/           # Worker entrypoint
├── internal/
│   ├── api/              # HTTP handlers and routing
│   ├── job/              # Job model, cron parsing, repository
│   ├── scheduler/        # Scheduling loop and job claiming
│   ├── lock/             # Redis distributed lock
│   ├── queue/            # Queue producer and consumer
│   ├── worker/           # Execution, retries, idempotency
│   ├── metrics/          # Prometheus instrumentation
│   └── config/           # Configuration loading
├── migrations/           # PostgreSQL schema migrations
├── deploy/
│   ├── docker/           # Dockerfiles
│   └── k8s/              # Kubernetes manifests
├── docker-compose.yml
├── go.mod
└── README.md
```

---

## Getting Started

### Prerequisites

- Go 1.22+
- Docker and Docker Compose

### Run the full stack with Docker Compose

```bash
git clone https://github.com/<your-username>/distributed-cron.git
cd distributed-cron
docker compose up --build
```

This starts:

| Service | Description |
|---------|-------------|
| `postgres` | Job and execution storage |
| `redis` | Distributed locks and queue |
| `api` | REST API on port `8080` |
| `scheduler-1`, `scheduler-2` | Two scheduler instances |
| `worker-1`, `worker-2` | Two workers |
| `prometheus`, `grafana` | Metrics and dashboards |

### Run locally without Docker

Start PostgreSQL and Redis, then in separate terminals:

```bash
go run ./cmd/api
go run ./cmd/scheduler
go run ./cmd/worker
```

### Create your first job

```bash
curl -X POST http://localhost:8080/jobs \
  -H "Content-Type: application/json" \
  -d '{
    "name": "send-email",
    "schedule": "*/1 * * * *"
  }'
```

Watch it run:

```
[Scheduler] Job send-email is due
[Scheduler] Lock acquired for job send-email
[Worker]    Executing send-email
[Worker]    Job send-email completed
```

---

## Configuration

All services are configured through environment variables.

| Variable | Default | Description |
|----------|---------|-------------|
| `HTTP_PORT` | `8080` | API listen port |
| `DATABASE_URL` | none | PostgreSQL connection string |
| `REDIS_ADDR` | `localhost:6379` | Redis address |
| `SCHEDULER_ID` | hostname | Unique ID for this scheduler instance |
| `SCHEDULER_POLL_INTERVAL` | `1s` | How often the scheduler checks for due jobs |
| `LOCK_TTL` | `30s` | Lifetime of a job lock |
| `WORKER_CONCURRENCY` | `10` | Jobs a single worker runs in parallel |
| `MAX_RETRIES` | `3` | Attempts before a job is dead-lettered |
| `RETRY_BACKOFF` | `5s,15s,60s` | Delay before each retry |
| `SHUTDOWN_TIMEOUT` | `30s` | Time allowed for in-flight jobs on shutdown |
| `LOG_LEVEL` | `info` | `debug`, `info`, `warn`, `error` |

---

## API Reference

### Create a job

`POST /jobs`

```json
{
  "name": "send-email",
  "schedule": "*/5 * * * *"
}
```

**Response** `201 Created`

```json
{
  "id": "6f1c2a9e-8d4b-4c1e-9a77-2b5d3e8f1a10",
  "name": "send-email",
  "schedule": "*/5 * * * *",
  "status": "active",
  "next_run_at": "2025-01-01T10:05:00Z",
  "created_at": "2025-01-01T10:00:00Z"
}
```

### Other endpoints

| Method | Endpoint | Description |
|--------|----------|-------------|
| `GET` | `/jobs` | List all jobs |
| `GET` | `/jobs/{id}` | Get a job |
| `PUT` | `/jobs/{id}` | Update a job's name or schedule |
| `POST` | `/jobs/{id}/pause` | Pause a job |
| `POST` | `/jobs/{id}/resume` | Resume a paused job |
| `DELETE` | `/jobs/{id}` | Delete a job |
| `GET` | `/jobs/{id}/executions` | Execution history for a job |
| `GET` | `/dead-letters` | List dead-lettered executions |
| `POST` | `/dead-letters/{id}/retry` | Re-enqueue a dead-lettered execution |
| `GET` | `/healthz` | Liveness probe |
| `GET` | `/metrics` | Prometheus metrics |

### Cron syntax

```
┌───────────── minute        (0 - 59)
│ ┌─────────── hour          (0 - 23)
│ │ ┌───────── day of month  (1 - 31)
│ │ │ ┌─────── month         (1 - 12)
│ │ │ │ ┌───── day of week   (0 - 6, Sunday = 0)
│ │ │ │ │
* * * * *
```

| Expression | Meaning |
|------------|---------|
| `*/1 * * * *` | Every minute |
| `*/5 * * * *` | Every 5 minutes |
| `0 22 * * *` | Every day at 10 PM |
| `0 9 * * 1-5` | 9 AM on weekdays |

---

## Reliability Guarantees

| Failure scenario | Behavior |
|------------------|----------|
| Two schedulers see the same due job | Only the one holding the Redis lock enqueues it |
| Scheduler crashes while holding a lock | Lock expires after its TTL and another scheduler takes over |
| Application restarts | Jobs and history are reloaded from PostgreSQL; missed runs are recovered |
| Worker crashes mid-execution | The job is redelivered; the idempotency key prevents repeated side effects |
| Job fails with a transient error | Retried with exponential backoff |
| Job keeps failing | Moved to the dead-letter store with its error for inspection |
| Process receives `SIGTERM` | Stops accepting new work, drains in-flight jobs, then exits |

The system provides **at-least-once delivery combined with idempotent execution**, which together give effectively-once behavior for the work a job performs.

---

## Observability

**Logging:** structured JSON logs with job ID, execution ID, scheduler ID and attempt number on every line.

**Metrics** are exposed at `/metrics` and scraped by Prometheus:

| Metric | Description |
|--------|-------------|
| `cron_jobs_scheduled_total` | Jobs enqueued by schedulers |
| `cron_lock_acquired_total` / `cron_lock_contended_total` | Lock outcomes across schedulers |
| `cron_executions_total{status}` | Executions by result (`success`, `failed`, `retried`) |
| `cron_execution_duration_seconds` | Execution time histogram |
| `cron_dead_letter_total` | Jobs moved to the dead-letter store |
| `cron_queue_depth` | Jobs waiting in the queue |

Grafana is available at `http://localhost:3000` when running via Docker Compose.

---

## Testing

```bash
# Unit tests
go test ./...

# With the race detector
go test -race ./...

# Integration tests (requires PostgreSQL and Redis)
go test -tags=integration ./...
```

The suite covers:

- Cron expression parsing and next-run calculation
- Lock acquisition, contention and expiry
- Concurrent schedulers claiming the same job, verifying exactly one execution
- Retry timing and dead-letter transitions
- Idempotent re-delivery after simulated crashes
- Graceful shutdown behavior

---

## Deployment

### Docker Compose

```bash
docker compose up --build -d
```

Scale services as needed:

```bash
docker compose up -d --scale scheduler=3 --scale worker=5
```

### Kubernetes

Manifests are in `deploy/k8s/`.

```bash
kubectl apply -f deploy/k8s/
```

- **Schedulers** run as a Deployment with multiple replicas; locking makes this safe.
- **Workers** run as a separate Deployment and can be autoscaled on queue depth.
- **API** runs as a Deployment behind a Service.
- PostgreSQL and Redis should be managed services or run with persistent volumes.

---
