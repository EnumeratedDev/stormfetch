source fetch_script_functions.sh

echo -e "${C3}Distribution: ${C4}${DISTRO_LONG_NAME} ($(uname -m))"
echo -e "${C3}Hostname: ${C4}$(cat /etc/hostname)"
echo -e "${C3}Kernel: ${C4}$(uname -s) $(uname -r)"
echo -e "${C3}Packages: ${C4}$(get_packages)"
echo -e "${C3}Shell: ${C4}$(get_shell)"
echo -e "${C3}CPU: ${C4}${CPU_MODEL} (${CPU_THREADS} threads)"
if [ ! -z "$GPU_MODEL" ]; then
  echo -e "${C3}GPU: ${C4}${GPU_MODEL})"
fi
echo -e "${C3}Memory: ${C4}${MEM_USED} MiB / ${MEM_TOTAL} MiB"
if xhost >& /dev/null ; then
  if [ ! -z "$DE_WM" ]; then
    echo -e "${C3}DE/WM: ${C4}${DE_WM}"
  fi
  if command_exists xdpyinfo ; then
    echo -e "${C3}Screen Resolution: ${C4}$(get_screen_resolution)"
  fi
fi
