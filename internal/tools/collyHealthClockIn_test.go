package tools

import (
	"fmt"
	"testing"
)

func TestCollyHealthClockIn(t *testing.T) {
	for i := 0; i < 1; i++ {
		_, err := CollyHealthClockIn("", "")
		if err != nil {
			fmt.Println(err)
			continue
		}
		fmt.Println("打卡成功")
	}
}
