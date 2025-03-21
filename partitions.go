package main

import (
	"os"
	"path/filepath"
	"slices"
	"strings"
	"syscall"
)

type partition struct {
	Device        string
	MountPoint    string
	Label         string
	FileystemType string
	TotalSize     uint64
	UsedSize      uint64
	FreeSize      uint64
}

func GetMountedPartitions(hiddenPartitions, hiddenFilesystems []string) []partition {
	// Get all filesystem and partition labels
	fslabels, err := os.ReadDir("/dev/disk/by-label")
	if err != nil && !os.IsNotExist(err) {
		return nil
	}
	partlabels, err := os.ReadDir("/dev/disk/by-partlabel")
	if err != nil && !os.IsNotExist(err) {
		return nil
	}
	labels := make(map[string]string)
	for _, entry := range partlabels {
		link, err := filepath.EvalSymlinks(filepath.Join("/dev/disk/by-partlabel/", entry.Name()))
		if err != nil {
			continue
		}
		labels[link] = entry.Name()
	}
	for _, entry := range fslabels {
		link, err := filepath.EvalSymlinks(filepath.Join("/dev/disk/by-label/", entry.Name()))
		if err != nil {
			continue
		}
		labels[link] = entry.Name()
	}

	// Get all mounted partitions
	file, err := os.ReadFile("/proc/mounts")
	if err != nil {
		return nil
	}

	var partitions []partition
	for _, entry := range strings.Split(string(file), "\n") {
		fields := strings.Fields(entry)
		if entry == "" {
			continue
		}

		// Skip virtual partitions not under /dev
		if !strings.HasPrefix(fields[0], "/dev") {
			continue
		}

		// Skip partition if explicitly hidden
		if slices.Contains(hiddenPartitions, fields[0]) {
			continue
		}

		// Skip filesystem if explicitely hidden
		if slices.Contains(hiddenFilesystems, fields[2]) {
			continue
		}

		p := partition{
			fields[0],
			fields[1],
			"",
			fields[2],
			0,
			0,
			0,
		}

		// Skip already added partitions
		skip := false
		for _, part := range partitions {
			if part.Device == p.Device {
				skip = true
			}
		}
		if skip {
			continue
		}

		// Set partition label if available
		if value, ok := labels[p.Device]; ok {
			p.Label = value
		}

		// Get partition total, used and free space
		buf := new(syscall.Statfs_t)
		err = syscall.Statfs(p.MountPoint, buf)
		if err != nil {
			continue
		}
		totalBlocks := buf.Blocks
		freeBlocks := buf.Bfree
		usedBlocks := totalBlocks - freeBlocks
		blockSize := uint64(buf.Bsize)

		p.TotalSize = totalBlocks * blockSize
		p.FreeSize = freeBlocks * blockSize
		p.UsedSize = usedBlocks * blockSize

		partitions = append(partitions, p)
	}
	return partitions
}
