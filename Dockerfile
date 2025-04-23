FROM golang:1.24.2-bookworm AS builder

WORKDIR /app

COPY go.sum go.mod ./
RUN --mount=type=cache,target=/go/pkg/mod/ \
    go mod download -x

COPY . .

RUN --mount=type=cache,target=/go/pkg/mod/ \
    --mount=type=bind,target=. \
    go install -v ./cmd/codeactions/main.go

FROM debian:bookworm-slim

RUN apt-get update && apt-get install -y --no-install-recommends \
    python3 \
    python3-dev \
    python3-pip \
    ffmpeg \
    libpq-dev \
    build-essential \
    && rm -rf /var/lib/apt/lists/*

RUN ln -s /usr/bin/python3 /usr/bin/python

COPY --from=builder /app/requirements.txt .

RUN pip3 install --break-system-packages --no-cache-dir -r requirements.txt

ENV APP_USER=app \
    APP_GROUP=app \
    USER_ID=1999 \
    GROUP_ID=1999

RUN addgroup --system --gid ${GROUP_ID} ${APP_GROUP} \
    && adduser --system --disabled-password --home /home/${APP_USER} \
    --uid ${USER_ID} --ingroup ${APP_GROUP} ${APP_USER}

COPY --from=builder --chown=${APP_USER}:${APP_GROUP} /go/bin /app

WORKDIR /app

COPY ./engines ./engines

USER ${APP_USER}:${APP_GROUP}

RUN chmod -R u+w /home/app

EXPOSE 8080
CMD ["./main"]
