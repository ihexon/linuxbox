package dism

import (
	"fmt"
	"testing"
)

func TestDism(t *testing.T) {
	dismSession, err := OpenSession(DISM_ONLINE_IMAGE,
		"",
		"",
		DismLogErrorsWarningsInfo,
		"",
		"")

	if err != nil {
		panic(err)
	}
	defer dismSession.Close()

	list, err := dismSession.GetFeatures()
	fmt.Printf("%v \n", list)
}
