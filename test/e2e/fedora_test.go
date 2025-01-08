package e2e

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"runtime"

	. "github.com/onsi/ginkgo/v2"
	"github.com/ulikunitz/xz"
)

const (
	fedoraBaseDirEndpoint = "https://kojipkgs.fedoraproject.org/compose/cloud/latest-Fedora-Cloud-41/compose"
)

// Fedora metadata for cloud downloads
/*
{
    "header": {
        "type": "productmd.images",
        "version": "1.2"
    },
    "payload": {
        "compose": {
            "date": "20231130",
            "id": "Fedora-Cloud-39-20231130.0",
            "respin": 0,
            "type": "production"
        },
        "images": {
            "Cloud": {
                "aarch64": [
                    {
                        "arch": "aarch64",
                        "bootable": false,
                        "checksums": {
                            "sha256": "09860169a88d39b865d6b5cc982134d68202d4f9b0ad36fdd14222c99749a5d3"
                        },
                        "disc_count": 1,
                        "disc_number": 1,
                        "format": "qcow2",
                        "implant_md5": null,
                        "mtime": 1701326607,
                        "path": "Cloud/aarch64/images/Fedora-Cloud-Base-39-20231130.0.aarch64.qcow2",
                        "size": 594280448,
                        "subvariant": "Cloud_Base",
                        "type": "qcow2",
                        "volume_id": null
                    },
...
*/

type fedoraCloudHeader struct {
	Type    string `json:"type"`
	Version string `json:"version"`
}

type fedoraCloudImage struct {
	Arch       string
	Bootable   bool
	Checksums  map[string]string `json:"checksums"`
	DiscCount  int               `json:"disc_count"`
	DiskNumber int               `json:"disk_number"`
	Format     string
	ImplantMd5 string `json:"implant_md5"`
	Mtime      int64  `json:"mtime"`
	Path       string
	Size       int64
	Subvariant string
	Type       string
	VolumeID   *int64 `json:"volume_id"`
}

type FedoraCloudCompose struct {
	Date   string
	ID     string `json:"id"`
	Respin int
	Kind   string `json:"type"`
}

type FedoraCloudPayload struct {
	Compose FedoraCloudCompose
	Images  map[string]map[string][]fedoraCloudImage
}

type fedoraCloudMetadata struct {
	Header  fedoraCloudHeader  `json:"'header'"`
	Payload FedoraCloudPayload `json:"payload"`
}

func (m fedoraCloudMetadata) downloadPathForVHD() (string, string, error) {
	arch := archFromGOOS()
	val, ok := m.Payload.Images["Cloud"][arch]
	if !ok {
		return "", "", errors.New("unable to parse metadata for cloud image")
	}
	for _, ci := range val {
		if ci.Format == "vhd.xz" { // fedora does not make a vhdx
			return fmt.Sprintf("%s/%s", fedoraBaseDirEndpoint, ci.Path), ci.Checksums["sha256"], nil
		}
	}
	return "", "", errors.New("unable to find proper image in meatadata")
}

func checkIfCacheExistsAndLatest(dst string, latestSha string) (bool, error) {
	// does a cache image exist
	if _, err := os.Stat(dst); err == nil {
		// is it the latest
		f, err := os.Open(dst)
		if err != nil {
			log.Fatal(err)
		}
		defer f.Close()

		h := sha256.New()
		_, err = io.Copy(h, f)
		if err != nil {
			return false, err
		}
		if hex.EncodeToString(h.Sum(nil)) == latestSha {
			// shas are same, return
			return true, nil
		}
		return false, fmt.Errorf("an old cache file exists at %q: remove it and rerun", dst)
	}
	return false, nil
}

func getTestImage() (string, error) {
	fedoraMetaData, err := get(fmt.Sprintf("%s/metadata/images.json", fedoraBaseDirEndpoint))
	if err != nil {
		Fail(fmt.Sprintf("unable to determine fedora cloud version: %q", err))
	}
	var fcmd fedoraCloudMetadata
	if err := json.Unmarshal(fedoraMetaData, &fcmd); err != nil {
		Fail(err.Error())
	}

	downloadPath, upstreamSha, err := fcmd.downloadPathForVHD()
	if err != nil {
		Fail(err.Error())
	}
	fileName := filepath.Base(downloadPath)
	dstPath := filepath.Join(defaultCacheDirPath, fileName)

	// check for cached image && if cached image is latest
	exists, err := checkIfCacheExistsAndLatest(dstPath, upstreamSha)
	if err != nil {
		Fail(err.Error())
	}
	if err != nil {
		Fail(err.Error())
	}
	if !exists {
		dst, err := os.OpenFile(dstPath, os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			return "", err
		}
		defer dst.Close()
		if err := pullWithProgress(downloadPath, dst); err != nil {
			return "", err
		}
	}
	// Decompress the downloaded vhdfixed.xz file
	vhdPath := filepath.Join(defaultCacheDirPath, filepath.Base(downloadPath[:len(downloadPath)-len(".vhdfixed.xz")])) + ".vhd"
	err = decompressVhdXZ(dstPath, vhdPath)
	if err != nil {
		return "", err
	}
	return vhdPath, nil
}

func decompressVhdXZ(src string, dst string) error {
	f, err := os.Open(src)
	if err != nil {
		return err
	}
	defer f.Close()

	r, err := xz.NewReader(f)
	if err != nil {
		return fmt.Errorf("xz decompression failed: %v", err)
	}

	// Directly copy the decompressed data to the destination file
	outFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer outFile.Close()

	_, err = io.Copy(outFile, r)
	if err != nil {
		return err
	}
	return nil
}

func archFromGOOS() string {
	if runtime.GOOS == "arm64" {
		return "aarch64"
	}
	return "x86_64"
}
