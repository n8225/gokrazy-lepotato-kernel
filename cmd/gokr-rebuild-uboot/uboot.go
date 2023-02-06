package main

import (
	"flag"
	"fmt"
	"log"
	"path/filepath"

	"github.com/n8225/gokrazy-lepotato-kernel/internal/utils"
)

var buildFiles = []string{"boot.cmd"}

func main() {
	pkgPath, err := utils.FindPkgDir()
	if err != nil {
		log.Fatalf("Unable to find pkg dir: %s", err)
	}
	fmt.Printf("%s\n", pkgPath)

	var overwriteContainerExecutable = flag.String("overwrite_container_executable",
		"",
		"E.g. docker or podman to overwrite the automatically detected container executable")
	flag.Parse()
	executable, err := utils.GetContainerExecutable()
	if err != nil {
		log.Fatal(err)
	}
	if *overwriteContainerExecutable != "" {
		executable = *overwriteContainerExecutable
	}
	execName := filepath.Base(executable)

	tmp := utils.CreateTmpDir("uboot")

	utils.BuildGoBinary("uboot", tmp, pkgPath)

	utils.CreateContainer("uboot", pkgPath, tmp, execName, buildFiles)

	utils.RunCompile("uboot", execName, tmp)

	if err := utils.CopyFile(filepath.Join(tmp, "u-boot.bin"), filepath.Join(pkgPath, "u-boot")); err != nil {
		log.Fatal(err)
	}

	if err := utils.CopyFile(filepath.Join(tmp, "boot.scr"), filepath.Join(pkgPath, "boot.scr")); err != nil {
		log.Fatal(err)
	}
}
