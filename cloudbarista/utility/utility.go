package common

import (
	uuid "github.com/google/uuid"
	"os/exec"
	"runtime"
)

func SysCall(cmdStr string) (string, error) {
	cmd := exec.Command("bash", "-c", cmdStr)

	cmdOut, err := cmd.Output()

	if err != nil {
		//panic(err)
		return "nil", err
	}
	return string(cmdOut), nil

}

func SysLookPath(cmdStr string) (string, error) {
	path, err := exec.LookPath(cmdStr)

	if err != nil {
		//panic(err)
		return "nil", err
	}
	return string(path), nil

}

func GenUuid() string {
	return uuid.New().String()
}

func GetNumCPU() int {
	return runtime.NumCPU()
}

const (
	METHOD_CREATE = "POST"
	METHOD_DELETE = "DELETE"
	METHOD_GET    = "GET"
	METHOD_PUT    = "PUT"
	METHOD_PATCH  = "PATCH"
	ON            = true
	OFF           = false
)