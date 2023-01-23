package local

import (
	"fmt"
	"testing"
)

func TestPrettyPrintFileSizes(t *testing.T) {
	fmt.Println(prettyPrintFileSizes(100))
	fmt.Println(prettyPrintFileSizes(1000))
	fmt.Println(prettyPrintFileSizes(10000))
	fmt.Println(prettyPrintFileSizes(100000))
	fmt.Println(prettyPrintFileSizes(1000000))
	fmt.Println(prettyPrintFileSizes(10000000))
	fmt.Println(prettyPrintFileSizes(100000000))
	fmt.Println(prettyPrintFileSizes(987000000))
	fmt.Println(prettyPrintFileSizes(1000000000))
}
