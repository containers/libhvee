package e2e

import (
	"fmt"
	"time"

	"github.com/containers/libhvee/pkg/hypervctl"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var (
	noDefer bool
)

func removeOnError(t *testVM) {
	if noDefer {
		return
	}
	err := t.stopAndRemove()
	if err != nil {
		fmt.Println(err)
	}
}

var _ = Describe("basic operation test", func() {

	BeforeEach(func() {
		// setup stuff can go here
	})

	AfterEach(func() {
		// teardown stuff can go here
	})

	It("create, start, stop, remove", func() {
		// create a basic vm
		tvm, err := newDefaultVM()
		Expect(err).To(BeNil())

		defer removeOnError(tvm)

		Expect(tvm.vm.IsStarting()).To(BeFalse())
		Expect(tvm.vm.State).To(Equal(hypervctl.Disabled))

		// start the vm
		err = tvm.vm.Start()
		Expect(err).To(BeNil())

		err = tvm.refresh()
		Expect(err).To(BeNil())
		Expect(tvm.vm.State).To(Equal(hypervctl.Enabled))

		// TODO I get an error when trying to immediately stop a VM so this is a placeholder
		// for "wait"
		time.Sleep(20 * time.Second)

		// stop the vm
		err = tvm.vm.StopWithForce()
		Expect(err).To(BeNil())
		err = tvm.refresh()
		Expect(err).To(BeNil())
		Expect(tvm.vm.State).To(Equal(hypervctl.Disabled))

		// remove the vm
		err = tvm.vm.Remove(tvm.config.DiskPath)
		Expect(err).To(BeNil())
		noDefer = true
	})
})
