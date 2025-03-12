package internal

import "github.com/hashicorp/go-getter"

func SetGetModuleFunc(f func(dst, src string, opts ...getter.ClientOption) error) {
	getModule = f
}

func SetMkdirTempFunc(f func(dir, pattern string) (string, error)) {
	mkdirTemp = f
}
