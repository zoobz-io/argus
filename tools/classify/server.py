"""Prompt injection classification gRPC server using DeBERTa."""

import logging
import os
import signal
import sys
from concurrent import futures

import grpc
from grpc_health.v1 import health, health_pb2_grpc
from transformers import pipeline

import classify_pb2
import classify_pb2_grpc

logging.basicConfig(level=logging.INFO, format="%(asctime)s %(message)s")
log = logging.getLogger(__name__)

MODEL_NAME = os.getenv(
    "CLASSIFY_MODEL", "protectai/deberta-v3-base-prompt-injection-v2"
)
THRESHOLD = float(os.getenv("CLASSIFY_THRESHOLD", "0.5"))
PORT = os.getenv("CLASSIFY_PORT", "50053")


class ClassifyServicer(classify_pb2_grpc.ClassifyServiceServicer):
    def __init__(self):
        log.info("loading model: %s", MODEL_NAME)
        self.pipe = pipeline(
            "text-classification",
            model=MODEL_NAME,
            truncation=True,
            max_length=512,
        )
        log.info("model loaded")

    def ClassifyText(self, request, context):
        if not request.text:
            context.abort(grpc.StatusCode.INVALID_ARGUMENT, "empty text")

        result = self.pipe(request.text)[0]
        label = result["label"]  # "SAFE" or "INJECTION"
        score = result["score"]

        # The model returns the confidence of the predicted label.
        # If the label is "INJECTION", score is the injection probability.
        # If the label is "SAFE", injection probability is 1 - score.
        if label == "INJECTION":
            injection_score = score
        else:
            injection_score = 1.0 - score

        safe = injection_score < THRESHOLD

        return classify_pb2.ClassifyResponse(
            safe=safe,
            score=injection_score,
            label=label,
        )


def serve():
    server = grpc.server(futures.ThreadPoolExecutor(max_workers=4))
    classify_pb2_grpc.add_ClassifyServiceServicer_to_server(
        ClassifyServicer(), server
    )

    # Health check.
    health_servicer = health.HealthServicer()
    health_pb2_grpc.add_HealthServicer_to_server(health_servicer, server)

    server.add_insecure_port(f"[::]:{PORT}")
    server.start()
    log.info("listening on :%s", PORT)

    def shutdown(signum, frame):
        log.info("shutting down...")
        server.stop(grace=5)
        sys.exit(0)

    signal.signal(signal.SIGTERM, shutdown)
    signal.signal(signal.SIGINT, shutdown)
    server.wait_for_termination()


if __name__ == "__main__":
    serve()
