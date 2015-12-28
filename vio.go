package vio

import (
	"fmt"
	"os"
	"reflect"
	"strconv"
	"strings"
	"time"

	"gopkg.in/ini.v1"
)

type Status int

const (
	Committed Status = iota
	Staged
)

type version struct {
	revision  string
	timestamp time.Time
}

func NewVersion(revision string) *version {
	fields := strings.Split(revision, "#")

	if len(fields) == 2 {
		i, err := strconv.ParseInt(strings.TrimSpace(fields[1]), 10, 64)
		if err != nil {
			panic(err)
		}
		return &version{revision: fields[0], timestamp: time.Unix(i, 0)}
	} else {
		return &version{revision: fields[0], timestamp: time.Now()}
	}
}

func ContainsVersion(vs []version, v *version) bool {
	for _, v_in := range vs {
		if reflect.DeepEqual(*v, v_in) {
			return true
		}
	}
	return false
}

func (v *version) String() string {
	return fmt.Sprintf("%s#%d", v.revision, v.timestamp.Unix())
}

type Backend interface {
	// inits backend in current directory
	Init() error

	// opens the backend
	Open() error

	// whether the backend has been initialized
	IsInitialized() bool

	// returns status of backend
	GetStatus() (Status, error)

	// checks out a commit
	Checkout(v *version) error

	// commits a version.
	Commit() (*version, error)

	// retrieves the string representation of the diff for a path
	Diff(v1 *version, v2 *version, path string) (string, error)

	// returns list of committed versions
	GetVersions() (versions []version, err error)
}

type AnError struct {
	Msg string
}

func (e AnError) Error() string {
	return e.Msg
}

func InstantiateBackend(opts *ini.File) (backend Backend, err error) {
	opts.Section("").Key("repo_path").SetValue(".")
	backendType := opts.Section("").Key("backend_type").Value()
	switch backendType {
	case "posix":
		backend, err = NewPosixBackend(opts)
	default:
		return nil, AnError{"unknown backend " + backendType}
	}
	return
}

func Init(snapsPath string, backend string) (err error) {
	if _, err = os.Stat(".vioconfig"); err == nil {
		return
	}

	opts := ini.Empty()
	opts.Section("").Key("repo_path").SetValue(".")
	opts.Section("").Key("snapshots_path").SetValue(snapsPath)
	opts.Section("").Key("backend_type").SetValue(backend)

	b, err := InstantiateBackend(opts)
	if err != nil {
		return
	}

	err = b.Init()
	if err != nil {
		return
	}

	opts.Section("").DeleteKey("repo_path")

	return opts.SaveTo(".vioconfig")
}

func Commit() (err error) {
	opts, err := ini.Load(".vioconfig")
	if err != nil {
		return
	}
	b, err := InstantiateBackend(opts)
	if err != nil {
		return
	}
	_, err = b.Commit()

	return
}
