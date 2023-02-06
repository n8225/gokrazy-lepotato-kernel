package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"

	_ "embed"

	"github.com/n8225/gokrazy-lepotato-kernel/internal/utils"
)

//go:embed config.txt
var configContents []byte

// see https://www.kernel.org/releases.json
var latest = "https://cdn.kernel.org/pub/linux/kernel/v6.x/linux-6.1.10.tar.xz"

func compile() error {
	defconfig := exec.Command("make", "ARCH=arm64", "defconfig")
	defconfig.Stdout = os.Stdout
	defconfig.Stderr = os.Stderr
	if err := defconfig.Run(); err != nil {
		return fmt.Errorf("make defconfig: %v", err)
	}

	// f, err := os.OpenFile(".config", os.O_APPEND|os.O_WRONLY, 0644)
	// if err != nil {
	// 	return err
	// }
	// defer f.Close()
	// if _, err := f.Write(configContents); err != nil {
	// 	return err
	// }
	// if err := f.Close(); err != nil {
	// 	return err
	// }

	// olddefconfig := exec.Command("make", "ARCH=arm64", "olddefconfig")
	// olddefconfig.Stdout = os.Stdout
	// olddefconfig.Stderr = os.Stderr
	// if err := olddefconfig.Run(); err != nil {
	// 	return fmt.Errorf("make olddefconfig: %v", err)
	// }

	utils.CopyFile("/usr/src/s905x.defconfig", ".config")

	make := exec.Command("make", "Image.gz", "dtbs", "-j"+strconv.Itoa(runtime.NumCPU()))
	make.Env = append(os.Environ(),
		"ARCH=arm64",
		"CROSS_COMPILE=aarch64-linux-gnu-",
		"KBUILD_BUILD_USER=gokrazy",
		"KBUILD_BUILD_HOST=docker",
		"KBUILD_BUILD_TIMESTAMP=Wed Mar  1 20:57:29 UTC 2017",
	)
	make.Stdout = os.Stdout
	make.Stderr = os.Stderr
	if err := make.Run(); err != nil {
		return fmt.Errorf("make: %v", err)
	}

	return nil
}

func main() {
	log.Printf("downloading kernel source: %s", latest)
	if err := utils.Download(latest); err != nil {
		log.Fatal(err)
	}

	log.Printf("unpacking kernel source")
	untar := exec.Command("tar", "xf", filepath.Base(latest))
	untar.Stdout = os.Stdout
	untar.Stderr = os.Stderr
	if err := untar.Run(); err != nil {
		log.Fatalf("untar: %v", err)
	}

	srcdir := strings.TrimSuffix(filepath.Base(latest), ".tar.xz")

	if err := os.Chdir(srcdir); err != nil {
		log.Fatal(err)
	}

	log.Printf("compiling kernel")
	if err := compile(); err != nil {
		log.Fatal(err)
	}

	if err := utils.CopyFile("arch/arm64/boot/Image.gz", "/tmp/buildresult/vmlinuz"); err != nil {
		log.Fatal(err)
	}

	if err := utils.CopyFile("arch/arm64/boot/dts/amlogic/meson-gxl-s905x-libretech-cc.dtb", "/tmp/buildresult/meson-gxl-s905x-libretech-cc.dtb"); err != nil {
		log.Fatal(err)
	}
}
