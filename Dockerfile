FROM golang:1.25-alpine AS builder

WORKDIR /app

COPY go.sum go.mod ./
RUN --mount=type=cache,target=/go/pkg/mod/ \
    go mod download -x

COPY . .

RUN --mount=type=cache,target=/go/pkg/mod/ \
    --mount=type=bind,target=. \
    go install -v ./cmd/codeactions/main.go \
    && go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest

FROM alpine:3.22.0

# Instalar todas as dependências necessárias incluindo bash e postgresql-client
RUN apk add --no-cache \
    bash \
    python3 python3-dev py3-pip \
    ffmpeg \
    postgresql-dev postgresql-client libpq libpq-dev \
    build-base \
    go \
    mongodb-tools

RUN pip install psycopg2 psycopg2-binary pymongo --break-system-packages

COPY --from=builder /app/requirements.txt .
RUN pip install --break-system-packages --no-cache-dir -r requirements.txt

ENV APP_USER=app \
    APP_GROUP=app \
    USER_ID=1999 \
    GROUP_ID=1999

RUN addgroup --system --gid ${GROUP_ID} ${APP_GROUP} \
    && adduser --system --disabled-password --home /home/${APP_USER} \
    --uid ${USER_ID} --ingroup ${APP_GROUP} ${APP_USER}

COPY --from=builder --chown=${APP_USER}:${APP_GROUP} /go/bin /app

WORKDIR /app

COPY --chown=${APP_USER}:${APP_GROUP} ./engines ./engines

# Copiar migrations e scripts para permitir execução manual
COPY --chown=${APP_USER}:${APP_GROUP} ./migrations ./migrations
COPY --chown=${APP_USER}:${APP_GROUP} ./scripts ./scripts

USER ${APP_USER}:${APP_GROUP}

RUN chmod -R u+w /home/app

EXPOSE 8080
CMD ["./main"]