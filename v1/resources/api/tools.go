package resourcesapi

//go:generate go tool oapi-codegen -config docs/openapi_config.yaml docs/openapi.yaml
//go:generate npx --yes @redocly/cli@latest bundle docs/openapi.yaml -o docs/openapi.bundled.yaml
//go:generate openapi-generator-cli generate -i docs/openapi.bundled.yaml -g typescript-fetch  -o ../../../clients/resources-typescript-fetch/
// go:generate go-swagger-merger
