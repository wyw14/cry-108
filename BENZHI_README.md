# FabChem

FabChem is a local control service for wafer-fab chemical blending, compatible
manifold routing, filter preparation, nitrogen blanketing, metered dispensing,
return handling, leak isolation, and scrubber interlocks.

Run the service with Go 1.26.2:

```text
go run -mod=vendor ./cmd/fabchem -listen 127.0.0.1:21208 -data ./data
```

The service exposes health and JSON APIs under `/healthz` and `/api`.
