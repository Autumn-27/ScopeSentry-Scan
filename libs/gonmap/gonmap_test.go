package gonmap

import (
	"fmt"
	"testing"
	"time"
)

func TestScan(t *testing.T) {
	var scanner = New()
	host1 := "192.168.123.25"
	port1 := 22
	_, response := scanner.ScanTimeout(host1, port1, time.Second*5)
	if response == nil {
		t.Errorf("Port[%d] not open\n", port1)
		return
	}
	if response.FingerPrint.CPE == "" {
		t.Errorf("No CPE info detected\n")
		return
	}
	fmt.Println(response.FingerPrint.CPE) //expected: cpe:/a:openbsd:openssh:9.6p1
}
