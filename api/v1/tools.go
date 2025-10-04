package v1

//go:generate go tool oapi-codegen -config authentication/openapi_config.yaml authentication/docs/openapi.yaml
//go:generate go tool oapi-codegen -config roles/openapi_config.yaml roles/docs/openapi.yaml
//go:generate go tool oapi-codegen -config users/openapi_config.yaml users/docs/openapi.yaml
//go:generate go tool oapi-codegen -config resources/openapi_config.yaml resources/docs/openapi.yaml
//go:generate go tool oapi-codegen -config permissions/openapi_config.yaml permissions/docs/openapi.yaml
//go:generate go tool oapi-codegen -config schema_config.yaml schemas.yaml

//go:generate openapi-generator generate -i authentication/docs/openapi.yaml -g typescript-angular -o gen/typescript-angular/authentication
//go:generate openapi-generator generate -i roles/docs/openapi.yaml -g typescript-angular -o gen/typescript-angular/roles
//go:generate openapi-generator generate -i users/docs/openapi.yaml -g typescript-angular -o gen/typescript-angular/users
//go:generate openapi-generator generate -i resources/docs/openapi.yaml -g typescript-angular -o gen/typescript-angular/resources
//go:generate openapi-generator generate -i permissions/docs/openapi.yaml -g typescript-angular -o gen/typescript-angular/permissions

//go:generate openapi-generator generate -i authentication/docs/openapi.yaml -g html2 -o authentication/docs
//go:generate openapi-generator generate -i roles/docs/openapi.yaml -g html2 -o roles/docs
//go:generate openapi-generator generate -i users/docs/openapi.yaml -g html2 -o users/docs
//go:generate openapi-generator generate -i resources/docs/openapi.yaml -g html2 -o resources/docs
//go:generate openapi-generator generate -i permissions/docs/openapi.yaml -g html2 -o permissions/docs
