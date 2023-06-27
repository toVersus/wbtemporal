# Vertex AI Workbench provisioner using Temporal

## Prerequisites

- [OrbStack](https://docs.orbstack.dev/install)

  ```sh
  brew install orbstack
  ```

- Docker CE / Docker Compose / Docker BuildKit

  ```sh
  brew install --cask docker
  ```

- [Temporal](https://docs.temporal.io/cli/#installation)

  ```sh
  brew install temporal
  ```

- Google Cloud

  ```sh
  gcloud compute networks create sample --subnet-mode=custom
  gcloud compute networks subnets create sample-0 --network=sample --range=10.10.0.0/24 --region=asia-northeast1
  ```

## Usage

Temporal をローカル環境で起動

```sh
temporal server start-dev
```

http://localhost:8233 で Web UI にアクセス可能

Temporal の Workflow を実装した worker を起動

```sh
go run main.go worker run
```

Temporal の Starter を起動して Workflow をトリガーします。

Workbench Instance の作成

```sh
GCP_PROJECT_ID=
GOOGLE_ACCOUNT_EMAIL=

go run main.go starter create \
  --name sample \
  --project-id ${GCP_PROJECT_ID} \
  --email ${GOOGLE_ACCOUNT_EMAIL} \
  --network sample \
  --subnet sample-0 \
  --wait
```

Workbench Instance の停止

```sh
GCP_PROJECT_ID=

go run main.go starter stop \
  --name sample \
  --project-id ${GCP_PROJECT_ID} \
  --wait
```

Workbench Instance の起動

```sh
GCP_PROJECT_ID=

go run main.go starter start \
  --name sample \
  --project-id ${GCP_PROJECT_ID} \
  --wait
```

Workbench Instance の削除

```sh
GCP_PROJECT_ID=

go run main.go starter delete \
  --name sample \
  --project-id ${GCP_PROJECT_ID} \
  --wait
```

## Clean up

```sh
gcloud compute networks subnets delete sample-0 --region=asia-northeast1
gcloud compute networks delete sample
```
