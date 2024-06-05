package main

import (
	"MyGoPj/dism"
	"MyGoPj/vhd"
)

func main() {
	testDisamAPI()
}

// CreateVhdx(path string, maxSizeInGb, blockSizeInMb uint32)
func testCreateVHD() {
	vhd.CreateVhdx("C:\\Users\\localuser\\Desktop\\test.vhdx", 1, 1)
}

func testDisamAPI() {
	dismSession, err := dism.OpenSession(dism.DISM_ONLINE_IMAGE,
		"",
		"",
		dism.DismLogErrorsWarningsInfo,
		"",
		"")

	if err != nil {
		panic(err)
	}
	defer dismSession.Close()

	err = dismSession.GetFeatures()

}
