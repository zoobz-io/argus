FROM python:3.12-slim AS builder

WORKDIR /build

# Install dependencies first for layer caching.
COPY tools/classify/requirements.txt .
RUN pip install --no-cache-dir -r requirements.txt

# Generate Python proto stubs.
COPY proto/classify.proto .
RUN python -m grpc_tools.protoc \
    --proto_path=. \
    --python_out=. \
    --grpc_python_out=. \
    classify.proto

# Download the model at build time so it's baked into the image.
RUN python -c "from transformers import pipeline; pipeline('text-classification', model='protectai/deberta-v3-base-prompt-injection-v2')"

FROM python:3.12-slim

WORKDIR /app

# Copy installed packages from builder.
COPY --from=builder /usr/local/lib/python3.12/site-packages /usr/local/lib/python3.12/site-packages
COPY --from=builder /usr/local/bin /usr/local/bin

# Copy generated proto stubs.
COPY --from=builder /build/classify_pb2.py /build/classify_pb2_grpc.py ./

RUN useradd -m -u 1000 sidecar

# Copy the model cache.
COPY --from=builder /root/.cache/huggingface /home/sidecar/.cache/huggingface
RUN chown -R sidecar:sidecar /home/sidecar/.cache

# Install grpc-health-probe for container health checks.
RUN apt-get update && apt-get install -y --no-install-recommends wget && \
    wget -qO /usr/local/bin/grpc-health-probe https://github.com/grpc-ecosystem/grpc-health-probe/releases/download/v0.4.35/grpc_health_probe-linux-amd64 && \
    chmod +x /usr/local/bin/grpc-health-probe && \
    apt-get remove -y wget && apt-get autoremove -y && rm -rf /var/lib/apt/lists/*

# Copy server.
COPY tools/classify/server.py .

USER sidecar
EXPOSE 50053
CMD ["python", "server.py"]
