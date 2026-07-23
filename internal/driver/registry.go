package driver

import "github.com/openforge-oss/anvil/internal/detect"

// For returns the driver for a detected stack. Dart is detectable but not
// buildable as an app, so it has no driver.
func For(stack detect.Stack, root string) (Driver, bool) {
	switch stack {
	case detect.Flutter:
		return Flutter{base{root}}, true
	case detect.ReactNative:
		return ReactNative{base{root}}, true
	case detect.Android:
		return Android{base{root}}, true
	case detect.IOS:
		return IOS{base{root}}, true
	case detect.Kotlin:
		return Kotlin{base{root}}, true
	case detect.Swift:
		return Swift{base{root}}, true
	default:
		return nil, false
	}
}
