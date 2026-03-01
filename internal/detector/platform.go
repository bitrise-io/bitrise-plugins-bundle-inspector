package detector

// Platform represents the target platform for platform-aware detectors
type Platform string

const (
	// PlatformIOS represents iOS artifacts (IPA, .app, XCArchive)
	PlatformIOS Platform = "ios"
	// PlatformAndroid represents Android artifacts (APK, AAB)
	PlatformAndroid Platform = "android"
)
