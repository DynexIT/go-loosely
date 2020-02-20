package main

import (
	"fmt"
	"github.com/google/uuid"
	. "github.com/otiai10/copy"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func main() {
	looseArgs, goArgs := sanitizeArgs(os.Args[1:]...)

	if len(goArgs) > 0 && goArgs[0] == "build" {
		executeBuildScript(looseArgs, goArgs)
	}

	executeShell("go", goArgs, nil)
}

func executeShell(command string, args []string, workingDirectory *string) {
	cmd := exec.Command(command, args...)
	if workingDirectory != nil {
		cmd.Dir = *workingDirectory
	}
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		panic(err.Error())
	}
}

func executeBuildScript(looseArgs []string, goArgs []string) {
	fmt.Println("Loose build starting...")

	workingDirectory, err := os.Getwd()
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}

	goPath := os.Getenv("GOPATH")
	if goPath == "" {
		fmt.Println("GOPATH not set - needs GOPATH to be set to build in non module mode")
		os.Exit(1)
	}

	tempDir, moduleName, deleteDirectory := createTemporaryDirectory(workingDirectory, goPath)
	defer deleteDirectory()

	fmt.Println("Running go mod download")
	executeShell("go", []string{"mod", "download"}, &tempDir)
	fmt.Println("Running go mod vendor")
	executeShell("go", []string{"mod", "vendor"}, &tempDir)

	_ = os.Remove(tempDir + "/go.mod")
	_ = os.Remove(tempDir + "/go.sum")

	moduleNames := getLooseArg(looseArgs, "module")
	if len(moduleNames) > 0 {
		fmt.Println("Moving vendors/modules")
		replace := moduleNames[0]
		replaceModuleName(tempDir, replace, moduleName)
	}

	moveVendors := getLooseArg(looseArgs, "moveVendor")
	if len(moveVendors) > 0 {
		fmt.Println("Moving vendors")
		for _, mv := range moveVendors {
			s := strings.Split(mv, ":")
			if len(s) != 2 {
				panic("vendor move incorrect format")
			}
			moveVendor(tempDir, s[0], s[1])
		}
	}

	fmt.Println("Building module")
	executeShell("env", append([]string{"GO111MODULE=off", "go"}, goArgs...), &tempDir)

	if copies := getLooseArg(looseArgs, "copy"); len(copies) > 0 {
		for _, fileName := range copies {
			copyBuiltFile(tempDir, fileName, workingDirectory)
		}
	}
	deleteDirectory()
	os.Exit(0)
}

func moveVendor(dir string, from string, to string) {
	if err := Copy(dir + "/vendor/" + from, dir + "/vendor/" + to); err != nil {
		panic(err.Error())
	}
	replaceModuleName(dir, from, to)
}

func replaceModuleName(dir string, replace string, with string) {
	filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if info.IsDir() {
			return nil
		}
		read, err := ioutil.ReadFile(path)
		if err != nil {
			return nil
		}

		newContents := strings.Replace(string(read), replace, with, -1)

		err = ioutil.WriteFile(path, []byte(newContents), 0)
		if err != nil {
			return nil
		}
		return nil
	})
}

func createTemporaryDirectory(workingDirectory string, goPath string) (temporaryDirectory string, moduleName string, deleteDirectoryFunction func()) {
	s := strings.Split(workingDirectory, "/")
	projectName := s[len(s)-1]
	moduleName = projectName + "-" + uuid.New().String()
	temporaryDirectory = goPath + "/src/" + moduleName
	fmt.Println("Copying to temporary directory:", temporaryDirectory)
	if err := Copy(workingDirectory, temporaryDirectory); err != nil {
		panic(err.Error())
	}
	deleteDirectoryFunction = func() {
		fmt.Println("Deleting temporary directory")
		err := os.RemoveAll(temporaryDirectory)
		if err != nil {
			fmt.Println(err.Error())
		}
	}
	return
}

func copyBuiltFile(directory, fileName, destination string) {
	if err := Copy(directory+"/"+fileName, destination+"/"+fileName); err != nil {
		panic(err.Error())
	}
}

func sanitizeArgs(args ...string) (looseArgs []string, goArgs []string) {
	for _, arg := range args {
		if strings.HasPrefix(arg, "~") {
			looseArgs = append(looseArgs, arg[1:])
		} else {
			goArgs = append(goArgs, arg)
		}
	}
	return
}

func getLooseArg(args []string, key string) []string {
	var o []string
	for _, arg := range args {
		split := strings.Split(arg, "=")
		if len(split) == 1 && split[0] == key {
			o = append(o, "true")
		} else if split[0] == key {
			o = append(o, strings.Join(split[1:], "="))
		}
	}
	return o
}
