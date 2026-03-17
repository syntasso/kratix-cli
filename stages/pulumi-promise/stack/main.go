package main

import (
	"fmt"
	"log"

	stage "github.com/syntasso/kratix-cli/stages/pulumi-promise/internal/stage"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

const (
	stackAPIVersion = "pulumi.com/v1"
	stackKind       = "Stack"

	pulumiComponentTokenEnvVar        = "PULUMI_COMPONENT_TOKEN"
	pulumiAccessTokenSecretNameEnvVar = "PULUMI_STACK_ACCESS_TOKEN_SECRET_NAME"
	pulumiAccessTokenSecretKeyEnvVar  = "PULUMI_STACK_ACCESS_TOKEN_SECRET_KEY"
	pulumiAccessTokenEnvRefFieldName  = "PULUMI_ACCESS_TOKEN"
	pulumiAccessTokenEnvRefSecretType = "Secret"
)

type secretKeyRef struct {
	Name string
	Key  string
}

func main() {
	componentToken := getRequiredEnv(pulumiComponentTokenEnvVar)
	log.Printf("starting transformation (componentToken=%q)", componentToken)

	if err := transformInputToStackOutput(componentToken); err != nil {
		log.Printf("failed: %v", err)
		log.Fatalf("%v", err)
	}
}

func transformInputToStackOutput(componentToken string) error {
	inputFile := stage.ResolveInputFilePath()
	outputFile := stage.ResolveStackOutputFilePath()
	log.Printf("using input file %q and output file %q", inputFile, outputFile)

	request, err := stage.ReadRequestFromFile(inputFile)
	if err != nil {
		log.Printf("unable to read request from %q: %v", inputFile, err)
		return err
	}
	log.Printf("loaded request (name=%q, kind=%q, namespace=%q)", request.GetName(), request.GetKind(), request.GetNamespace())

	requestName, err := stage.RequireRequestName(request)
	if err != nil {
		log.Printf("request validation failed: %v", err)
		return err
	}

	requestNamespace := stage.RequestNamespaceWithDefault(request)

	programName := stage.BuildProgramName(requestName, requestNamespace, request.GetKind(), componentToken)
	stackResourceName := buildStackResourceName(programName)
	stackName := stackResourceName
	log.Printf("computed names (programName=%q, stackName=%q)", programName, stackName)

	output := &unstructured.Unstructured{}
	output.SetAPIVersion(stackAPIVersion)
	output.SetKind(stackKind)
	output.SetName(stackResourceName)
	output.SetNamespace(requestNamespace)
	output.SetLabels(request.GetLabels())
	output.SetAnnotations(request.GetAnnotations())

	if err := unstructured.SetNestedField(output.Object, map[string]any{
		"name": programName,
	}, "spec", "programRef"); err != nil {
		return fmt.Errorf("failed to set spec.programRef: %w", err)
	}

	if err := unstructured.SetNestedField(output.Object, stackName, "spec", "stack"); err != nil {
		return fmt.Errorf("failed to set spec.stack: %w", err)
	}

	accessTokenSecretRef, err := getOptionalAccessTokenSecretRef()
	if err != nil {
		return err
	}
	if accessTokenSecretRef != nil {
		if err := unstructured.SetNestedField(output.Object, map[string]any{
			pulumiAccessTokenEnvRefFieldName: map[string]any{
				"type": pulumiAccessTokenEnvRefSecretType,
				"secret": map[string]any{
					"name": accessTokenSecretRef.Name,
					"key":  accessTokenSecretRef.Key,
				},
			},
		}, "spec", "envRefs"); err != nil {
			return fmt.Errorf("failed to set spec.envRefs: %w", err)
		}
	}

	if err := stage.WriteOutputObject(outputFile, stackKind, output); err != nil {
		log.Printf("failed to write Stack to %q: %v", outputFile, err)
		return err
	}

	log.Printf("wrote Stack %q to %q", stackResourceName, outputFile)
	return nil
}

func buildStackResourceName(programName string) string {
	name := fmt.Sprintf("%s-stack", programName)
	if len(name) <= 63 {
		return name
	}
	return name[:63]
}

func getRequiredEnv(key string) string {
	value := stage.GetEnvWithDefault(key, "")
	if value == "" {
		log.Fatalf("missing required environment variable %s", key)
	}
	return value
}

func getOptionalAccessTokenSecretRef() (*secretKeyRef, error) {
	secretName := stage.GetEnvWithDefault(pulumiAccessTokenSecretNameEnvVar, "")
	secretKey := stage.GetEnvWithDefault(pulumiAccessTokenSecretKeyEnvVar, "")

	if secretName == "" && secretKey == "" {
		return nil, nil
	}
	if secretName == "" || secretKey == "" {
		return nil, fmt.Errorf("invalid Pulumi Stack access token secret configuration: set both %s and %s", pulumiAccessTokenSecretNameEnvVar, pulumiAccessTokenSecretKeyEnvVar)
	}

	return &secretKeyRef{
		Name: secretName,
		Key:  secretKey,
	}, nil
}
