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

# Copy the model cache.
COPY --from=builder /root/.cache/huggingface /root/.cache/huggingface

# Copy server.
COPY tools/classify/server.py .

EXPOSE 50053
CMD ["python", "server.py"]
