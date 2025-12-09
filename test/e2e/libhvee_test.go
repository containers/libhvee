package e2e

import (
	"fmt"

	"io"
	"net/http"
	"os"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/schollz/progressbar/v3"
)

var (
	cachedImagePath string
)

func TestMain(m *testing.M) {
	os.Exit(m.Run())
}

func TestLibhvee(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Libhvee Suite")
}

func get(endpoint string) ([]byte, error) {

	getReq, err := http.NewRequest("GET", endpoint, nil)
	if err != nil {
		return nil, err
	}
	addHeaders(getReq)

	resp, err := http.DefaultClient.Do(getReq)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		Fail(fmt.Sprintf("get %s: status code: %d", endpoint, resp.StatusCode))
	}
	defer func() {
		_ = resp.Body.Close()
	}()
	body, err := io.ReadAll(resp.Body)
	return body, err
}

// these headers are required to bypass anubis bot protection
// maybe subject to change if the bot protection is updated
func addHeaders(getReq *http.Request) {
	getReq.Header.Set("Accept", "application/json")
	getReq.Header.Set("Accept-Encoding", "gzip, deflate, br")
	getReq.Header.Set("Accept-Language", "en-US,en;q=0.9")
	getReq.Header.Set("Connection", "keep-alive")
	getReq.Header.Set("Referer", "https://kojipkgs.fedoraproject.org/compose/cloud/latest-Fedora-Cloud-41/compose")
	getReq.Header.Set("Sec-Ch-Ua-Mobile", "?0")
	getReq.Header.Set("Sec-Ch-Ua-Platform", "macOS")
	getReq.Header.Set("Sec-Fetch-Dest", "document")
	getReq.Header.Set("Sec-Fetch-Mode", "navigate")
	getReq.Header.Set("Sec-Fetch-Site", "same-origin")
	getReq.Header.Set("Sec-User", "?1")
	getReq.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/142.0.0.0 Safari/537.36")
}

func pullWithProgress(endpoint string, dst *os.File) error {
	fmt.Println("trying to pull: ", endpoint)
	req, err := http.NewRequest("GET", endpoint, nil)
	if err != nil {
		Fail(err.Error())
	}
	addHeaders(req)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		Fail(err.Error())
	}
	if resp.StatusCode != http.StatusOK {
		Fail(fmt.Sprintf("get %s: status code: %d", endpoint, resp.StatusCode))
	}
	defer func() {
		_ = resp.Body.Close()
	}()
	return doWithProgress("downloading", resp.Body, dst, resp.ContentLength)
}

func copyWithProgress(srcPath string, dst *os.File) error {
	s, err := os.Stat(srcPath)
	if err != nil {
		return err
	}
	f, err := os.Open(srcPath)
	if err != nil {
		return err
	}
	defer func() {
		_ = f.Close()
	}()
	return doWithProgress("copying", f, dst, s.Size())
}

func doWithProgress(status string, src io.Reader, dst io.Writer, maxBytes int64) error {
	bar := progressbar.DefaultBytes(
		maxBytes,
		status,
	)
	_, err := io.Copy(io.MultiWriter(dst, bar), src)
	return err

}

var _ = BeforeSuite(func() {
	// placeholder for stuff
	fmt.Println("Before")
	var err error
	cachedImagePath, err = getTestImage()
	if err != nil {
		Fail(err.Error())
	}
	fmt.Println(cachedImagePath)

})

var _ = AfterSuite(func() {
	// placeholder for stuff
	fmt.Println("After")

})

var (
	_ = BeforeEach(func() {
	})

	_ = AfterEach(func() {
		// we can do stuff after
	})
)
