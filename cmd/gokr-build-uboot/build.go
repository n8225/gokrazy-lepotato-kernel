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

	"github.com/n8225/gokrazy-lepotato-kernel/internal/utils"
)

const ubootRev = "a209c3e6b48cf042d0220245a2d1636f74389c17"
const ubootTS = 1675438245

var latest = "https://github.com/u-boot/u-boot/archive/" + ubootRev + ".zip"
var GxlimgUrl = "https://github.com/repk/gxlimg/archive/refs/heads/master.zip"

func compile() error {
	defconfig := exec.Command("make", "ARCH=arm64", "libretech-cc_defconfig")
	defconfig.Stdout = os.Stdout
	defconfig.Stderr = os.Stderr
	if err := defconfig.Run(); err != nil {
		return fmt.Errorf("make defconfig: %v", err)
	}

	make := exec.Command("make", "u-boot.bin", "-j"+strconv.Itoa(runtime.NumCPU()))
	make.Env = append(os.Environ(),
		"ARCH=arm64",
		"CROSS_COMPILE=aarch64-linux-gnu-",
		"SOURCE_DATE_EPOCH="+strconv.Itoa(ubootTS),
	)
	make.Stdout = os.Stdout
	make.Stderr = os.Stderr
	if err := make.Run(); err != nil {
		return fmt.Errorf("make: %v", err)
	}

	return nil
}

func generateBootScr(bootCmdPath string) error {
	mkimage := exec.Command("./tools/mkimage", "-A", "arm64", "-O", "linux", "-T", "script", "-C", "none", "-a", "0", "-e", "0", "-n", "Gokrazy Boot Script", "-d", bootCmdPath, "boot.scr")
	mkimage.Env = append(os.Environ(),
		"ARCH=arm64",
		"CROSS_COMPILE=aarch64-linux-gnu-",
		"SOURCE_DATE_EPOCH=1600000000",
	)
	mkimage.Stdout = os.Stdout
	mkimage.Stderr = os.Stderr
	if err := mkimage.Run(); err != nil {
		return fmt.Errorf("mkimage: %v", err)
	}

	return nil
}

func compileGxl() error {
	makeGxl := exec.Command("make")
	makeGxl.Stdout = os.Stdout
	makeGxl.Stderr = os.Stderr
	if err := makeGxl.Run(); err != nil {
		return fmt.Errorf("gxlimg:: make: %v", err)
	}

	genImage := exec.Command("make", "image", "UBOOT=../u-boot-dtb.bin")
	genImage.Stdout = os.Stdout
	genImage.Stderr = os.Stderr
	if err := genImage.Run(); err != nil {
		return fmt.Errorf("gxlimg:: make image: %v", err)
	}
	return nil
}

func main() {

	if err := utils.Download(latest); err != nil {
		log.Fatal(err)
	}

	utils.Unzip(latest)

	srcdir := "u-boot-" + strings.TrimSuffix(filepath.Base(latest), ".zip")

	var bootCmdPath string
	if p, err := filepath.Abs("boot.cmd"); err != nil {
		log.Fatal(err)
	} else {
		bootCmdPath = p
		log.Print(bootCmdPath)
	}

	if err := os.Chdir(srcdir); err != nil {
		log.Fatal(err)
	}

	log.Printf("compiling uboot")
	if err := compile(); err != nil {
		log.Fatal(err)
	}

	log.Printf("generating boot.scr")
	if err := generateBootScr(bootCmdPath); err != nil {
		log.Fatal(err)
	}

	if err := utils.CopyFile("boot.scr", "/tmp/buildresult/boot.scr"); err != nil {
		log.Fatal(err)
	}

	if err := utils.Download(GxlimgUrl); err != nil {
		log.Fatal(err)
	}

	utils.Unzip(GxlimgUrl)

	gxlimgsrcdir := "gxlimg-" + strings.TrimSuffix(filepath.Base(GxlimgUrl), ".zip")

	if err := os.Chdir(gxlimgsrcdir); err != nil {
		log.Fatal(err)
	}

	log.Printf("compiling Gxlimg")
	if err := compileGxl(); err != nil {
		log.Fatal(err)
	}

	if err := utils.CopyFile("build/gxl-boot.bin", "/tmp/buildresult/u-boot.bin"); err != nil {
		log.Fatal(err)
	}

}
