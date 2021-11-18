setenv root "/dev/mmcblk0p2"
fatload mmc 2:1 0x40008000 vmlinuz

fatload mmc 2:1 0x42000000 cmdline.txt

env import -t 0x42000000 ${filesize}

fatload mmc 2:1 0x44000000 exynos5422-odroidhc1.dtb

setenv bootargs "console=ttySAC2,115200n8 consoleblank=0 loglevel=7 root=${root} rootwait panic=10 oops=panic init=/gokrazy/init"

fdt addr 0x44000000

bootz 0x40008000 - 0x44000000
