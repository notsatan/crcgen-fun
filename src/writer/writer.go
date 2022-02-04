/*
Package writer handles the part of writing output to a file

Use the function Start to initialize the package
*/
package writer

import (
	"fmt"
	"path/filepath"
	"strings"
	"sync"

	"github.com/pkg/errors"
	"github.com/spf13/viper"

	"github.com/notsatan/crcgen/src/logger"
)

const pkgName = "writer"

var once sync.Once

// output is the central instance of `viper` being used by package `writer`
var output = viper.New()

var readViperConfigs = (*viper.Viper).ReadInConfig // maps to viper.ReadInConfig

var (
	errInvalidFile = fmt.Errorf("(%s): could not detect output file in path", pkgName)
	errInvalidExt  = fmt.Errorf("(%s): output file has invalid extension", pkgName)
)

/*
IsInvalidFileErr checks if an error returned by package writer was caused because the
output file could not be located from the path
*/
func IsInvalidFileErr(err error) bool {
	return errors.Is(err, errInvalidFile)
}

/*
IsInvalidExtErr checks if an error returned by package `writer` was caused by the file
having an invalid extension
*/
func IsInvalidExtErr(err error) bool {
	return errors.Is(err, errInvalidExt)
}

/*
Start initializes package `writer` - should be executed before calls to any function
from this package

Returns an error if the output file could not be parsed from the path, if the output
file contains an invalid extension, or if function Start is being called multiple times
Use the functions IsInvalidFileErr, and IsInvalidExtErr to explicitly check for these
errors

Note: This function can be run once
*/
func Start(confPath string) error {
	var err error
	once.Do(func() {
		err = start(confPath)
	})

	return err
}

/*
start initializes package `writer`, setting up the configurations needed
*/
func start(confFile string) error {
	const logTag = "(" + pkgName + "/Start)"

	// Split path to get directory, filename and extension (remove `dot` from ext)
	_, file := filepath.Split(confFile)
	ext := strings.ToLower(strings.TrimLeft(filepath.Ext(confFile), "."))

	switch {
	case file == "":
		logger.Errorf(`%s: config path is empty`, logTag)
		return errors.Wrap(errInvalidFile, logTag)

	case ext == "":
		logger.Errorf(`%s: config path has no extension: "%s"`, logTag, confFile)
		return errors.Wrap(errInvalidExt, logTag)

	case !validateExt(ext):
		logger.Errorf(`%s: config file invalid ext detected: "%s"`, logTag, confFile)
		return errors.Wrap(errInvalidExt, logTag)
	}

	output.SetConfigFile(confFile)

	err := readViperConfigs(output)
	return errors.Wrap(err, logTag)
}

/*
validateExt validates if an extension is valid for the output file
*/
func validateExt(ext string) bool {
	switch strings.ToLower(ext) {
	case "json", "yaml", "yml":
		return true

	default:
		return false
	}
}