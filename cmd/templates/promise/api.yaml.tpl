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
{{ .CRDSchema | indent 14 }}
      served: true
      storage: true
