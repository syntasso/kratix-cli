package internal

func SetMkdirTempFunc(f func(dir, pattern string) (string, error)) {
	mkdirTemp = f
}

func SetTerraformInitFunc(f func(dir string) error) {
	terraformInit = f
}

func RunTerraformInit(dir string) error {
	return runTerraformInit(dir)
}
