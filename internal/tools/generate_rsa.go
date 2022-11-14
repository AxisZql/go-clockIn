package tools

import (
	"github.com/dop251/goja"
	"go-clockIn/pkg/constant"
	"log"
)

func GenerateRsa(data string) (rsa string, err error) {
	firstKye, secondKey, thirdKey := "1", "2", "3"
	vm := goja.New()
	_, err = vm.RunString(constant.DesScript)
	if err != nil {
		log.Fatal(err, "des.js must be have some problem")
	}
	var strEnc func(string, string, string, string) string
	err = vm.ExportTo(vm.Get("strEnc"), &strEnc)
	if err != nil {
		log.Fatal(err, "strEnc cannot map from JavaScript to Go")
		return
	}
	rsa = strEnc(data, firstKye, secondKey, thirdKey)
	return
}
