package workloads

import (
	"errors"
	"io/ioutil"
	"math/rand"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/nu7hatch/gouuid"
	"github.com/onsi/gomega/gbytes"
	. "github.com/pivotal-cf-experimental/cf-test-helpers/cf"
)

//Todo(simon) Remove, for dev testing only
func random(min, max int) int {
	rand.Seed(time.Now().UTC().UnixNano())
	r := min + rand.Intn(max-min)
	return r
}

func Dummy() error {
	time.Sleep(time.Duration(random(1, 5)) * time.Second)
	return nil
}

func DummyWithErrors() error {
	Dummy()
	if random(0, 10) > 8 {
		return errors.New("Random (dummy) error")
	}
	return nil
}

func Push() error {
	guid, _ := uuid.NewV4()
	pathToApp := path.Join("assets", "dora")
	return expectCfToSay("App Started", "push", "pats-"+guid.String(), "-m", "64M", "-p", pathToApp)
}

func CopyAndReplaceText(srcDir string, dstDir string, searchText string, replaceText string) error {
	return filepath.Walk(srcDir, func(file string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		pathTail := strings.SplitAfter(file, srcDir)[1]
		if info.IsDir() {
			err = os.Mkdir(path.Join(dstDir, pathTail), 0777)
			if err != nil {
				return err
			}
		} else if info.Mode().IsRegular() {
			input, err := ioutil.ReadFile(file)
			if err != nil {
				return err
			}
			inputString := strings.Replace(string(input), searchText, replaceText, -1)
			input = []byte(inputString)
			output, err := os.Create(path.Join(dstDir, pathTail))
			if err != nil {
				return err
			}
			defer output.Close()
			output.Write(input)
		}
		return err
	})
}

func GenerateAndPush() error {
	guid, _ := uuid.NewV4()
	srcDir := path.Join("assets", "dora")
	rand.Seed(time.Now().UTC().UnixNano())
	salt := strconv.FormatInt(rand.Int63(), 10)
	dstDir := path.Join(os.TempDir(), salt)
	defer os.RemoveAll(dstDir)
	err := CopyAndReplaceText(srcDir, dstDir, "$RANDOM_TEXT", salt)
	if err != nil {
		return err
	}

	return expectCfToSay("App Started", "push", "pats-"+guid.String(), "-m", "128M", "-p", dstDir)
}

func expectCfToSay(expect string, args ...string) error {
	success, err := gbytes.Say(expect).Match(Cf(args...))

	if success {
		return nil
	} else {
		return err
	}
}
