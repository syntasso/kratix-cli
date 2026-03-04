package cmd

const (
	pulumiGeneratorImageRepository = "ghcr.io/syntasso/kratix-cli/pulumi-generator"
	pulumiGeneratorImageVersion    = "v0.1.1"

	pulumiProgramGeneratorContainerName = "pulumi-program-generator"
	pulumiStackGeneratorContainerName   = "pulumi-stack-generator"
	pulumiProgramGeneratorCommand       = "/pulumi-program-generator"
	pulumiStackGeneratorCommand         = "/pulumi-stack-generator"
)

func pulumiGeneratorImage() string {
	return pulumiGeneratorImageRepository + ":" + pulumiGeneratorImageVersion
}
