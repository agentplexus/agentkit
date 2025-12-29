{{- /*
Generic agent deployment template for agentkit-based applications.
Usage: Include this template with agent name and configuration.
*/ -}}

{{- define "agentkit.deployment" -}}
{{- $agent := .agent -}}
{{- $name := .name -}}
{{- $values := .values -}}
{{- if $agent.Enabled }}
apiVersion: apps/v1
kind: Deployment
metadata:
  name: {{ include "agentkit.fullname" $values }}-{{ $name }}
  namespace: {{ include "agentkit.namespace" $values }}
  labels:
    {{- include "agentkit.agentLabels" (dict "context" $values "agent" $name) | nindent 4 }}
spec:
  {{- if not $agent.Autoscaling.Enabled }}
  replicas: {{ $agent.ReplicaCount | default 1 }}
  {{- end }}
  selector:
    matchLabels:
      {{- include "agentkit.agentSelectorLabels" (dict "context" $values "agent" $name) | nindent 6 }}
  template:
    metadata:
      labels:
        {{- include "agentkit.agentLabels" (dict "context" $values "agent" $name) | nindent 8 }}
    spec:
      {{- with $values.Global.ImagePullSecrets }}
      imagePullSecrets:
        {{- toYaml . | nindent 8 }}
      {{- end }}
      {{- if $values.ServiceAccount.Create }}
      serviceAccountName: {{ include "agentkit.serviceAccountName" $values }}
      {{- end }}
      securityContext:
        {{- toYaml $values.PodSecurityContext | nindent 8 }}
      containers:
        - name: {{ $name }}
          image: {{ include "agentkit.image" (dict "global" $values.Global "agent" $agent) }}
          imagePullPolicy: {{ $values.Global.Image.PullPolicy }}
          ports:
            - name: http
              containerPort: {{ $agent.Service.Port }}
              protocol: TCP
            {{- if $agent.Service.A2APort }}
            - name: a2a
              containerPort: {{ $agent.Service.A2APort }}
              protocol: TCP
            {{- end }}
          env:
            - name: PORT
              value: {{ $agent.Service.Port | quote }}
            - name: LLM_PROVIDER
              valueFrom:
                configMapKeyRef:
                  name: {{ include "agentkit.fullname" $values }}-config
                  key: LLM_PROVIDER
            {{- range $agent.Env }}
            - name: {{ .Name }}
              {{- if .ValueFrom }}
              valueFrom:
                {{- toYaml .ValueFrom | nindent 16 }}
              {{- else }}
              value: {{ .Value | quote }}
              {{- end }}
            {{- end }}
          envFrom:
            - configMapRef:
                name: {{ include "agentkit.fullname" $values }}-config
            {{- if $values.Secrets.Create }}
            - secretRef:
                name: {{ include "agentkit.fullname" $values }}-secrets
            {{- end }}
          livenessProbe:
            httpGet:
              path: /health
              port: http
            initialDelaySeconds: 10
            periodSeconds: 30
          readinessProbe:
            httpGet:
              path: /health
              port: http
            initialDelaySeconds: 5
            periodSeconds: 10
          resources:
            {{- toYaml $agent.Resources | nindent 12 }}
          securityContext:
            {{- toYaml $values.SecurityContext | nindent 12 }}
      {{- with $agent.NodeSelector }}
      nodeSelector:
        {{- toYaml . | nindent 8 }}
      {{- end }}
      {{- with $agent.Affinity }}
      affinity:
        {{- toYaml . | nindent 8 }}
      {{- end }}
      {{- with $agent.Tolerations }}
      tolerations:
        {{- toYaml . | nindent 8 }}
      {{- end }}
{{- end }}
{{- end -}}
