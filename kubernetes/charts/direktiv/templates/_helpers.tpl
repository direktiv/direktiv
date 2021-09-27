{{/*
Expand the name of the chart.
*/}}
{{- define "direktiv.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Create a default fully qualified app name.
We truncate at 63 chars because some Kubernetes name fields are limited to this (by the DNS naming spec).
If release name contains chart name it will be used as a full name.
*/}}
{{- define "direktiv.fullname" -}}
{{- if .Values.fullnameOverride }}
{{- .Values.fullnameOverride | trunc 63 | trimSuffix "-" }}
{{- else }}
{{- $name := default .Chart.Name .Values.nameOverride }}
{{- if contains $name .Release.Name }}
{{- .Release.Name | trunc 63 | trimSuffix "-" }}
{{- else }}
{{- printf "%s-%s" .Release.Name $name | trunc 63 | trimSuffix "-" }}
{{- end }}
{{- end }}
{{- end }}

{{/*
Create chart name and version as used by the chart label.
*/}}
{{- define "direktiv.chart" -}}
{{- printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Common labels
*/}}
{{- define "direktiv.labels" -}}
helm.sh/chart: {{ include "direktiv.chart" . }}
{{ include "direktiv.selectorLabels" . }}
{{- if .Chart.AppVersion }}
app.kubernetes.io/version: {{ .Chart.AppVersion | quote }}
{{- end }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
{{- end }}

{{/*
Selector labels
*/}}
{{- define "direktiv.selectorLabels" -}}
app.kubernetes.io/name: {{ include "direktiv.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end }}

{{/*
Selector labels api
*/}}
{{- define "direktiv.selectorLabelsAPI" -}}
app.kubernetes.io/name: {{ include "direktiv.name" . }}-api
app.kubernetes.io/instance: {{ .Release.Name }}-api
{{- end }}

{{- define "direktiv.labelsAPI" -}}
helm.sh/chart: {{ include "direktiv.chart" . }}
{{ include "direktiv.selectorLabelsAPI" . }}
{{- if .Chart.AppVersion }}
app.kubernetes.io/version: {{ .Chart.AppVersion | quote }}
{{- end }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
{{- end }}


{{/*
Selector labels functions
*/}}
{{- define "direktiv.selectorLabelsFunctions" -}}
app.kubernetes.io/name: {{ include "direktiv.name" . }}-functions
app.kubernetes.io/instance: {{ .Release.Name }}-functions
{{- end }}

{{- define "direktiv.labelsFunctions" -}}
helm.sh/chart: {{ include "direktiv.chart" . }}
{{ include "direktiv.selectorLabelsFunctions" . }}
{{- if .Chart.AppVersion }}
app.kubernetes.io/version: {{ .Chart.AppVersion | quote }}
{{- end }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
{{- end }}

{{/*
Selector labels ui
*/}}
{{- define "direktiv.selectorLabelsUI" -}}
app.kubernetes.io/name: {{ include "direktiv.name" . }}-ui
app.kubernetes.io/instance: {{ .Release.Name }}-ui
{{- end }}

{{- define "direktiv.labelsUI" -}}
helm.sh/chart: {{ include "direktiv.chart" . }}
{{ include "direktiv.selectorLabelsUI" . }}
{{- if .Chart.AppVersion }}
app.kubernetes.io/version: {{ .Chart.AppVersion | quote }}
{{- end }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
{{- end }}

{{/*
Create the name of the service account to use
*/}}
{{- define "direktiv.serviceAccountName" -}}
{{- default (include "direktiv.fullname" .) .Values.serviceAccount.name }}
{{- end }}

{{/*
otlp sidecar
*/}}
{{- define "direktiv.opentelemetry" -}}
- command:
    - "/otelcol"
    - "--config=/conf/otel-agent-config.yaml"
    - "--mem-ballast-size-mib=165"
  image: otel/opentelemetry-collector-dev:latest
  name: otel-agent
  resources:
    limits:
      cpu: 500m
      memory: 500Mi
    requests:
      cpu: 100m
      memory: 100Mi
  volumeMounts:
  - name: otel-agent-config-vol
    mountPath: /conf
{{- end }}

{{- define "direktiv.opentelemetry.volume" -}}
- configMap:
    name: {{ include "direktiv.fullname" . }}-otel-agent-config
    items:
      - key: otel-agent-config
        path: otel-agent-config.yaml
  name: otel-agent-config-vol
{{- end }}
