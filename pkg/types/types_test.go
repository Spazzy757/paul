package types

import (
	"bytes"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLoadConfig(t *testing.T) {
	var paulConfig PaulConfig

	yamlFile, err := ioutil.ReadFile("../../PAUL.yaml")
	if err != nil {
		t.Errorf("yamlFile.Get err   #%v ", err)
	}

	paulConfig.LoadConfig(yamlFile)
	t.Run("Test Loading Config - OpenMessage", func(t *testing.T) {
		assert.NotEqual(t, paulConfig.PullRequests.OpenMessage, "")
	})
	t.Run("Test Loading Config - CatsEnabled", func(t *testing.T) {
		assert.NotEqual(t, paulConfig.PullRequests.CatsEnabled, false)
	})
	t.Run("Test Loading Config - DogsEnabled", func(t *testing.T) {
		assert.NotEqual(t, paulConfig.PullRequests.DogsEnabled, false)
	})

}

func TestLoadConfigFails(t *testing.T) {
	// Only run actual test in subprocess
	if os.Getenv("SUB_PROCESS") == "1" {
		var paulConfig PaulConfig
		var buf bytes.Buffer
		log.SetOutput(&buf)
		paulConfig.LoadConfig([]byte(`% ^ & HHH`))
		return
	}
	cmd := exec.Command(os.Args[0], "-test.run=TestLoadConfigFails")
	cmd.Env = append(os.Environ(), "SUB_PROCESS=1")
	err := cmd.Run()
	// Cast the error as *exec.ExitError and compare the result
	e, ok := err.(*exec.ExitError)
	expectedErrorString := "exit status 1"
	assert.Equal(t, true, ok)
	assert.Equal(t, expectedErrorString, e.Error())
}
