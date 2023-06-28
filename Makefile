.PHONY: prepare
prepare:
	go install github.com/deepmap/oapi-codegen/cmd/oapi-codegen@latest

.PHONY: codegen
codegen: prepare
	oapi-codegen -o pkg/jupyterhub/client.go -package jupyterhub api/openapi-spec/jupyterhub.yaml
