package e2e

import (
	"fmt"
	"github.com/containers/common/pkg/strongunits"
	"github.com/containers/libhvee/pkg/hypervctl"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var (
	err error
	tvm *testVM
)
var _ = Describe("Disk tests", func() {
	BeforeEach(func() {
		tvm, err = newDefaultVM()
		if err != nil {
			Fail("unable to create vm")
		}
	})

	AfterEach(func() {
		if err := tvm.stopAndRemove(); err != nil {
			fmt.Printf("unable to complete teardown: %q\n", err)
		}
	})

	It("Disk resize", func() {
		// The VM should have been made with the defaultDiskSize.  Check the resulting VHD
		// and compare for equality. Comparision is done in the MB level due to slight rounding
		// of bytes between go and hyperv.
		defaultDiskSizeinGib := strongunits.GiB(defaultDiskSize)
		s, err := hypervctl.GetDiskSize(tvm.config.DiskPath)
		Expect(err).To(BeNil())
		Expect(strongunits.ToMib(s)).To(Equal(strongunits.ToMib(defaultDiskSizeinGib)))

		// default vm disk size is set to 10GB
		// test will change that to 15
		newDiskSize := strongunits.GiB(15)
		err = hypervctl.ResizeDisk(tvm.config.DiskPath, newDiskSize)
		Expect(err).To(BeNil())
		if err := tvm.refresh(); err != nil {
			Fail(err.Error())
		}

		// check again after resizing
		newSize, err := hypervctl.GetDiskSize(tvm.config.DiskPath)
		Expect(err).To(BeNil())
		Expect(strongunits.ToMib(newSize)).To(Equal(strongunits.ToMib(newDiskSize)))

	})
})
