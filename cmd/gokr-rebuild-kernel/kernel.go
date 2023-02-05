package main

import (
	"flag"
	"fmt"
	"log"
	"path/filepath"

	"github.com/n8225/gokrazy-lepotato-kernel/internal/utils"
)

var buildFiles = []string{"config.txt", "s905x.defconfig"}

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

	tmp := utils.CreateTmpDir("kernel")

	utils.BuildGoBinary("kernel", tmp, pkgPath)

	utils.CreateContainer("kernel", pkgPath, tmp, execName, buildFiles)

	utils.RunCompile("kernel", execName, tmp)

	if err := utils.CopyFile(filepath.Join(tmp, "vmlinuz"), filepath.Join(pkgPath, "vmlinuz")); err != nil {
		log.Fatal(err)
	}

	if err := utils.CopyFile(filepath.Join(tmp, "meson-gxl-s905x-libretech-cc.dtb"), filepath.Join(pkgPath, "meson-gxl-s905x-libretech-cc.dtb")); err != nil {
		log.Fatal(err)
	}
}
