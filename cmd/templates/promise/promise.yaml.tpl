apiVersion: platform.kratix.io/v1alpha1
kind: Promise
metadata:
  name: {{ .Name }}
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
                  type: object
                  properties: {}
          served: true
          storage: true
