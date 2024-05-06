source fetch_script_functions.sh

echo "Distribution: ${DISTRO_LONG_NAME} ($(uname -m))"
echo "Hostname: $(cat /etc/hostname)"
echo "Kernel: $(uname -s) $(uname -r)"
echo "Packages: $(get_packages)"
echo "Shell: $(get_shell)"
echo "CPU: $(get_cpu_name) ($(nproc) threads)"
if command_exists lshw; then
  echo "GPU: $(lshw -class display 2> /dev/null | grep 'product' | cut -d":" -f2 | xargs)"
fi
echo "Memory: $(get_used_mem) MiB / $(get_total_mem) MiB"
if xhost >& /dev/null ; then
  echo "DE/WM: $(get_de_wm)"
  echo "Screen Resolution: $(get_screen_resolution)"
fi

