package main

//go:generate go tool oapi-codegen -config authentication_config.yaml authentication.yaml
//go:generate go tool oapi-codegen -config authorization_roles_config.yaml authorization_roles.yaml
//go:generate go tool oapi-codegen -config authorization_users_config.yaml authorization_users.yaml
//go:generate go tool oapi-codegen -config authorization_resources_config.yaml authorization_resources.yaml
//go:generate go tool oapi-codegen -config authorization_permissions_config.yaml authorization_permissions.yaml
//go:generate go tool oapi-codegen -config schema_config.yaml schemas.yaml

//go:generate openapi-generator generate -i authentication.yaml -g typescript-angular -o gen/typescript-angular/authentication
//go:generate openapi-generator generate -i authorization_roles.yaml -g typescript-angular -o gen/typescript-angular/authorization_roles
//go:generate openapi-generator generate -i authorization_users.yaml -g typescript-angular -o gen/typescript-angular/authorization_users
//go:generate openapi-generator generate -i authorization_resources.yaml -g typescript-angular -o gen/typescript-angular/authorization_resources
//go:generate openapi-generator generate -i authorization_permissions.yaml -g typescript-angular -o gen/typescript-angular/authorization_permissions

//go:generate openapi-generator generate -i authentication.yaml -g html2 -o docs/authentication
//go:generate openapi-generator generate -i authorization_roles.yaml -g html2 -o docs/authorization_roles
//go:generate openapi-generator generate -i authorization_users.yaml -g html2 -o docs/authorization_users
//go:generate openapi-generator generate -i authorization_resources.yaml -g html2 -o docs/authorization_resources
//go:generate openapi-generator generate -i authorization_permissions.yaml -g html2 -o docs/authorization_permissions
