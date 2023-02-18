//go:build windows
// +build windows

package wmiext

import (
	"github.com/go-ole/go-ole"
	"github.com/sirupsen/logrus"
	"golang.org/x/sys/windows"
)

var modoleaut32 = windows.NewLazySystemDLL("oleaut32.dll")
var procSafeArrayCreateVector = modoleaut32.NewProc("SafeArrayCreateVector")
var procSafeArrayPutElement = modoleaut32.NewProc("SafeArrayPutElement")
var procSafeArrayDestroy = modoleaut32.NewProc("SafeArrayDestroy")
var clsidWbemObjectTextSrc = ole.NewGUID("{8d1c559d-84f0-4bb3-a7d5-56a7435a9ba6}")
var iidIWbemObjectTextSrc = ole.NewGUID("{bfbf883a-cad7-11d3-a11b-00105a1f515a}")
var wmiWbemTxtLocator *ole.IUnknown

func init() {
	var err error
	// IID_IWbemObjectTextSrc Obtain the initial locator to WMI
	wmiWbemTxtLocator, err = ole.CreateInstance(clsidWbemObjectTextSrc, iidIWbemObjectTextSrc)
	if err != nil {
		logrus.Errorf("Could not initialize Wbem components, WMI operations will likely fail %s", err.Error())
	}
}
