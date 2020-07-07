// Package crds Code generated by go-bindata. (@generated) DO NOT EDIT.
// sources:
// appcatalog.appscode.com_appbindings.v1.yaml
// appcatalog.appscode.com_appbindings.yaml
package crds

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"time"
)

func bindataRead(data []byte, name string) ([]byte, error) {
	gz, err := gzip.NewReader(bytes.NewBuffer(data))
	if err != nil {
		return nil, fmt.Errorf("read %q: %v", name, err)
	}

	var buf bytes.Buffer
	_, err = io.Copy(&buf, gz)
	clErr := gz.Close()

	if err != nil {
		return nil, fmt.Errorf("read %q: %v", name, err)
	}
	if clErr != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

type asset struct {
	bytes []byte
	info  os.FileInfo
}

type bindataFileInfo struct {
	name    string
	size    int64
	mode    os.FileMode
	modTime time.Time
}

// Name return file name
func (fi bindataFileInfo) Name() string {
	return fi.name
}

// Size return file size
func (fi bindataFileInfo) Size() int64 {
	return fi.size
}

// Mode return file mode
func (fi bindataFileInfo) Mode() os.FileMode {
	return fi.mode
}

// ModTime return file modify time
func (fi bindataFileInfo) ModTime() time.Time {
	return fi.modTime
}

// IsDir return file whether a directory
func (fi bindataFileInfo) IsDir() bool {
	return fi.mode&os.ModeDir != 0
}

// Sys return file is sys mode
func (fi bindataFileInfo) Sys() interface{} {
	return nil
}

var _appcatalogAppscodeCom_appbindingsV1Yaml = []byte("\x1f\x8b\x08\x00\x00\x00\x00\x00\x00\xff\xb4\x5a\xdf\x73\xdb\x36\xf2\x7f\xf7\x5f\xb1\xa3\x3e\x38\x99\xb1\xa4\xe9\xb7\x2f\xdf\xd1\x3d\xf4\x1c\xc7\x9d\xc9\xd5\x71\x32\x96\x9b\x9b\xce\xe5\xe6\x0c\x12\x2b\x11\x35\x08\xb0\x00\x28\x59\xd7\xe9\xff\x7e\xb3\x0b\x50\xa2\x24\x92\x52\x9a\x09\x1f\xda\x88\x00\x16\xfb\xf3\xb3\x3f\x68\x51\xa9\x4f\xe8\xbc\xb2\x66\x06\xa2\x52\xf8\x12\xd0\xd0\x2f\x3f\x79\xfe\x7f\x3f\x51\x76\xba\xfa\xfe\xe2\x59\x19\x39\x83\x9b\xda\x07\x5b\x3e\xa0\xb7\xb5\xcb\xf1\x2d\x2e\x94\x51\x41\x59\x73\x51\x62\x10\x52\x04\x31\xbb\x00\xc8\x1d\x0a\x7a\xf9\xa8\x4a\xf4\x41\x94\xd5\x0c\x4c\xad\xf5\x05\x80\x16\x19\x6a\x4f\x7b\x00\x44\x55\x4d\x9e\xeb\x0c\x9d\xc1\x80\x7c\x8b\x11\x25\xce\x20\x17\x41\x68\xbb\xbc\x00\x88\xbf\x45\x55\x65\xca\x48\x65\x96\x7e\x22\xaa\x2a\x2d\xd3\x3f\x7d\x6e\x25\x4e\x72\x5b\x5e\xf8\x0a\x73\xa2\xba\x74\xb6\xae\xf8\x48\xe7\xb6\x48\x32\xdd\x9f\x8b\x80\x4b\xeb\x54\xf3\x7b\xdc\xba\x99\x7e\x35\x27\x9b\x9f\x2c\x00\x40\xd4\xc3\x75\x55\xbd\x89\x4c\xf1\x4b\xad\x7c\xf8\xf9\x60\xe1\x4e\xf9\xc0\x8b\x95\xae\x9d\xd0\x7b\x82\xf0\x7b\xaf\xcc\xb2\xd6\xc2\xb5\x57\x2e\x00\x7c\x6e\x2b\x9c\xc1\x3d\x71\x5a\x89\x1c\xe5\x05\xc0\x2a\x5a\x87\x39\x1d\x83\x90\x92\x95\x2e\xf4\x47\xa7\x4c\x40\x77\x63\x75\x5d\x9a\xad\x1c\xbf\x79\x6b\x3e\x8a\x50\xcc\x60\x42\x8a\x99\x84\x4d\x15\xa5\x68\x54\xfa\xb8\x7b\x41\x6b\x33\xf0\xc1\x35\xa2\x1c\x1f\x4f\x97\xef\x51\xf8\xb4\xf7\x6e\x98\x48\xe3\x1a\x93\x23\xbf\xd8\x23\x79\xbd\xdc\xe7\x49\x8a\x10\x5f\xc4\xe5\xd5\xf7\x42\x57\x85\xf8\x3e\xaa\x2e\x2f\xb0\x14\xb3\xb4\xdf\x56\x68\xae\x3f\xbe\xfb\xf4\xc3\x7c\xef\x35\x40\xe5\x6c\x85\x2e\x6c\x4d\x1c\x9f\x96\xb7\xb7\xde\x02\x48\xf4\xb9\x53\x55\xe0\x30\xb8\x24\x82\x71\x17\x48\x72\x73\xf4\x10\x0a\x6c\x2c\x81\x32\xf1\x00\x76\x01\xa1\x50\x1e\x1c\x56\x0e\x3d\x9a\xc0\x22\xee\x11\x06\xda\x24\x0c\xd8\xec\x37\xcc\xc3\x04\xe6\xe8\x88\x0c\xf8\xc2\xd6\x5a\x42\x6e\xcd\x0a\x5d\x00\x87\xb9\x5d\x1a\xf5\xdf\x2d\x6d\x0f\xc1\xf2\xa5\x5a\x04\x4c\xce\xb4\x7b\xd8\xf2\x46\x68\x58\x09\x5d\xe3\x15\x08\x23\xa1\x14\x1b\x70\x48\xb7\x40\x6d\x5a\xf4\x78\x8b\x9f\xc0\x7b\xeb\x10\x94\x59\xd8\x19\x14\x21\x54\x7e\x36\x9d\x2e\x55\x68\xa2\x3c\xb7\x65\x59\x1b\x15\x36\xd3\xdc\x9a\xe0\x54\x56\x07\xeb\xfc\x54\xe2\x0a\xf5\xd4\xab\xe5\x58\xb8\xbc\x50\x01\xf3\x50\x3b\x9c\x8a\x4a\x8d\x99\x75\x13\x18\x2a\x4a\xf9\x9d\x4b\xb8\xe0\x2f\xf7\x78\x3d\x72\x8f\xf8\x70\x24\x0d\x58\x80\x02\x0a\x94\x07\x91\x8e\x46\x29\x76\x8a\xa6\x57\xa4\x9d\x87\xdb\xf9\x23\x34\x57\xb3\x31\x0e\xb5\xcf\x7a\xdf\x1d\xf4\x3b\x13\x90\xc2\x94\x59\xa0\x8b\x46\x5c\x38\x5b\x32\x4d\x34\xb2\xb2\xca\x04\xfe\x91\x6b\x85\xe6\x50\xfd\xbe\xce\x4a\x15\xc8\xee\xbf\xd7\xe8\x03\xd9\x6a\x02\x37\xc2\x18\x1b\x20\x43\xa8\x2b\xf2\x5f\x39\x81\x77\x06\x6e\x44\x89\xfa\x46\x78\xfc\xe6\x06\x20\x4d\xfb\x31\x29\xf6\x3c\x13\xb4\x51\xfb\x70\x73\xd4\x5a\x6b\xa1\x01\xd9\x1e\x7b\xed\x90\x6f\x5e\x61\x4e\x86\x23\xdd\xd1\x21\x58\x58\x47\x18\xb7\x77\xb6\x3b\x36\xe9\x89\xea\xbe\xb1\x66\xa1\x96\x87\x6b\x07\x77\xde\xb4\xb6\x6e\xc3\xb4\xb0\x6b\x0a\x9c\xa4\x4c\x82\x79\x58\xab\x50\x30\x3b\x94\x74\x8e\x48\x02\x3c\xe0\xef\xb5\x72\x0c\xb5\xfb\x4f\x3f\x97\xcc\xa9\x78\x53\x1b\xa9\xb1\x6b\xed\x90\xd3\xeb\xb8\x35\x3a\xf4\xc7\xdb\xf7\x80\x86\xb2\x8b\x84\x9b\x6b\xc8\xe2\xd2\xba\x50\x79\x01\x6b\xa5\x35\x64\xd8\x49\x12\xa0\xf6\x28\x49\xba\x95\xd0\x8a\x3c\x2c\x2a\x19\xdd\x8a\xa2\x21\x27\x56\x17\x51\xe4\x06\x97\x7a\x24\x06\x32\x4a\x29\xc2\x0c\xb2\x4d\xe8\xbe\xac\xc7\x67\x9a\x47\x19\x8f\x79\xed\x70\xfe\xac\xaa\xc7\xbb\xf9\x27\x74\x6a\xb1\x39\x43\x13\xef\xba\xce\x81\x54\x5e\x64\x1a\x3d\x3c\xde\xcd\xf7\xe4\x58\xd1\x3a\xfd\xf3\x18\x55\x9b\x67\x5d\xa0\x69\x99\x9b\x34\x91\x0c\x9e\xe4\x87\x47\xfa\x97\xf2\x24\x8c\x35\x4b\xcd\xd7\xe5\xb6\x76\x62\x49\x21\x0a\xbf\xda\xba\x87\x74\x82\xe8\xda\x47\x45\xef\xac\x68\x7c\x40\x21\xbb\x35\x1b\x15\x97\x59\xab\x51\x74\xf1\xcc\xe6\xca\xcf\xf1\x9a\xd1\x53\xda\xfb\x14\xfd\xc6\xe1\x02\x1d\x1a\x82\x39\xbb\xb3\x7c\x8e\x1c\x61\x1d\xc8\xd7\x3c\xac\x84\x5b\x15\x0a\x74\xb0\x23\x69\x1d\x3c\xd5\x4e\x3f\x41\x59\x7b\x06\x2d\x0a\x56\xb5\x50\xa4\x93\xcf\x06\xde\x91\x07\xf5\xf9\xe1\x1a\xb3\xc2\xda\x67\x62\xcb\xd5\xc6\x34\x3a\x57\x26\x21\x66\xed\x03\xba\x2b\xfa\x61\x60\x63\xeb\xb6\x22\xb7\x0c\x4c\x46\x9d\xc4\x87\x63\x0e\x9a\x8a\xa0\x67\xed\x30\x8b\x3c\xd1\xe6\xa7\x06\x8e\xe8\x47\x0c\x8d\xad\xee\x26\xdb\xe8\xbf\xec\x25\x79\x22\x14\x98\x6b\x2a\x76\xce\xe5\x89\x36\x47\x93\x1a\xb0\x55\xac\xe5\xe0\x97\x87\x3b\xa6\x72\x16\x0e\x00\xfb\x91\x09\xa0\x0c\x08\xb3\x69\xd2\x50\xf4\x0b\xf2\xf4\x24\xdc\xd7\xc9\x64\x5d\x38\x53\xa6\xc7\x02\x79\x3b\x84\x42\x84\x86\x77\xc0\x97\xca\x12\x60\x65\x9b\x13\x60\x04\x2d\x40\x52\x26\xfc\xf0\x7f\x27\xd8\xa6\xe2\x67\x89\xae\x67\xd7\xef\x35\xba\x1e\x28\x3a\x62\xfc\xf2\x89\x77\xb3\x35\xb6\xa6\x68\xb0\x99\x97\x92\x8e\xae\xd8\xc1\x6d\x7d\x58\x08\xb4\x9f\xcb\xcb\x1f\x2f\x2f\xf7\xed\xf7\xed\xad\xc4\xc5\xe2\xd9\xf1\x30\x4f\x31\xee\x13\x9b\xf1\x34\x71\x54\x7b\xbc\x62\x20\xc1\x17\x51\x56\x7d\x59\x2d\x3e\x54\xbc\x5c\xc5\x12\x86\x70\x62\x0b\x1c\x29\xe2\x55\x72\x01\x51\x55\x5a\xa1\x04\xe1\xa1\x72\xb8\x50\x2f\x03\x24\x19\x3a\xa8\x06\x4b\x6e\x90\xc4\x9a\x4e\xe9\x02\xaa\xaa\x0e\x2f\x31\x96\xf0\xa6\x4f\x2b\xf4\x34\x26\x88\x77\x7f\x55\x80\xbb\x84\x11\xdd\x4a\x19\x33\xb0\xf4\x2c\x51\x58\xf4\x2c\x45\x19\x07\x92\xc8\x51\x11\xd6\x3c\xb5\xd3\x67\xe5\x0f\xc6\xf7\xa5\x5a\xa5\xf6\x45\xdb\x98\x49\x1b\x0c\x14\x55\x75\x45\x9a\xf7\x41\x18\x29\xdc\x71\x01\x14\x1f\x82\x26\xb2\x0b\xbc\x7a\xfa\xd7\xd6\x2e\xff\x2e\xac\x0f\x33\x92\x6e\xca\x78\xf6\x7a\x02\xb7\x2f\x22\x0f\x7a\x03\xd6\x30\xca\xf2\xed\x3d\x24\x6d\x3b\x13\x75\x27\x20\xc2\x94\x27\xba\xe4\xa9\x49\x1f\xe4\x06\x9c\x03\x7b\x88\x06\x4b\xdd\x42\xca\x89\x4d\x5e\xda\xcf\x49\x7f\xdb\x26\xf3\xdd\xf5\x0b\x85\xba\x4f\xf4\x26\xd3\x33\x37\xc4\x0c\x94\x6a\x59\x30\xb7\xd4\x73\xe8\x15\xb5\x57\x4a\x00\xbe\xa4\x76\xec\xed\xfd\x9c\x35\x6a\x7b\x0c\xcb\x0d\xa8\x4f\xfd\xc7\x2b\x9c\x2c\x27\x57\xf0\xf4\x5c\x67\x38\xde\xbe\x7f\x82\x3c\x36\x12\xe9\x06\x50\x66\x9c\xd8\xef\x21\x49\x97\x52\xbf\xc8\xe0\xcb\xaa\xca\x10\x04\x68\xb1\xc1\xd8\x3a\x29\xab\xd9\xf0\xaf\x27\x8d\x4a\xa9\xf5\x11\xda\xdb\x1e\x8a\x74\xde\xc0\xbb\x8f\x20\xa4\x74\xe8\x3d\x5b\xe4\x3a\x26\xa8\x16\x54\xc6\xbe\x53\x2d\x20\xf5\x56\x44\x76\x88\x62\x83\xa6\x50\xa1\x2b\x95\xf7\x2a\xe3\x6a\x0a\x04\xf9\xd8\x84\x2a\x31\x66\xac\xb1\x11\x5f\x17\xfa\x78\xac\x84\xe7\x14\x2a\x5c\xa6\x82\x13\x5b\xa8\x6e\xaa\x23\xf6\xee\x16\xa2\x5d\x81\x80\x61\x3d\x2a\x49\xdd\xd4\x42\xa1\x8b\xf2\x86\x80\x65\x15\x12\x49\x62\x4a\xd0\x7f\x1d\x79\x6f\x26\xbc\xca\x41\xd4\xa1\x00\x32\x22\x7c\x1e\xd1\xca\x8c\x78\x5a\x5b\x27\xff\xfe\xb9\xbb\xba\x01\xd2\x1e\xd9\x56\x68\x6d\xd7\xe4\xe9\x3f\x39\xb1\x2c\xa9\x2d\x85\x57\x9f\x47\xdf\x4d\x26\x93\xcf\xa3\xd7\xac\xd5\x98\x7d\x2a\xe1\x44\x89\x81\xbd\xe5\xf3\xe8\xc7\xb8\xde\xe7\x59\x0e\xdb\xb4\xaf\x00\xb9\xe6\xeb\x29\xb4\x06\x41\x6f\x00\x7f\x76\x1c\x9d\x68\xcf\x46\x1f\x77\xbc\xc7\x46\x1e\x43\x83\x3c\x2d\xb1\x82\xe5\x8e\x39\x76\x36\x5d\x6d\x96\x35\x86\x1a\xf8\x9d\x55\x63\x34\x2a\xa3\x95\x41\xf8\xf5\xfa\xfd\xdd\xf4\x1f\xf3\x0f\xf7\x50\x89\x8d\xb6\x42\x26\x82\xc1\x09\xe3\x35\x75\xe1\x9d\xdd\x4b\xb0\x40\x98\xbe\x12\x9a\xdc\x96\xcf\x37\x03\x9a\x84\x3d\x2d\xee\x19\x21\x48\x86\xfb\x0f\x8f\xe0\x31\x77\xd8\x05\xca\xd6\x41\xec\x6d\x64\x93\xf0\xd7\x14\x64\x46\x36\xf8\x75\x7f\xfb\xe9\xf6\xa1\x25\x2c\x14\x56\x4b\xaa\x10\xbc\x0a\x6a\xd5\x85\x17\xca\xc4\x7c\xa8\xac\x99\xc0\xa3\x65\x0d\xb6\x55\x47\x01\x9f\x5b\x13\x04\x41\x0e\xf3\xd5\x3e\x72\xd5\x41\xb1\x55\x8d\x5f\xdf\xfd\xf3\xfa\xd7\x39\xf8\x60\x1d\x46\x52\xad\xb3\x31\x2a\xe7\x4c\xb3\xc3\x81\x06\xf3\xd3\xcb\x78\x37\xd9\x1d\x63\x99\xa1\x94\x28\xc7\xcd\x8c\x66\x06\xc1\xd5\xc7\xc2\xee\x1d\x62\x38\x71\x2b\x1c\xd7\xe6\xd9\xd8\xb5\x19\xb3\x05\x7c\xe7\xd1\x28\xf7\x09\x5f\x9c\x27\xe5\x74\xf5\x01\xbc\x12\x6c\x1c\x5c\x63\x93\x30\x76\x03\x8d\xcb\xae\xbe\xca\x34\x03\xda\x56\xc9\xcb\xe6\xe4\x64\xe3\x90\x91\x44\x68\x0f\xc2\x7b\x9b\x2b\xf2\xc3\xdd\x1c\x62\x47\xfb\xb8\x1e\x1e\xee\x7f\xfa\x7b\x9f\xfd\x3a\xef\xbe\x25\x61\x6a\x1b\x43\xe7\xfc\x69\x7f\x06\x2f\x6d\xee\xa7\xb9\x35\x39\x56\xc1\x4f\xed\x8a\x52\x24\xae\xa7\x6b\xeb\x9e\x95\x59\x8e\x49\x80\x71\x34\xba\xe7\x79\xbd\x9f\x7e\xc7\xff\xeb\x01\xa4\xc7\x0f\x6f\x3f\xcc\xe0\x5a\x4a\xb0\xdc\x7c\xd6\x1e\x17\xb5\x8e\xe1\xe4\x27\xad\x51\xec\x15\x8f\x03\xaf\xa0\x56\xf2\xc7\xee\x32\xed\xaf\xa2\x55\x34\xef\x23\x81\x01\xf9\xf6\x29\xcc\xba\x53\x3e\x62\x54\x73\x80\x83\x21\x45\x5a\x8a\x9b\x0c\xb7\x95\x6d\xc4\xa4\x2e\xd0\x3a\xe1\x01\xf3\x58\x7c\x24\x2f\x80\x0c\x17\x31\x08\x71\xc3\x28\xae\x8c\x47\x37\x00\x5d\x91\x04\xc7\xe6\xd1\x16\x15\xb0\x4b\xcc\xa3\x4e\x60\x5f\x31\x09\xa1\x95\x59\x6a\x3c\x90\x3e\x61\x43\xb7\x91\xf7\x35\xb1\x27\xb7\xc3\x50\x3b\x83\x72\x37\x57\xcd\x9c\x7d\x46\xd7\x96\xb6\x9b\x66\x4b\x03\x87\xf2\x9e\xa1\xcd\xee\x1e\xf3\x0d\xe6\x82\x52\xb8\x54\x8b\x18\x0e\x89\x1b\xea\x4d\xec\x4a\xc9\x66\x9e\xec\x29\x72\xc8\xa1\xc8\x0d\x9a\x62\xb2\xaf\xac\x41\x91\x17\x49\x4e\x10\x2d\xd2\x6d\x35\xf8\xe0\x6a\x1e\xd9\x5e\x71\xf1\xe0\xa9\xba\x4b\xa5\x6e\x37\x51\xe2\xe2\x8b\xfc\xaf\x51\x0d\xd5\xbf\x52\x54\xdd\xed\x86\x0a\x1e\xd0\x04\x47\xbd\x5f\xb0\xb0\x2e\x44\xc0\x15\x4f\xbe\x77\x73\xa4\xdc\x1a\x5f\x97\x48\x15\x53\x45\x31\x3e\x81\x9f\x5a\xe5\x53\x2f\xb3\x9d\x46\xe7\xa6\x7f\x6b\xf2\x38\x69\xcf\x75\x2d\x63\x65\xf7\x8c\x1b\x18\xfd\x32\xbf\x7d\xb8\xbf\x7e\x7f\x3b\xea\x26\x9d\xd5\x69\x00\xdf\x70\x95\xba\xb0\x88\xe1\xa4\x4b\xc6\xf1\x98\xee\x9b\x59\x43\x6d\x64\x14\xaa\x93\x24\x5f\xfb\xf6\xcd\x7f\xe8\xe6\x51\xab\xb8\xb7\x50\x88\x15\xb6\x7d\x09\x6e\xe2\xe7\xc0\x9d\x25\x7a\x89\x46\xed\x73\x5b\x0a\x0b\x4b\xb5\x17\xf9\xd2\x61\x7c\x1d\x35\x39\x94\x68\x0e\x1c\x97\x3f\xb8\x1d\x20\x56\x5f\xcb\xf9\xc7\xc8\x21\xc9\xff\x33\x6e\x46\x33\xf8\x63\x44\x41\x36\x9a\xb5\x95\x0a\xa3\x60\xe9\x4d\x23\xef\x9f\x7f\xc2\x07\x13\xdb\xb3\x4e\x9a\x29\x5d\x1c\x30\x7e\x79\xe9\xa1\xa4\x24\x9e\xbe\x97\xec\xf5\x69\x5d\x58\x7d\x6a\x80\x27\xa4\xfc\x19\x7b\xe7\x33\xfb\x1f\x15\x78\x6b\xeb\xd3\x0d\x88\x3d\x7b\x88\x40\xd4\x62\x13\xb0\xfd\x2a\xda\xdb\xe5\x93\xf1\x3b\x60\x6a\xde\x57\xcf\x9d\x23\x4c\xa2\x3b\x34\x2c\x39\x1a\x94\xb5\xab\x8f\xc4\x93\x90\x7d\x0d\x28\x9c\x37\x05\x82\xb4\xfc\x49\xe8\x7a\x70\x74\x73\xc4\x4d\xea\x99\x5e\x19\x6b\xc6\x99\x32\xc2\x6d\x5e\xa7\x4f\x6d\x91\xaf\xfe\x1c\xb7\x7b\x12\xfe\x6c\x63\xaf\xe5\xe4\xcf\xb8\xe9\x9f\xf9\x9d\x29\xda\xea\x8b\x85\x8a\x82\x24\x39\x5e\x55\x96\x3b\xcd\x0d\x90\x8c\xf1\xae\xd7\xa7\xb5\x0e\x07\xe8\xda\x27\x1d\xbc\x5b\x40\x66\x43\x91\x6e\x13\x66\x98\x68\xcb\x4e\x9c\xe8\x0e\xe7\x5a\x91\x8a\xf2\xa0\x96\xc6\x52\x2f\xc1\x0d\xc4\xee\xd0\x20\x71\xfe\xc8\x41\xa7\x86\x74\x7e\xf2\xd3\x4f\x92\xfe\xb4\x69\x86\xc7\x62\x84\x52\xcf\x3d\x29\x9e\xc7\x5f\x27\x85\x1a\x47\x6d\xf4\x8d\x7b\x86\x67\x64\x0d\xd2\xf8\x9f\x9c\xed\x41\xd1\x4e\xb8\xe1\xfd\x83\x98\x53\xa2\x5b\xf6\xd6\xbc\x40\xed\x77\xfa\x8a\x1c\xb3\x6d\xfc\xfc\x8f\x2f\xca\x87\x5d\x62\xd8\xd5\x35\x2d\x2c\xea\x25\xf9\xd5\x18\x15\x93\xca\x03\x2e\xbe\x28\x8c\x8e\x3e\x38\x35\x85\xc6\x5e\x35\x32\xe8\x91\xac\x2b\xd9\x29\x6d\x6f\x05\xfb\x65\xa2\xc1\xc9\x4f\x42\x1d\xd2\x75\xf6\x47\x27\x08\x9c\x85\x55\xd0\xee\x0d\xbf\x98\xa5\xd8\x51\x7e\x1b\xbe\x4e\x86\xcb\x19\x5b\x1c\x96\x76\x85\xe7\xa6\xef\x87\x66\xf7\x60\x34\x45\x9a\x1e\x44\x2f\xef\xc7\x3e\xc3\xb1\xd5\x87\x2c\xdf\x22\x67\xa7\x3c\x1d\x79\xdd\x35\x34\x83\x71\x09\xdf\x1e\x44\xcf\x30\x58\xaa\x11\xcf\x34\x58\xda\x7d\xc2\x60\xec\xe0\x7f\xc1\x60\x97\x7e\x40\x96\x73\xcc\xb6\x18\x80\xf2\x23\x69\x7a\x6a\xad\xc8\xfe\xd7\xd6\x24\xc1\x7e\x19\x1f\xb8\x8e\xbc\xc4\x4f\xf5\x38\xa0\x87\x33\x59\x38\xed\x36\xa4\xac\xde\xc5\xde\xa9\xfa\x09\x97\x1a\x5c\x8e\x8b\xc2\xb9\xa3\x76\x99\x57\x86\x47\x2f\x8f\xd4\x73\x37\x93\xd1\x85\xc8\x95\x56\x41\x04\x24\xbf\x58\x3a\x51\x52\x27\x9c\x43\x21\x8c\xd4\x94\x45\x29\xa9\x52\x13\x1c\x3f\x68\x1d\x43\xe4\x80\x06\x57\x5d\x7f\xfe\x77\xc4\x4e\xf3\xe7\x7f\xdf\x9e\xa3\x6e\x4b\x8e\xf7\xfe\x12\xea\xe2\xa4\x0d\x8e\x5e\xf2\x00\x55\xb6\x46\xa6\x54\x19\x8a\x65\x7b\xfe\xea\xeb\x6c\xfb\x47\x7b\x33\xf8\xe3\xcf\x8b\xff\x05\x00\x00\xff\xff\x12\x18\x71\xff\x0b\x2c\x00\x00")

func appcatalogAppscodeCom_appbindingsV1YamlBytes() ([]byte, error) {
	return bindataRead(
		_appcatalogAppscodeCom_appbindingsV1Yaml,
		"appcatalog.appscode.com_appbindings.v1.yaml",
	)
}

func appcatalogAppscodeCom_appbindingsV1Yaml() (*asset, error) {
	bytes, err := appcatalogAppscodeCom_appbindingsV1YamlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "appcatalog.appscode.com_appbindings.v1.yaml", size: 11275, mode: os.FileMode(420), modTime: time.Unix(1573722179, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _appcatalogAppscodeCom_appbindingsYaml = []byte("\x1f\x8b\x08\x00\x00\x00\x00\x00\x00\xff\xb4\x5a\xdd\x6f\xdb\x38\x12\x7f\xcf\x5f\x31\xf0\x3e\xa4\x05\xfc\x81\xde\xbe\x1c\x7c\x0f\x7b\x69\x9a\x02\xdd\x6d\xd3\x22\xce\xf6\xb0\xb8\x1e\x2e\x94\x38\xb2\x78\xa1\x48\x2d\x49\xd9\xf1\x2d\xf6\x7f\x3f\x0c\x49\x59\xb2\x4d\xc9\xe9\xde\xd6\x0f\x6d\x2c\x92\xc3\xf9\xfc\xcd\x87\xcc\x6a\xf1\x19\x8d\x15\x5a\x2d\x81\xd5\x02\x9f\x1c\x2a\xfa\x66\xe7\x8f\x7f\xb5\x73\xa1\x17\x9b\x57\x19\x3a\xf6\xea\xe2\x51\x28\xbe\x84\xeb\xc6\x3a\x5d\xdd\xa1\xd5\x8d\xc9\xf1\x0d\x16\x42\x09\x27\xb4\xba\xa8\xd0\x31\xce\x1c\x5b\x5e\x00\xe4\x06\x19\x3d\xbc\x17\x15\x5a\xc7\xaa\x7a\x09\xaa\x91\xf2\x02\x40\xb2\x0c\xa5\xa5\x3d\x00\xac\xae\xe7\x8f\x4d\x86\x46\xa1\x43\x7f\x95\x62\x15\x2e\x21\x67\x8e\x49\xbd\xbe\x00\x08\xdf\x59\x5d\x67\x42\x71\xa1\xd6\x76\xce\xea\x3a\x2e\xd3\x9f\x36\xd7\x1c\xe7\xb9\xae\x2e\x6c\x8d\x39\x51\x65\x9c\x7b\x76\x98\xfc\x64\x84\x72\x68\xae\xb5\x6c\x2a\xe5\x6f\x9c\xc1\x8f\xab\x8f\xb7\x9f\x98\x2b\x97\x30\xa7\x03\x73\xb7\xab\xd1\xb3\x12\x2e\xba\x6f\xbf\xd2\xf3\x25\x58\x67\x84\x5a\x27\x0f\x6e\x82\xc6\x7a\x67\x3f\xf7\x9e\x8c\x1d\x6f\xd5\x34\x3f\xd1\x51\x8f\xd8\xd5\xba\xcf\x07\x67\x8e\xbe\xae\x8d\x6e\x6a\xaf\x8d\xa4\x06\xc2\xd9\xa8\xda\x9c\x39\x5c\x6b\x23\xda\xef\xb3\x9e\x52\xe9\x5b\x7b\xb2\xfd\xea\x6d\x03\x10\x4c\x7c\x55\xd7\xaf\x83\xbe\xfd\x43\x29\xac\xfb\xe9\x68\xe1\xbd\xb0\xce\x2f\xd6\xb2\x31\x4c\x1e\xd8\xc8\x3f\xb7\x42\xad\x1b\xc9\x4c\x7f\xe5\x02\xa0\x36\x68\xd1\x6c\xf0\x67\xf5\xa8\xf4\x56\xbd\x15\x28\xb9\x5d\x42\xc1\xa4\x25\x5e\x6c\xae\x49\xe0\x5b\x12\xa4\x66\x39\x72\x7a\xd6\x64\x26\x7a\x9b\x5d\xc2\x6f\xbf\x5f\x00\x6c\x98\x14\xdc\x2b\x2f\x48\xa7\x6b\x54\x57\x9f\xde\x7d\xfe\x7e\x95\x97\x58\xb1\xf0\x90\x2e\xd3\x35\x1a\xb7\x57\x42\xf0\xb9\xbd\xb7\xef\x9f\x01\x70\xb4\xb9\x11\xb5\xa7\x08\x97\x44\x2a\xec\x01\x4e\xfe\x8d\x16\x5c\x89\x10\x6d\x8e\x1c\xac\xbf\x06\x74\x01\xae\x14\x16\x0c\x7a\xb1\x94\xf3\x2c\xf5\xc8\x02\x6d\x61\x0a\x74\xf6\x1f\xcc\xdd\x1c\x56\x24\xba\xb1\x60\x4b\xdd\x48\x0e\xb9\x56\x1b\x34\x0e\x0c\xe6\x7a\xad\xc4\x7f\xf7\x94\x2d\x38\xed\xaf\x94\xcc\x61\x54\x74\xfb\xf1\x4e\xad\x98\x24\x25\x34\x38\x05\xa6\x38\x54\x6c\x07\x06\xe9\x0e\x68\x54\x8f\x9a\xdf\x62\xe7\xf0\x41\x1b\x04\xa1\x0a\xbd\x84\xd2\xb9\xda\x2e\x17\x8b\xb5\x70\x6d\x7c\xe7\xba\xaa\x1a\x25\xdc\x6e\x91\x6b\xe5\x8c\xc8\x1a\xa7\x8d\x5d\x70\xdc\xa0\x5c\x58\xb1\x9e\x31\x93\x97\xc2\x61\xee\x1a\x83\x0b\x56\x8b\x99\x67\x5c\x39\x0f\x12\x15\xff\x6e\x6f\x9e\xcb\x1e\xa7\x47\x31\x10\x3e\xde\xbf\x06\xf5\x4e\x4e\x06\xc2\x02\x8b\xc7\x02\xff\x9d\x7a\xe9\x11\x69\xe5\xee\x66\x75\x0f\xed\xa5\xde\x04\x87\x3a\xf7\xda\xee\x8e\xd9\x4e\xf1\xa4\x28\xa1\x0a\x34\xc1\x70\x85\xd1\x95\xa7\x88\x8a\xd7\x5a\x28\xe7\xbf\xe4\x52\xa0\x3a\x54\xba\x6d\xb2\x4a\x38\xb2\xf4\xaf\x0d\x5a\x47\xf6\x99\xc3\x35\x53\x4a\x3b\xc8\x10\x9a\x9a\x42\x94\xcf\xe1\x9d\x82\x6b\x56\xa1\xbc\x66\x16\xbf\xb9\xda\x49\xc3\x76\x46\x2a\x3d\xaf\xf8\x3e\x38\x1f\x6e\x0c\xda\xda\x3f\x6e\x71\x34\x69\xa1\x2e\xfe\x57\x35\xe6\x64\x2a\xd2\x17\x1d\x81\x42\x1b\x8a\xf4\xde\xc9\x54\xf4\x79\x68\xf2\xea\xbd\xd6\xaa\x10\xeb\xc3\x95\xa3\xdb\xae\x7b\x1b\xf7\x81\x58\xea\x2d\x05\x47\x54\x1e\xc1\x1c\x6c\x85\x2b\x3d\x23\x94\x4f\xe0\x0e\x7f\x6d\x84\xf1\xc8\xd1\xff\x0c\x71\xe3\x39\x62\xaf\x1b\xc5\x25\x9e\xae\x1c\x73\x74\x15\x36\x06\x27\xfd\x74\xf3\x01\x50\x11\x8a\x72\xb8\xbe\x82\x2c\x2c\x6d\x4b\x91\x97\xb0\x15\x52\x7a\xcf\xb0\x27\x9c\x44\xe5\xeb\x16\xc5\x30\x28\x11\xcd\x86\xfc\x3b\x27\x26\x8b\x20\x58\x8b\x2f\x24\x57\x82\x48\xa1\x4d\xc5\xdc\x12\xb2\x9d\xc3\xc4\x72\xd2\x0f\xda\x8f\x50\x16\xf3\xc6\xe0\xea\x51\xd4\xf7\xef\x57\x9f\xd1\x88\x62\x77\x56\xfe\x77\xa9\x53\xc0\x85\x65\x99\x44\x0b\xf7\xef\x57\x07\xfc\x6f\x68\x9d\xfe\x3c\x46\xc5\xf6\xb3\x2d\x51\xf5\x4c\x49\xf2\x47\x63\x46\xa9\xe1\x9e\xfe\x12\x96\xc4\xd0\x6a\x2d\xfd\x65\xb9\x6e\x0c\x5b\x53\xb8\xc1\x2f\xba\x49\x12\x8e\x00\xdb\xd8\xa0\xdc\xce\x6e\xca\x3a\x64\x3c\xa5\xcd\xa0\xae\x4c\x6b\x89\xec\x94\x5b\x6f\x9e\xfc\xbc\x87\x4c\x1e\xe2\xce\x87\xe0\x23\x06\x0b\x34\xa8\x08\xa6\x74\x67\xe7\x1c\x7d\xbc\x8c\x19\x17\xe0\x46\xb8\x12\x0d\x74\x04\xb5\x81\x87\xc6\xc8\x07\xa8\x1a\xeb\x61\x87\x02\x4f\x14\x82\x34\xf1\x45\xc1\xbb\xc2\x5f\xb0\xc5\xac\xd4\xfa\x31\x49\x92\x72\x55\xa3\x54\xab\x67\xa1\x22\xde\x35\xd6\xa1\x99\xd2\x17\x05\x3b\xdd\xf4\xd5\xb7\xbf\x7e\x3e\x49\x90\x1c\x8b\x2a\x68\xab\x99\xe4\xca\x31\xf6\x3f\xd0\xd6\x87\x16\x52\xe8\x4b\x70\xff\xbd\xc6\xba\xc8\xbe\x1c\x20\x38\xea\xf0\x9e\x5b\xaa\xc0\x9e\xc7\x0d\x6d\x0d\x26\x54\xa0\xeb\x50\x50\xc2\xcf\x77\xef\x3d\x8d\xa3\x18\xb7\xc7\xd9\xe2\x40\xe5\x0a\x98\xda\xb5\x89\x23\x78\x01\xf9\x73\x14\xea\x8f\xcb\xa2\x8d\x7b\x96\x2c\xf7\x25\xfa\xcd\xe0\x4a\xe6\xf6\x3c\xe3\x53\xad\x2d\x72\xc8\x76\x67\xbc\xb0\x83\x19\xa1\xdc\xf7\x7f\x19\x65\x97\x4a\x93\x35\x9a\xe4\x9e\x5f\x1b\x34\x49\x80\x39\x61\xf8\xf2\xc1\xef\xf5\xda\xdf\xab\xbe\xc5\x59\xbf\x14\xf5\x32\xf5\x4e\xac\x9b\x61\xe5\x5f\x5e\xfe\x70\x79\x99\xb0\xd6\x37\xb3\x8a\x2f\xdf\x9e\xe9\xf1\xab\x18\xbd\x36\x32\x18\xce\x12\x2f\x8d\xc5\xa9\x07\x08\x7c\x62\x55\x2d\x31\x94\x0f\xd3\x41\x31\x7d\x71\x41\xf1\xbf\x07\x84\x18\xcb\x22\x1a\x9c\xd5\xb5\x14\xc8\x81\x59\x2a\xc0\x0b\xf1\x04\x3e\xf4\x8f\xea\xa6\xfe\xa7\x35\x7a\x14\x68\xb1\x20\xf2\x54\xed\x1c\x5f\xa1\x34\xe1\xc8\x7a\xaf\xde\x40\xff\x0f\x07\xa9\x89\x31\x9e\x52\xe1\xcc\xc3\x42\x72\x81\x1c\x3c\xb9\x10\xf8\x1f\x84\xfb\xa3\xe2\xa7\xfd\x34\x46\x3e\x03\xe9\x3d\x16\xaf\xc5\x26\xb6\x07\x52\x87\x4c\xd7\xe2\x16\xab\xeb\x29\xe9\xd9\x3a\xa6\x38\x33\x9c\xe0\x23\xa9\x14\xd2\x35\xbc\x78\xf8\xe7\x5e\xd7\xff\x2a\xb5\x75\x4b\x92\x69\xe1\x71\xe8\xe5\x1c\x6e\x9e\x58\xee\xe4\x0e\xb4\xf2\xb8\x18\xee\xd6\xbd\xec\x90\xa4\x9c\x4e\x14\x84\x08\x0f\x74\xc5\x43\x0b\xf4\x64\x58\x9f\xa9\xc8\xfb\x58\x1b\x06\x49\x92\x6d\xfe\x38\xcc\x1d\x7f\xdb\xa7\xda\x2e\x5d\x15\xd4\xdb\xed\x33\xae\xbf\x95\x2e\x4d\x33\x2a\xd6\xa5\xe7\x94\xaa\x7a\xb9\xa1\xd6\x45\x30\xc0\xa7\xd8\xea\xbc\xb9\x5d\x79\x4d\xea\x8a\xd4\x2a\x6c\xac\xe6\x5f\xe0\x7c\x3d\x9f\xc2\xc3\x63\x93\xe1\x6c\xff\x3c\xad\x8a\x3c\x14\xeb\x91\x3e\x08\x35\x8b\xac\x7b\xe2\xd4\x71\x79\x78\xf4\xea\xc8\x10\x18\x48\xb6\xc3\xd0\x84\x08\x2d\xbd\x61\x5f\xa6\x11\x32\xaa\x92\x5a\x0b\x26\xad\xf6\xa7\x15\xbc\xfb\x04\x8c\x73\x83\xd6\x7a\x9d\x5f\x85\xc4\xd1\x83\xb4\xd0\xb9\x89\x22\x8d\xee\xa1\x73\xf1\x44\x3d\xbd\x16\xf3\xa0\x46\x53\x09\x6b\x45\xe6\xab\x19\x60\xe4\x55\x73\xaa\x83\xfc\xde\x68\x85\xc1\xec\x47\xf6\xad\x99\xf5\x69\x8d\x99\x4c\x38\xc3\xf6\x70\xda\x56\x28\xde\x6f\x7b\xe8\x33\x05\xd6\x9a\x39\x5d\x54\x70\xea\x49\x0a\x81\x26\x48\xea\x1c\x56\xb5\x8b\x04\x89\x21\x46\xff\x1a\xf2\xd6\x8c\x59\x91\x03\x6b\x5c\x09\x64\x3a\xf8\x32\xa1\x95\x25\x71\xb4\xd5\x86\xff\xfd\x4b\xaa\xc6\xf0\x65\x0b\xd9\x8e\x49\xa9\xb7\xe4\xc3\x6f\x0d\x5b\x57\xd4\xd8\xc1\x8b\x2f\x93\xef\xe6\xf3\xf9\x97\xc9\x4b\xaf\xcd\x90\x1d\x6a\x66\x58\x85\xce\x7b\xc8\x97\xc9\x0f\x61\x3d\x49\x98\x19\xec\x53\x9e\x02\xfa\x9a\x2b\x59\xea\x8c\x00\xd7\x20\x96\x74\x9c\x8c\x36\x3a\x93\x4f\x1d\xc7\xa1\xfd\x45\xd7\xa2\x48\x4f\x18\xa7\xdb\x8e\x22\x74\x40\x4a\xa5\xb0\xab\xb3\x62\x88\x39\xa1\xa4\x50\x08\xbf\x5c\x7d\x78\xbf\xf8\x71\xf5\xf1\x16\x6a\xb6\x93\x9a\xf1\x48\xce\x19\xa6\xac\xa4\xee\x95\xd2\xb7\x06\xc2\xdf\x0d\x93\xa9\x92\xc6\x9f\x6e\x47\x19\x11\x47\x7a\x9c\xc7\x78\xb7\x70\xfb\xf1\x1e\x2c\xe6\x86\x84\x30\x10\x3a\x06\x1e\x53\xee\x09\xd1\x2d\x85\x8d\xe2\x2d\x12\xdd\xde\x7c\xbe\xb9\xeb\x8b\x59\x6a\xc9\x29\x67\x5b\xe1\xc4\x26\x74\xd3\x94\x99\x84\x56\x73\xb8\xd7\xa4\xa9\x13\x92\x7d\x95\x51\x50\x53\x7b\xcd\x08\x3e\x02\x4f\x3d\x12\xd3\x7e\xb5\x7b\xf5\xfe\x1f\x57\xbf\xac\xc0\x3a\x6d\x4e\x03\xc8\x13\xea\x9d\x0c\xb1\xb7\xf2\x14\x4f\xdc\x65\x24\xb7\x3c\xcd\xba\x81\xe7\x0c\xab\x0c\x39\x47\x3e\x6b\x67\x19\x4b\x70\xa6\x39\xbe\xfc\xe0\x48\x3b\x3f\x9b\x35\x61\x80\x36\x2b\xe2\x04\xed\xe4\x60\x90\x76\xd4\xef\x56\x51\x21\xa9\x9a\xdb\xaf\x90\x9b\x19\xa4\x56\x2e\xc2\x7d\x37\x00\xb8\x3c\xad\x1d\x54\x3b\xb5\xeb\x95\x9a\xde\x7c\x3e\x51\x18\xf4\x38\xc1\xa4\x05\x66\xad\xce\x85\xf7\xb9\x7d\xef\xde\x51\x3e\x46\xd9\xb1\x1e\x63\xa8\xbf\x38\xac\xb4\x6e\x7b\x92\xc5\x86\xcc\x25\xa7\x33\x87\xc3\x68\xae\x73\xbb\xc8\xb5\xca\xb1\x76\x76\xa1\x37\x94\xd8\x70\xbb\xd8\x6a\xf3\x28\xd4\x7a\x46\xac\xcf\x82\x91\xad\x1f\x5c\xdb\xc5\x77\xfe\xbf\x24\xd4\xdc\x7f\x7c\xf3\x71\x09\x57\x9c\x83\xf6\x6d\x5d\x63\xb1\x68\x64\x08\x1a\x3b\xef\x8d\x25\xa7\x7e\x48\x36\x85\x46\xf0\x1f\x52\x45\xd4\x1f\xc1\xa1\x60\xce\x7b\x0a\x75\xf2\xe0\x71\x34\x7a\x2f\x6c\x40\x9f\x76\xbb\x77\xf8\x18\x4b\x31\x56\x32\xdc\xd7\x94\x11\x6f\x7a\xf6\x3d\x61\x3a\x65\xef\x55\x28\x13\xa2\xcd\x21\xc3\x82\xcc\xe1\x4a\xdc\x79\x54\x16\xca\xa2\xd9\x83\x52\x2a\xa5\xc5\xd8\x3b\x7a\x2e\x1c\x9e\x8a\x77\x52\x79\x1f\xaa\x23\x62\xae\x50\x6b\x89\x47\x52\xc7\xb8\xb7\xad\xb4\x29\x7b\x9c\xc8\x0f\x06\x5d\x63\x14\xf2\x6e\xbe\x98\x19\xfd\x88\x66\x50\xca\x04\xd9\x56\xee\x36\x48\xcf\xeb\x70\x0e\xaf\x31\x67\x94\x70\xb9\x28\x82\x93\x27\xe8\x06\x4e\xa8\x0f\xd0\x1b\xc1\xdb\x89\xaa\xa5\x08\x21\xf7\x21\xc3\xb7\x23\x0a\x2a\x28\x90\xe5\x65\x94\x07\xd8\x28\xe1\xbe\x02\xac\x33\x8d\x1f\x5b\x4e\x7d\xea\xb7\x54\x7d\xc5\x22\x74\xe7\xef\x4b\xf9\x56\x82\xe6\xa0\xb7\xad\xf6\xf8\xc4\x38\xab\x1d\x08\x67\x01\x95\x33\xd4\x4d\x39\x0d\xdb\x92\x39\xdc\x24\xeb\x95\xfe\x0c\x26\xd7\xca\x36\x15\x52\xa5\x53\x53\x14\xcf\xe1\x6d\xbf\xec\x19\x32\x6b\x4a\xab\xbb\xbe\x99\xc3\x94\x39\x97\x0d\x0f\x35\xf1\x23\xee\x60\xf2\xf3\xea\xe6\xee\xf6\xea\xc3\xcd\x64\x0a\x59\x13\x07\xcd\xed\xfd\xb1\xeb\x49\x21\x07\xed\x23\x1d\x7a\x74\x0e\x29\xbb\xed\xdd\x1b\xc5\xfd\x20\x3b\x5e\xf0\xe6\xf5\xbf\xe9\x8e\x49\xaf\xe4\xd6\x50\xb2\x4d\xb2\xfb\xe9\xbc\x07\xae\xc3\x8b\xa1\xce\x26\x3d\x0d\x07\x25\x14\x9a\xea\x23\xf2\x95\xa3\xc8\x49\x50\x3e\x69\x39\x28\x75\x1c\x39\xaa\x7f\x83\x76\x84\x49\x4b\x98\xc1\x6f\x13\x83\x24\xe7\x4f\xb8\x9b\xa4\x50\xfd\xb7\x09\x05\xd4\x64\x79\xa0\xcc\x89\xd3\xf4\xa4\x95\xfe\xf7\xdf\xe1\xa3\xea\x1a\xa5\x4e\x94\xfd\x4d\x97\x89\xd4\x05\x50\x51\x32\x8e\x6f\x08\x0e\x3a\xa6\x53\x0c\x1e\x1f\x7a\x31\xce\x7f\xc2\x81\x49\xc7\xe1\x30\xdd\x6f\xec\xbd\xa6\x00\x76\x60\x03\xe6\x88\x56\x28\xd5\xf7\x2f\x35\x07\xba\x6a\x72\x80\x04\x10\x05\xc9\x07\x3a\x8c\xf1\xc9\x9d\xa7\x39\xb4\x94\x18\x31\xf5\xab\x87\xc8\x0d\xe3\xe9\xd1\x37\x3c\x67\x92\x02\x71\xf1\x33\x93\xcd\xe0\x40\x25\xc1\x47\xec\x65\x5e\x28\xad\x66\x99\x50\xcc\xec\x5e\xc6\xd7\x48\x81\xa3\x43\x04\x19\xa4\x0b\xbd\xe8\xea\x5c\xf9\x11\x77\x43\x53\xb2\x67\x89\xb4\xf9\x4a\x61\x82\x00\x91\xff\x17\xb5\xf6\x7d\xdf\x0e\x48\xb6\x70\xcf\xcb\x73\x7a\x86\x23\xc4\x1c\x92\x0a\xde\x15\x90\x69\x57\xc6\xbb\x98\x1a\x23\xd9\xb3\x8c\x4f\x63\xc7\x33\xa1\x40\x43\x58\x10\x6b\xa5\xa9\xf6\xf7\x05\x7e\x77\x68\x84\xb4\x1f\xf5\xd3\x99\x61\x3d\x9f\x79\xed\x11\xa5\x3e\x67\x8c\xb1\x41\x13\xc0\x8c\x94\x32\xb0\x72\x4e\x90\x59\x90\x3f\xfd\xd6\x67\x6c\xea\xd4\x62\x87\x7d\x6b\x74\xf5\x6c\x00\xf1\xbb\x47\x51\xa4\x42\xb3\x46\xbb\x7f\xc7\x9f\xe0\xca\xbf\x01\x0d\xd9\x33\xbc\xb0\xc6\x27\x61\x5d\x07\xf8\x5d\x35\xf2\xa7\xa1\x4b\x80\xff\x3b\x2c\xbe\x22\x1c\x4e\x5e\xa6\xb4\x65\xc0\x61\x65\xea\xe5\x1d\xf3\xdf\x11\x69\x86\xdd\xee\xbc\x48\x70\xe6\x85\x47\x42\xaa\x64\x67\x32\x7a\xfc\x19\x28\x03\xfd\x5e\xec\x2b\x99\x09\xfd\xdb\x9f\xcf\xd1\x19\xc7\x3f\xbb\xc1\x60\xa5\x37\xf8\xbc\xc4\x7a\xd7\xee\x1d\x8d\x8a\x40\x91\x16\xc6\x5a\x97\xf0\x89\x7e\x46\x31\x92\x46\x85\x3f\x3b\x9b\xc6\x0c\x1a\x78\xec\xda\x88\x33\x49\xeb\x9b\x81\xde\x59\xe3\xc4\xaa\xed\x59\xc6\x89\x7b\xcf\x18\xc7\x3b\xf0\x57\x1b\xe7\xd2\x0e\xca\x70\xde\x44\xc5\x20\xec\x9e\x48\x31\x50\xf1\x04\xb6\xff\x9f\x0a\xc1\xe9\xaf\xe1\x00\xb7\x81\x8b\xf0\x32\x19\x07\x65\x7f\xd6\xe5\xe7\x9c\x83\xd4\x33\xb0\xe4\xf4\xd7\xbb\xcd\xc8\x62\x58\x62\xc6\xb0\x43\x71\xfc\xf3\xb1\xd1\xc5\x3d\x75\xb0\xed\xe4\xb0\x60\xb9\x90\xc2\x31\x87\x64\xfb\xb5\x61\x15\x75\x9b\x39\x94\x4c\x71\x49\xb9\x8d\x52\x1d\xb5\x95\xe1\xb5\xcd\x31\xc8\x0d\xea\x6b\x73\xfa\xf3\xb1\x13\x46\xda\x9f\x8f\x7d\x5b\x5e\x52\x16\x9b\x1d\xfc\xbe\xe6\x62\x54\xdf\x47\x8f\x5a\xc1\x60\xf3\x8a\xc9\xba\x64\xaf\xba\x67\xf1\xe7\x93\xe1\xc7\x89\xbd\xe5\xf0\xc3\x08\xe4\xbd\xf9\x23\x95\x6f\x6c\xdd\x8e\x32\xff\x17\x00\x00\xff\xff\x6a\x69\x77\xd4\x5d\x2a\x00\x00")

func appcatalogAppscodeCom_appbindingsYamlBytes() ([]byte, error) {
	return bindataRead(
		_appcatalogAppscodeCom_appbindingsYaml,
		"appcatalog.appscode.com_appbindings.yaml",
	)
}

func appcatalogAppscodeCom_appbindingsYaml() (*asset, error) {
	bytes, err := appcatalogAppscodeCom_appbindingsYamlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "appcatalog.appscode.com_appbindings.yaml", size: 10845, mode: os.FileMode(420), modTime: time.Unix(1573722179, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

// Asset loads and returns the asset for the given name.
// It returns an error if the asset could not be found or
// could not be loaded.
func Asset(name string) ([]byte, error) {
	cannonicalName := strings.Replace(name, "\\", "/", -1)
	if f, ok := _bindata[cannonicalName]; ok {
		a, err := f()
		if err != nil {
			return nil, fmt.Errorf("Asset %s can't read by error: %v", name, err)
		}
		return a.bytes, nil
	}
	return nil, fmt.Errorf("Asset %s not found", name)
}

// MustAsset is like Asset but panics when Asset would return an error.
// It simplifies safe initialization of global variables.
func MustAsset(name string) []byte {
	a, err := Asset(name)
	if err != nil {
		panic("asset: Asset(" + name + "): " + err.Error())
	}

	return a
}

// AssetInfo loads and returns the asset info for the given name.
// It returns an error if the asset could not be found or
// could not be loaded.
func AssetInfo(name string) (os.FileInfo, error) {
	cannonicalName := strings.Replace(name, "\\", "/", -1)
	if f, ok := _bindata[cannonicalName]; ok {
		a, err := f()
		if err != nil {
			return nil, fmt.Errorf("AssetInfo %s can't read by error: %v", name, err)
		}
		return a.info, nil
	}
	return nil, fmt.Errorf("AssetInfo %s not found", name)
}

// AssetNames returns the names of the assets.
func AssetNames() []string {
	names := make([]string, 0, len(_bindata))
	for name := range _bindata {
		names = append(names, name)
	}
	return names
}

// _bindata is a table, holding each asset generator, mapped to its name.
var _bindata = map[string]func() (*asset, error){
	"appcatalog.appscode.com_appbindings.v1.yaml": appcatalogAppscodeCom_appbindingsV1Yaml,
	"appcatalog.appscode.com_appbindings.yaml":    appcatalogAppscodeCom_appbindingsYaml,
}

// AssetDir returns the file names below a certain
// directory embedded in the file by go-bindata.
// For example if you run go-bindata on data/... and data contains the
// following hierarchy:
//     data/
//       foo.txt
//       img/
//         a.png
//         b.png
// then AssetDir("data") would return []string{"foo.txt", "img"}
// AssetDir("data/img") would return []string{"a.png", "b.png"}
// AssetDir("foo.txt") and AssetDir("notexist") would return an error
// AssetDir("") will return []string{"data"}.
func AssetDir(name string) ([]string, error) {
	node := _bintree
	if len(name) != 0 {
		cannonicalName := strings.Replace(name, "\\", "/", -1)
		pathList := strings.Split(cannonicalName, "/")
		for _, p := range pathList {
			node = node.Children[p]
			if node == nil {
				return nil, fmt.Errorf("Asset %s not found", name)
			}
		}
	}
	if node.Func != nil {
		return nil, fmt.Errorf("Asset %s not found", name)
	}
	rv := make([]string, 0, len(node.Children))
	for childName := range node.Children {
		rv = append(rv, childName)
	}
	return rv, nil
}

type bintree struct {
	Func     func() (*asset, error)
	Children map[string]*bintree
}

var _bintree = &bintree{nil, map[string]*bintree{
	"appcatalog.appscode.com_appbindings.v1.yaml": {appcatalogAppscodeCom_appbindingsV1Yaml, map[string]*bintree{}},
	"appcatalog.appscode.com_appbindings.yaml":    {appcatalogAppscodeCom_appbindingsYaml, map[string]*bintree{}},
}}

// RestoreAsset restores an asset under the given directory
func RestoreAsset(dir, name string) error {
	data, err := Asset(name)
	if err != nil {
		return err
	}
	info, err := AssetInfo(name)
	if err != nil {
		return err
	}
	err = os.MkdirAll(_filePath(dir, filepath.Dir(name)), os.FileMode(0755))
	if err != nil {
		return err
	}
	err = ioutil.WriteFile(_filePath(dir, name), data, info.Mode())
	if err != nil {
		return err
	}
	err = os.Chtimes(_filePath(dir, name), info.ModTime(), info.ModTime())
	if err != nil {
		return err
	}
	return nil
}

// RestoreAssets restores an asset under the given directory recursively
func RestoreAssets(dir, name string) error {
	children, err := AssetDir(name)
	// File
	if err != nil {
		return RestoreAsset(dir, name)
	}
	// Dir
	for _, child := range children {
		err = RestoreAssets(dir, filepath.Join(name, child))
		if err != nil {
			return err
		}
	}
	return nil
}

func _filePath(dir, name string) string {
	cannonicalName := strings.Replace(name, "\\", "/", -1)
	return filepath.Join(append([]string{dir}, strings.Split(cannonicalName, "/")...)...)
}