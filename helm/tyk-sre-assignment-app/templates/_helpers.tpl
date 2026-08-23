{{/*
Expand the name of the chart.
*/}}
{{- define "tyk-are-assignment-app.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Create the namespace name.
*/}}
{{- define "tyk-are-assignment-app.namespace" -}}
{{- .Values.namespace.name | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Create a fully qualified app name.
*/}}
{{- define "tyk-are-assignment-app.fullname" -}}
{{- if .Values.fullnameOverride }}
{{- .Values.fullnameOverride | trunc 63 | trimSuffix "-" }}
{{- else }}
{{- include "tyk-are-assignment-app.name" . }}
{{- end }}
{{- end }}

{{/*
Create the ServiceAccount name.
*/}}
{{- define "tyk-are-assignment-app.serviceAccountName" -}}
{{- if .Values.serviceAccount.create }}
{{- default (include "tyk-are-assignment-app.fullname" .) .Values.serviceAccount.name }}
{{- else }}
{{- default "default" .Values.serviceAccount.name }}
{{- end }}
{{- end }}

{{/*
Common labels.
*/}}
{{- define "tyk-are-assignment-app.labels" -}}
app.kubernetes.io/name: {{ include "tyk-are-assignment-app.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
helm.sh/chart: {{ .Chart.Name }}-{{ .Chart.Version | replace "+" "_" }}
{{- end }}

{{/*
Selector labels.
*/}}
{{- define "tyk-are-assignment-app.selectorLabels" -}}
app.kubernetes.io/name: {{ include "tyk-are-assignment-app.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end }}