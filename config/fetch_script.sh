source fetch_script_functions.sh

echo -e "${C3}Distribution: ${C4}${DISTRO_LONG_NAME} ($(uname -m))"
echo -e "${C3}Hostname: ${C4}$(cat /etc/hostname)"
echo -e "${C3}Kernel: ${C4}$(uname -s) $(uname -r)"
echo -e "${C3}Packages: ${C4}$(get_packages)"
echo -e "${C3}Shell: ${C4}$(get_shell)"
echo -e "${C3}CPU: ${C4}$(get_cpu_name) ($(nproc) threads)"
if command_exists lshw; then
  echo -e "${C3}GPU: ${C4}$(lshw -class display 2> /dev/null | grep 'product' | cut -d":" -f2 | xargs)"
fi
echo -e "${C3}Memory: ${C4}$(get_used_mem) MiB / $(get_total_mem) MiB"
if xhost >& /dev/null ; then
  if get_de_wm &> /dev/null; then
    echo -e "${C3}DE/WM: ${C4}$(get_de_wm)"
  fi
  if command_exists xdpyinfo ; then
    echo -e "${C3}Screen Resolution: ${C4}$(get_screen_resolution)"
  fi
fi

