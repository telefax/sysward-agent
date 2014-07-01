package main

import (
	"fmt"
	"os"
	"os/exec"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

type TestRunner struct{}

func (r TestRunner) Run(command string, args ...string) ([]byte, error) {
	cs := []string{"-test.run=TestHelperProcess", "--"}
	cs = append(cs, command)
	cs = append(cs, args...)
	cmd := exec.Command(os.Args[0], cs...)
	cmd.Env = []string{"GO_WANT_HELPER_PROCESS=1"}
	out, err := cmd.CombinedOutput()
	return out, err
}

func Setup() {
	runner = TestRunner{}
}

func TestPackagesThatNeedUpdates(t *testing.T) {
	Setup()
	Convey("Given pending updates", t, func() {

		Convey("There should be a list of packages available for update", func() {
			osPackages := buildPackageList()
			So(osPackages[0].Name, ShouldEqual, "apt")
			So(osPackages[0].Security, ShouldEqual, true)
		})
	})

}

func TestPackageUpdates(t *testing.T) {
	Setup()
	Convey("Given a package name", t, func() {

		Convey("The package should be upgraded", func() {
			err := updatePackage("apt")
			So(err, ShouldBeNil)
		})

		Convey("The package should not upgrade if held", func() {
			err := updatePackage("apt-held")
			fmt.Println(err)
			So(err, ShouldNotBeNil)
		})
	})

}

func TestPackageHolding(t *testing.T) {
	Setup()
	Convey("Given holding a package", t, func() {

		Convey("The package should be held", func() {
			err := holdPackage("apt")
			So(err, ShouldBeNil)
		})
	})

	Convey("Given unholding a package", t, func() {

		Convey("The package should be unheld", func() {
			err := unholdPackage("apt-held")
			So(err, ShouldBeNil)
		})
	})

}

func TestSourceList(t *testing.T) {
	Setup()
	Convey("Given /etc/apt/sources.list exists", t, func() {

		Convey("There should be a source list", func() {
			sourcesList := getSourcesList()
			src_one := sourcesList[0]
			src_two := sourcesList[1]
			So(src_one.Url, ShouldEqual, "http://us.archive.ubuntu.com/ubuntu/")
			So(src_one.Src, ShouldBeFalse)
			So(src_two.Src, ShouldBeTrue)
		})
	})

}

func TestInstalledPackages(t *testing.T) {
	Setup()
	Convey("Given I want to view all installed packages", t, func() {

		Convey("It returns a list of all installed packages", func() {
			packages := buildInstalledPackageList()
			So(packages[0], ShouldEqual, "apt")
			So(len(packages), ShouldEqual, 1)
		})
	})

}

func TestUpdatingThePackageList(t *testing.T) {
	Setup()
	Convey("Given I want to have the latest source list", t, func() {

		Convey("apt-update gets run", func() {
			err := updatePackageLists()
			So(err, ShouldBeNil)
		})
	})

}

func TestUpdateCounts(t *testing.T) {
	Setup()
	Convey("Given there are security and regular updates", t, func() {

		Convey("The number of security and regular updates is > 0", func() {
			updates := updateCounts()
			So(updates.Regular, ShouldEqual, 1)
			So(updates.Security, ShouldEqual, 2)
		})

	})
}

func TestHelperProcess(*testing.T) {
	if os.Getenv("GO_WANT_HELPER_PROCESS") != "1" {
		return
	}
	cmd := os.Args[3]
	defer os.Exit(0)

	if cmd == "python" && os.Args[4] == "trex.py" {
		fmt.Println(`[{"name": "apt", "section": "admin", "priority": "important", "current_version": "1.0.1ubuntu2", "security": true, "candidate_version": "1.0.1ubuntu2.1"}]`)
	} else if cmd == "apt-get" && os.Args[4] == "install" {
		if os.Args[6] == "apt-held" {
			os.Exit(-1)
		}
	} else if cmd == "apt-mark" {
		// nothing

	} else if cmd == "apt-get" && os.Args[4] == "update" {
		// nothing
	} else if cmd == "grep" && os.Args[4] == "-h" {
		src := `deb http://us.archive.ubuntu.com/ubuntu/ trusty main restricted
deb-src http://us.archive.ubuntu.com/ubuntu/ trusty main restricted`
		fmt.Println(src)
	} else if cmd == "dpkg" && os.Args[4] == "--get-selections" {
		fmt.Println("apt\u0009install")
	} else if cmd == "/usr/lib/update-notifier/apt-check" {
		fmt.Println("1;2")
	} else {
		fmt.Println(os.Args)
		for index, arg := range os.Args {
			fmt.Println(fmt.Sprintf("arg[%d]: %s", index, string(arg)))
		}
		os.Exit(-1)
	}
}