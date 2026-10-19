# syntax=docker/dockerfile:1
# meshctl image: the server, plus the binaries it serves at /dl/ and pins in
# /install.sh (amendment A14). Built by `./scripts/dev.sh image` and, for
# linux/amd64 and linux/arm64, by the release workflow; both pass VERSION
# because the build context has no git history.
#
# Bumping Go: change go.mod's toolchain line and this base image together
# (GOTOOLCHAIN=local makes a mismatch fail the build instead of downloading).

# The build stage runs natively and cross-compiles every shipped target once;
# the per-platform final stage only copies.
FROM --platform=$BUILDPLATFORM golang:1.27.2-alpine@sha256:f92b6ef800e499660581efdabdf25d9d817a9d124eaf900924f0504e7e27e12d AS build
ARG VERSION=dev
ENV CGO_ENABLED=0 GOTOOLCHAIN=local GOFLAGS=-trimpath
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
# The Windows binary carries wintun.dll (amendment A34): fetchwintun fetches
# the pinned release and checks the zip and the DLL before it is embedded.
RUN set -eu; \
    case "$VERSION" in ""|*[!A-Za-z0-9._+-]*) echo "VERSION must match [A-Za-z0-9._+-]+ (got '$VERSION')" >&2; exit 1 ;; esac; \
    mkdir -p /out/dist /out/data; \
    go run ./internal/tools/fetchwintun -arch amd64 -out internal/agent/wintun; \
    for t in linux/amd64 linux/arm64 windows/amd64; do \
      os="${t%/*}"; arch="${t#*/}"; out="/out/dist/meshctl-$os-$arch"; tags=""; \
      if [ "$os" = windows ]; then out="$out.exe"; tags=wintun; fi; \
      GOOS="$os" GOARCH="$arch" go build -tags "$tags" -ldflags "-s -w -X main.version=$VERSION" -o "$out" ./cmd/meshctl; \
    done; \
    cd /out/dist && sha256sum meshctl-* > SHA256SUMS

FROM gcr.io/distroless/static-debian13:nonroot@sha256:e2e927ec666bae08560abb3c55d0659eceabb657f56b6782ab500a9fc7f555e3
ARG TARGETARCH
COPY --from=build /out/dist/meshctl-linux-${TARGETARCH} /app/meshctl
COPY --from=build /out/dist /app/dist
COPY --from=build --chown=65532:65532 /out/data /data
ENV MESHCTL_DIST_DIR=/app/dist
USER 65532:65532
EXPOSE 8080
HEALTHCHECK --interval=30s --timeout=5s --start-period=10s --retries=3 CMD ["/app/meshctl", "healthcheck", "server"]
ENTRYPOINT ["/app/meshctl"]
CMD ["server", "--data", "/data"]
