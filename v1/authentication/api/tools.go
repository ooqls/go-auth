package authenticationapi

//go:generate go tool oapi-codegen -config openapi_config.yaml docs/openapi.yaml
//go:generate npx --yes @redocly/cli@latest bundle docs/openapi.yaml -o docs/openapi.bundled.yaml
//go:generate openapi-generator-cli generate -i docs/openapi.bundled.yaml -g typescript-fetch  -o ../../../clients/authentication-typescript-fetch/
