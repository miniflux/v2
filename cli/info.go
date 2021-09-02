package cli

import (
	"fmt"
	"runtime"

	"miniflux.app/version"
)

func info() {
	fmt.Println("Version:", version.Version)
	fmt.Println("Commit:", version.Commit)
	fmt.Println("Build Date:", version.BuildDate)
	fmt.Println("Go Version:", runtime.Version())
	fmt.Println("Compiler:", runtime.Compiler)
	fmt.Println("Arch:", runtime.GOARCH)
	fmt.Println("OS:", runtime.GOOS)
}
