{{- /* Usage: {{ include "helm_lib_prometheus_rules_recursion" (list . <namespace> <root dir> [current dir]) }} */ -}}
{{- /* returns all the prometheus rules from <root dir>/ */ -}}
{{- /* current dir is optional — used for recursion but you can use it for partially generating rules */ -}}
{{- define "helm_lib_prometheus_rules_recursion" -}}
  {{- $context := index . 0 }}
  {{- $namespace := index . 1 }}
  {{- $rootDir := index . 2 }}
  {{- $currentDir := "" }}
  {{- if gt (len .) 3 }} {{- $currentDir = index . 3 }} {{- else }} {{- $currentDir = $rootDir }} {{- end }}
  {{- $currentDirIndex := (sub ($currentDir | splitList "/" | len) 1) }}
  {{- $rootDirIndex := (sub ($rootDir | splitList "/" | len) 1) }}
  {{- $folderNamesIndex := (add1 $rootDirIndex) }}

  {{- range $path, $_ := $context.Files.Glob (print $currentDir "/*.{yaml,tpl}") }}
    {{- $fileName := ($path | splitList "/" | last ) }}
    {{- $definition := "" }}
    {{- if eq ($path | splitList "." | last) "tpl" -}}
      {{- $definition = tpl ($context.Files.Get $path) $context }}
    {{- else }}
      {{- $definition = $context.Files.Get $path }}
    {{- end }}

    {{- $definition = $definition | replace "__SCRAPE_INTERVAL__" (printf "%ds" ($context.Values.global.discovery.prometheusScrapeInterval | default 30)) | replace "__SCRAPE_INTERVAL_X_2__" (printf "%ds" (mul ($context.Values.global.discovery.prometheusScrapeInterval | default 30) 2)) | replace "__SCRAPE_INTERVAL_X_3__" (printf "%ds" (mul ($context.Values.global.discovery.prometheusScrapeInterval | default 30) 3)) | replace "__SCRAPE_INTERVAL_X_4__" (printf "%ds" (mul ($context.Values.global.discovery.prometheusScrapeInterval | default 30) 4)) }}

    {{- $resourceName := (regexReplaceAllLiteral "\\.(yaml|tpl)$" $path "") }}
    {{- $resourceName = ($resourceName | replace " " "-" | replace "." "-" | replace "_" "-") }}
    {{- $resourceName = (slice ($resourceName | splitList "/") $folderNamesIndex | join "-") }}
    {{- $resourceName = (printf "%s-%s" $context.Chart.Name $resourceName) }}
---
apiVersion: monitoring.coreos.com/v1
kind: PrometheusRule
metadata:
  name: {{ $resourceName }}
  namespace: {{ $namespace }}
  {{- include "helm_lib_module_labels" (list $context (dict "app" "prometheus" "prometheus" "main" "component" "rules")) | nindent 2 }}
spec:
  groups:
    {{- $definition | nindent 4 }}
  {{- end }}

  {{- $subDirs := list }}
  {{- range $path, $_ := ($context.Files.Glob (print $currentDir "/**.{yaml,tpl}")) }}
    {{- $pathSlice := ($path | splitList "/") }}
    {{- $subDirs = append $subDirs (slice $pathSlice 0 (add $currentDirIndex 2) | join "/") }}
  {{- end }}

  {{- range $subDir := ($subDirs | uniq) }}
{{ include "helm_lib_prometheus_rules_recursion" (list $context $namespace $rootDir $subDir) }}
  {{- end }}
{{- end }}


{{- /* Usage: {{ include "helm_lib_prometheus_rules" (list . <namespace>) }} */ -}}
{{- /* returns all the prometheus rules from monitoring/prometheus-rules/ */ -}}
{{- define "helm_lib_prometheus_rules" -}}
  {{- $context := index . 0 }}
  {{- $namespace := index . 1 }}
  {{- if ( $context.Values.global.enabledModules | has "operator-prometheus-crd" ) }}
{{- include "helm_lib_prometheus_rules_recursion" (list $context $namespace "monitoring/prometheus-rules") }}
  {{- end }}
{{- end }}
