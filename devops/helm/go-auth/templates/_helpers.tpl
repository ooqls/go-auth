{{/*
Expand the name of the chart.
*/}}
{{- define "go-auth.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Create a default fully qualified app name.
*/}}
{{- define "go-auth.fullname" -}}
{{- if .Values.fullnameOverride }}
{{- .Values.fullnameOverride | trunc 63 | trimSuffix "-" }}
{{- else }}
{{- $name := default .Chart.Name .Values.nameOverride }}
{{- printf "%s-%s" .Release.Name $name | trunc 63 | trimSuffix "-" }}
{{- end }}
{{- end }}

{{/*
Common labels
*/}}
{{- define "go-auth.labels" -}}
helm.sh/chart: {{ .Chart.Name }}-{{ .Chart.Version }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end }}

{{/*
Selector labels for a given API component
*/}}
{{- define "go-auth.selectorLabels" -}}
app.kubernetes.io/name: {{ include "go-auth.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
app.kubernetes.io/component: {{ .component }}
{{- end }}

{{/*
Resolved postgres host — use override if set, otherwise point at subchart service
*/}}
{{- define "go-auth.postgresHost" -}}
{{- if .Values.postgres.host }}
{{- .Values.postgres.host }}
{{- else }}
{{- printf "%s-postgresql" .Release.Name }}
{{- end }}
{{- end }}

{{/*
Resolved valkey host
*/}}
{{- define "go-auth.valkeyHost" -}}
{{- if .Values.valkey.host }}
{{- .Values.valkey.host }}
{{- else }}
{{- printf "%s-valkey-primary" .Release.Name }}
{{- end }}
{{- end }}

{{/*
TLS secret name — use existingSecret name if mode is existingSecret, else chart-managed name
*/}}
{{- define "go-auth.tlsSecretName" -}}
{{- if eq .Values.tls.mode "existingSecret" }}
{{- .Values.tls.secretName }}
{{- else }}
{{- printf "%s-tls" (include "go-auth.fullname" .) }}
{{- end }}
{{- end }}

{{/*
Standard volume mounts shared by all API containers
*/}}
{{- define "go-auth.commonVolumeMounts" -}}
- name: app-config
  mountPath: /config/config.yaml
  subPath: config.yaml
- name: registry-config
  mountPath: /config/registry/registry.yaml
  subPath: registry.yaml
- name: db-passwords
  mountPath: /run/secrets/db
  readOnly: true
- name: jwt-secret
  mountPath: /config/jwt
  readOnly: true
{{- if .Values.tls.enabled }}
- name: tls-certs
  mountPath: /config/tls
  readOnly: true
{{- end }}
{{- end }}

{{/*
Standard volumes shared by all API pods
*/}}
{{- define "go-auth.commonVolumes" -}}
- name: app-config
  configMap:
    name: {{ include "go-auth.fullname" . }}-app-config
- name: registry-config
  configMap:
    name: {{ include "go-auth.fullname" . }}-registry
- name: db-passwords
  secret:
    secretName: {{ include "go-auth.fullname" . }}-db-passwords
- name: jwt-secret
  secret:
    secretName: {{ include "go-auth.fullname" . }}-jwt
{{- if .Values.tls.enabled }}
- name: tls-certs
  secret:
    secretName: {{ include "go-auth.tlsSecretName" . }}
{{- end }}
{{- end }}
