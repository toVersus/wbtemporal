# Vertex AI Workbench provisioner using Temporal

## Prerequisites

[OrbStack](https://docs.orbstack.dev/install) をインストール

```sh
brew install orbstack
```

Docker CE / Docker Compose / Docker BuildKit をインストール

```sh
brew install --cask docker
```

[Kind](https://kind.sigs.k8s.io/docs/user/quick-start/#installation) をインストール

```sh
brew install kind
```

[Helm](https://helm.sh/docs/intro/install/) をインストール

```sh
brew install helm
```

[Helmfile](https://helmfile.readthedocs.io/en/latest/#installation) をインストール

```sh
brew install helmfile
```

[Temporal](https://docs.temporal.io/cli/#installation) をインストール

```sh
brew install temporal
```

Google Cloud Vertex AI Workbench Instance を起動するネットワークを作成

```sh
gcloud compute networks create sample --subnet-mode=custom
gcloud compute networks subnets create sample-0 --network=sample --range=10.10.0.0/24 --region=asia-northeast1
```

ローカル環境に JupyterHub をセットアップ

```sh
kind create cluster

# Install metallb
kubectl apply -f https://raw.githubusercontent.com/metallb/metallb/v0.13.10/config/manifests/metallb-native.yaml
kubectl wait --namespace metallb-system --for=condition=ready pod --selector=app=metallb --timeout=90s

# Check Docker Network IPAM config
docker network inspect -f '{{.IPAM.Config}}' kind

# Configure IP address pool
#
# e.g.) .spec.addresses => 198.19.195.240-198.19.195.255
# ❯ docker network inspect -f '{{.IPAM.Config}}' kind
# [{198.19.194.0/23  198.19.194.1 map[]} {fc00:f853:ccd:e793::/64   map[]}]
cat << EOF | kubectl apply -f -
apiVersion: metallb.io/v1beta1
kind: IPAddressPool
metadata:
  name: example
  namespace: metallb-system
spec:
  addresses:
  - 198.19.195.240-198.19.195.255
---
apiVersion: metallb.io/v1beta1
kind: L2Advertisement
metadata:
  name: empty
  namespace: metallb-system
EOF
# Replacing IP address specified in install/helm/otel-demo/examples/values.local.yaml

helmfile sync --environment local .
```

## Usage

### Vertex AI Workbench

Temporal をローカル環境で起動

```sh
temporal server start-dev
```

http://localhost:8233 で Web UI にアクセス可能

Temporal の Workflow を実装した worker を起動

```sh
go run main.go worker workbench run
```

Temporal の Starter を起動して Workflow をトリガーします。

Workbench Instance の作成

```sh
# Workbench を作成する Google Cloud プロジェクト
GCP_PROJECT_ID=
# Workbench へのアクセスを許可する Google アカウントのメールアドレス
GOOGLE_ACCOUNT_EMAIL=

go run main.go starter workbench create \
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

go run main.go starter workbench stop \
  --name sample \
  --project-id ${GCP_PROJECT_ID} \
  --wait
```

Workbench Instance の起動

```sh
GCP_PROJECT_ID=

go run main.go starter workbench start \
  --name sample \
  --project-id ${GCP_PROJECT_ID} \
  --wait
```

Workbench Instance の削除

```sh
GCP_PROJECT_ID=

go run main.go starter workbench delete \
  --name sample \
  --project-id ${GCP_PROJECT_ID} \
  --wait
```

### JupyterHub

Temporal をローカル環境で起動

```sh
temporal server start-dev
```

http://localhost:8233 で Web UI にアクセス可能

Temporal の Workflow を実装した worker を起動

```sh
# JupyterHub の API が公開されている URL を指定
# /hub/api のパスは不要
# e.g.) http://198.19.195.240
JUPYTERHUB_BASE_URL=
# JupyterHub で生成した API トークンを指定
# install/jupyterhub/values.local.yaml の hub.services.wbtemporal.apiToken で
# インストール時に API トークンを発行しているのでそれを指定
JUPYTERHUB_API_TOKEN=

go run main.go worker jupyterhub run \
  --executor-name jupyterhub \
  --base-url ${JUPYTERHUB_BASE_URL} \
  --token ${JUPYTERHUB_API_TOKEN}
```

Temporal の Starter を起動して Workflow をトリガーします。

JupyterHub のユーザーサーバの作成

```sh
go run main.go starter jupyterhub create \
  --user sample \
  --server sample \
  --wait
```

JupyterHub のユーザーサーバの削除 (停止)

```sh
go run main.go starter jupyterhub delete \
  --user sample \
  --server sample \
  --wait
```

## Clean up

```sh
gcloud compute networks subnets delete sample-0 --region=asia-northeast1
gcloud compute networks delete sample
```

## Development

OpenAPI スキーマから JupyterHub クライアントを生成

```sh
make generate
```
