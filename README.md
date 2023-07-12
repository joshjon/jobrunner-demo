# 🏃‍ Job Runner Demo

> **Note**
> This project exists for demo purposes only and would not be suitable for production use.

A job worker pool service that provides a gRPC API to run arbitrary processes on Linux or Darwin hosts.

## 🚀 Running

1. Run the gRPC server.

   ```shell
   make run
   ```

2. Make requests to `localhost:50051` using a gRPC client.

### Example Usage

Queue a job e.g. `ping google.com`.

```shell
grpcurl -plaintext -d '{"command":{"cmd": "ping", "args": ["google.com"]}}' localhost:50051 rpc.service.v1.Service/QueueJob
```

Subscribe to job logs.

```shell
grpcurl -plaintext -d '{"job_id": "<JOB_ID_HERE>"}' localhost:50051 rpc.service.v1.Service/Subscribe
```

Stop a running job.

```shell
 grpcurl -plaintext -d '{"job_id": "<JOB_ID_HERE>"}' localhost:50051 rpc.service.v1.Service/StopJob
```
