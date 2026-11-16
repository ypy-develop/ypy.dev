# One artifact: a static binary and the Markdown it renders, in a distroless image.
#
# No runtime, no package manager, no shell.

FROM golang:1.26-alpine AS build

# libstdc++ is the C++ runtime the standalone Tailwind binary is linked against;
# Alpine ships neither it nor libgcc. It is not a toolchain.
RUN apk add --no-cache git ca-certificates libstdc++

WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .

ARG VERSION=dev
ARG COMMIT=unknown

# The views are build output and are not in the repository, so the image
# compiles them. aru is pinned, not @latest: an image built from @latest depends
# on the day it was built.
ARG ARU_VERSION=v0.76.0
RUN go install github.com/arandu-io/aru@${ARU_VERSION} && \
    "$(go env GOPATH)/bin/aru" view:build

RUN CGO_ENABLED=0 go build \
      -trimpath \
      -ldflags "-s -w -X main.version=${VERSION} -X main.commit=${COMMIT}" \
      -o /app .

FROM gcr.io/distroless/static-debian12:nonroot

COPY --from=build /app /app
# The pages are the Markdown files, read when the process starts.
COPY --from=build /src/content /content

USER nonroot:nonroot

# Production unless the deployment says otherwise.
ENV APP_ENV=prod
ENV CONTENT_DIR=/content

EXPOSE 8080

ENTRYPOINT ["/app"]
CMD ["serve"]
