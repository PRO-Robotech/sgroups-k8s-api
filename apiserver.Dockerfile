FROM golang:1.25 AS builder
WORKDIR /workspace
COPY sgroups-proto/ sgroups-proto/
COPY sgroups-k8s-api/ sgroups-k8s-api/
WORKDIR /workspace/sgroups-k8s-api
RUN CGO_ENABLED=0 go build -o /sgroups-k8s-apiserver ./cmd/sgroups-k8s-apiserver

FROM gcr.io/distroless/static:nonroot
COPY --from=builder /sgroups-k8s-apiserver /sgroups-k8s-apiserver
USER 65534:65534
ENTRYPOINT ["/sgroups-k8s-apiserver"]
