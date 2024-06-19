source fetch_script_functions.sh

echo -e "${C3}Distribution: ${C4}${DISTRO_LONG_NAME} ($(uname -m))"
echo -e "${C3}Hostname: ${C4}$(cat /etc/hostname)"
echo -e "${C3}Kernel: ${C4}$(uname -s) $(uname -r)"
echo -e "${C3}Packages: ${C4}$(get_packages)"
echo -e "${C3}Shell: ${C4}${USER_SHELL}"
if [ ! -z "$CPU_MODEL" ]; then echo -e "${C3}CPU: ${C4}${CPU_MODEL} (${CPU_THREADS} threads)"; fi
for i in $(seq ${CONNECTED_GPUS}); do
    gpu="GPU$i"
    echo -e "${C3}GPU: ${C4}${!gpu}"
  done
if [ ! -z "$MEM_TOTAL" ] && [ ! -z "$MEM_USED" ]; then echo -e "${C3}Memory: ${C4}${MEM_USED} MiB / ${MEM_TOTAL} MiB"; fi
for i in $(seq ${MOUNTED_PARTITIONS}); do
    device="PARTITION${i}_DEVICE"
    mountpoint="PARTITION${i}_MOUNTPOINT"
    label="PARTITION${i}_LABEL"
    type="PARTITION${i}_TYPE"
    total="PARTITION${i}_TOTAL_SIZE"
    used="PARTITION${i}_USED_SIZE"
    if [ -z "${!type}" ]; then
      if [ -z "${!label}" ]; then
        echo -e "${C3}Partition ${!mountpoint}: ${C4}${!used}/${!total}"
      else
        echo -e "${C3}Partition ${!label}: ${C4}${!used}/${!total}"
      fi
    else
      if [ -z "${!label}" ]; then
              echo -e "${C3}Partition ${!mountpoint} (${!type}): ${C4}${!used}/${!total}"
            else
              echo -e "${C3}Partition ${!label} (${!type}): ${C4}${!used}/${!total}"
            fi
    fi
  done
if [ ! -z "$DISPLAY_PROTOCOL" ]; then
  echo -e "${C3}Display Protocol: ${C4}${DISPLAY_PROTOCOL}"
  for i in $(seq ${CONNECTED_MONITORS}); do
    monitor="MONITOR$i"
    echo -e "${C3}Screen $i: ${C4}${!monitor}"
  done
fi
if [ ! -z "$DE_WM" ]; then echo -e "${C3}DE/WM: ${C4}${DE_WM}"; fi
