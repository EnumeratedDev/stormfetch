command_exists() {
    if [ -z "$1" ]; then
        return 1
    fi
    if command -v "$1" &> /dev/null; then
        return 0
    else
        return 1
    fi
}

get_packages() {
    ARRAY=()
    if command_exists dpkg; then
        ARRAY+=("$(dpkg-query -f '.\n' -W | wc -l) (dpkg)")
    fi
    if command_exists pacman; then
        ARRAY+=("$(pacman -Q | wc -l) (pacman)")
    fi
    if command_exists rpm; then
        ARRAY+=("$(rpm -qa | wc -l) (rpm)")
    fi
    if command_exists bpm; then
        ARRAY+=("$(bpm list -n) (bpm)")
    fi
    if command_exists emerge; then
        ARRAY+=("$(ls -l /var/db/pkg/ | wc -l) (emerge)")
    fi
    if command_exists flatpak; then
        ARRAY+=("$(flatpak list | wc -l) (flatpak)")
    fi
    if command_exists snap; then
        ARRAY+=("$(snap list | wc -l) (snap)")
    fi
    echo "${ARRAY[@]}"
    unset ARRAY
}
