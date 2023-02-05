setenv bootargs "console=ttyAML0,115200n8 root=/dev/mmcblk0p2 rootwait panic=10 oops=panic init=/gokrazy/init"

fatload mmc 0:1 ${kernel_addr_r} vmlinuz

fatload mmc 0:1 ${fdt_addr_r} meson-gxl-s905x-libretech-cc.dtb

fatload mmc 0:1 ${script_addr_r} cmdline.txt
env import -t ${script_addr_r} ${filesize}

booti ${kernel_addr_r} - ${fdt_addr_r}
