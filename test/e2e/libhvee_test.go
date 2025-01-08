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
	resp, err := http.Get(endpoint)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		Fail(fmt.Sprintf("get %s: status code: %d", endpoint, resp.StatusCode))
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	return body, err
}

func pullWithProgress(endpoint string, dst *os.File) error {
	fmt.Println("trying to pull: ", endpoint)
	req, err := http.NewRequest("GET", endpoint, nil)
	if err != nil {
		Fail(err.Error())
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		Fail(err.Error())
	}
	if resp.StatusCode != http.StatusOK {
		Fail(fmt.Sprintf("get %s: status code: %d", endpoint, resp.StatusCode))
	}
	defer resp.Body.Close()
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
	defer f.Close()
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
