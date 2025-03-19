apiVersion: platform.kratix.io/v1alpha1
kind: Promise
metadata:
  name: {{ .Name }}
  labels:
    kratix.io/promise-version: v0.0.1
spec:
  api:
    apiVersion: apiextensions.k8s.io/v1
    kind: CustomResourceDefinition
    metadata:
      name: {{ .Plural }}.{{ .Group }}
    spec:
      group: {{ .Group }}
      names:
        kind: {{ .Kind }}
        plural: {{ .Plural }}
        singular: {{ .Singular }}
      scope: Namespaced
      versions:
        - name: {{ .Version }}
          schema:
            openAPIV3Schema:
              type: object
              properties:
                spec:
{{ .CRDSchema | indent 18 }}
          served: true
          storage: true
{{- if .DestinationSelectors }}
  destinationSelectors:
{{ .DestinationSelectors | indent 4 }}
{{- end }}
  workflows:
    promise:
      configure:
{{ .PromiseConfigure | indent 8 }}
    resource:
      configure:
{{ .ResourceConfigure | indent 8 }}
