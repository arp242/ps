// +build linux

package ps

import (
	"fmt"
	"io/ioutil"
	"strings"
)

// Refresh reloads all the data associated with this process.
func (p *UnixProcess) Refresh() error {
	d, err := ioutil.ReadFile(fmt.Sprintf("/proc/%d/stat", p.pid))
	if err != nil {
		return err
	}

	// First, parse out the image name
	data := string(d)
	binStart := strings.IndexRune(data, '(') + 1
	binEnd := strings.IndexRune(data[binStart:], ')')
	p.binary = data[binStart : binStart+binEnd]

	// Move past the image name and start parsing the rest
	data = data[binStart+binEnd+2:]
	_, err = fmt.Sscanf(data,
		"%c %d %d %d",
		&p.state,
		&p.ppid,
		&p.pgrp,
		&p.sid)
	return err
}
