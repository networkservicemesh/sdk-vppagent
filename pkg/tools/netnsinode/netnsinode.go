// Copyright (c) 2020 Cisco Systems, Inc.
//
// SPDX-License-Identifier: Apache-2.0
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at:
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Package netnsinode provides utility functions for working with Linux Network Namespace Inodes
package netnsinode

import (
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"unicode"

	"github.com/pkg/errors"
)

const (
	// location of network namespace for a process
	netnsfile = "/proc/self/ns/net"
)

func isDigits(s string) bool {
	for _, c := range s {
		if !unicode.IsDigit(c) {
			return false
		}
	}
	return true
}

// getInode returns Inode for file
func getInode(file string) (uint64, error) {
	fileinfo, err := os.Stat(file)
	if err != nil {
		return 0, errors.Wrap(err, "error stat file")
	}
	stat, ok := fileinfo.Sys().(*syscall.Stat_t)
	if !ok {
		return 0, errors.New("not a stat_t")
	}
	return stat.Ino, nil
}

// GetMyNetNSInodeNum - returns the inodeNumber of the Linux Network Namespace of the running process
func GetMyNetNSInodeNum() (uint64, error) {
	return getInode(netnsfile)
}

// resolvePodNsByInode Traverse /proc/<pid>/<suffix> files,
// compare their inodes with inode parameter and returns file if inode matches
func resolvePodNSByInode(inode uint64) (string, error) {
	files, err := ioutil.ReadDir("/proc")
	if err != nil {
		return "", errors.Wrap(err, "can't read /proc directory")
	}

	for _, f := range files {
		name := f.Name()
		if isDigits(name) {
			filename := path.Join("/proc", name, "/ns/net")
			tryInode, err := getInode(filename)
			if err != nil {
				continue
			}
			if tryInode == inode {
				if cmdline, err := getCmdline(name); err == nil && strings.Contains(cmdline, "pause") {
					return filename, nil
				}
			}
		}
	}

	return "", errors.New("not found")
}

// LinuxNetNSFileName returns a filename of a file from /proc/*/ns/net that has an inode matching inodeString
func LinuxNetNSFileName(inodeString string) (string, error) {
	inodeNum, err := strconv.ParseUint(inodeString, 10, 64)
	if err != nil {
		return "", errors.Errorf("inodeString must be an unsigned int, instead was: \"%s\"", inodeString)
	}
	return resolvePodNSByInode(inodeNum)
}

func getCmdline(pid string) (string, error) {
	data, err := ioutil.ReadFile(path.Join("/proc/", filepath.Clean(pid), "cmdline"))
	if err != nil {
		return "", err
	}
	return string(data), nil
}
