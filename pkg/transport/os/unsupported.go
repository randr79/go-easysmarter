//go:build !linux

package os

import "runtime"

type _ERROR_os_not_implemented bool

const _ _ERROR_os_not_implemented = runtime.GOOS
