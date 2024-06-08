package shell

import (
	"fmt"
	"testing"
)

func TestShellExecute(t *testing.T) {
	DropAllPrivilegesExcept()
}

func TestExec(t *testing.T) {

	envs := []string{"TEST=1"}

	output, err := Exec(nil, envs, "dism", "test.txt")
	if err == nil {
		fmt.Println(string(output))
		return
	}
	fmt.Printf("%s", string(output))
	fmt.Println(err)

}
