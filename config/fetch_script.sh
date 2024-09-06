source fetch_script_functions.sh

echo -e "${C3}Distribution: ${C4}${DISTRO_LONG_NAME} ($(uname -m))"
echo -e "${C3}Hostname: ${C4}$(cat /etc/hostname)"
echo -e "${C3}Kernel: ${C4}$(uname -s) $(uname -r)"
echo -e "${C3}Packages: ${C4}$(get_packages)"
echo -e "${C3}Shell: ${C4}${USER_SHELL}"
echo -e "${C3}Init: ${C4}${INIT_SYSTEM}"
echo -e "${C3}Libc: ${C4}${LIBC}"
[ -n "$CPU_MODEL" ] && echo -e "${C3}CPU: ${C4}${CPU_MODEL} (${CPU_THREADS} threads)"
for i in $(seq "${CONNECTED_GPUS}"); do
    gpu="GPU$i"
    echo -e "${C3}GPU: ${C4}${!gpu}"
  done
[ -n "$MEM_TOTAL" ] && [ -n "$MEM_USED" ] && echo -e "${C3}Memory: ${C4}${MEM_USED} MiB / ${MEM_TOTAL} MiB"
for i in $(seq "${MOUNTED_PARTITIONS}"); do
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
[ -n "$LOCAL_IPV4" ] && echo -e "${C3}Local IPv4 Address: ${C4}${LOCAL_IPV4}"
if [ -n "$DISPLAY_PROTOCOL" ]; then
  echo -e "${C3}Display Protocol: ${C4}${DISPLAY_PROTOCOL}"
  for i in $(seq "${CONNECTED_MONITORS}"); do
    monitor="MONITOR$i"
    echo -e "${C3}Screen $i: ${C4}${!monitor}"
  done
fi
[ -n "$DE_WM" ] && echo -e "${C3}DE/WM: ${C4}${DE_WM}"
