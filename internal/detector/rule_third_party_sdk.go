package detector

import (
	"strings"

	"github.com/bitrise-io/bitrise-plugins-bundle-inspector/pkg/types"
)

// ThirdPartySDKRule detects resources bundled within third-party SDKs
// These are not under developer control and should not be flagged as duplicates
type ThirdPartySDKRule struct {
	analyzer *PathAnalyzer
}

// NewThirdPartySDKRule creates a new third-party SDK detection rule
func NewThirdPartySDKRule() *ThirdPartySDKRule {
	return &ThirdPartySDKRule{
		analyzer: NewPathAnalyzer(""),
	}
}

// ID returns the rule identifier
func (r *ThirdPartySDKRule) ID() string {
	return "rule-7-third-party-sdk"
}

// Name returns the rule name
func (r *ThirdPartySDKRule) Name() string {
	return "Third-Party SDK Bundled Resources Detection"
}

// Well-known third-party SDK framework names
// This list covers the most common SDKs found in production iOS apps
var thirdPartySDKs = map[string]bool{
	// Google SDKs
	"GoogleMaps":                 true,
	"GoogleMapsBase":             true,
	"GoogleMapsCore":             true,
	"GoogleMapsM4B":              true,
	"GoogleMapsPlaces":           true,
	"GooglePlaces":               true,
	"GoogleUtilities":            true,
	"GoogleSignIn":               true,
	"GoogleAppMeasurement":       true,
	"GoogleDataTransport":        true,
	"GoogleToolboxForMac":        true,
	"GTMSessionFetcher":          true,
	"GTMAppAuth":                 true,
	"GULAppDelegateSwizzler":     true,
	"GULMethodSwizzler":          true,
	"GULNetwork":                 true,

	// Firebase SDKs
	"Firebase":                   true,
	"FirebaseCore":               true,
	"FirebaseCoreInternal":       true,
	"FirebaseAnalytics":          true,
	"FirebaseAuth":               true,
	"FirebaseDatabase":           true,
	"FirebaseDynamicLinks":       true,
	"FirebaseMessaging":          true,
	"FirebaseCrashlytics":        true,
	"FirebasePerformance":        true,
	"FirebaseRemoteConfig":       true,
	"FirebaseStorage":            true,
	"FirebaseFirestore":          true,
	"FirebaseInstallations":      true,
	"FirebaseABTesting":          true,
	"FirebaseInAppMessaging":     true,

	// Facebook SDKs
	"FBSDKCoreKit":               true,
	"FBSDKLoginKit":              true,
	"FBSDKShareKit":              true,
	"FacebookSDK":                true,
	"FBSDKCoreKit_Basics":        true,
	"FacebookCore":               true,
	"FacebookLogin":              true,
	"FacebookShare":              true,
	"FBAEMKit":                   true,
	"FBSDKGamingServicesKit":     true,

	// Networking
	"Alamofire":                  true,
	"AFNetworking":               true,
	"SocketRocket":               true,
	"Starscream":                 true,

	// Image Loading
	"SDWebImage":                 true,
	"Kingfisher":                 true,
	"PINRemoteImage":             true,
	"AlamofireImage":             true,
	"Nuke":                       true,

	// Analytics & Crash Reporting
	"Crashlytics":                true,
	"Fabric":                     true,
	"Amplitude":                  true,
	"Mixpanel":                   true,
	"Segment":                    true,
	"Flurry":                     true,
	"NewRelic":                   true,
	"Bugsnag":                    true,
	"Sentry":                     true,
	"AppCenter":                  true,

	// Payment SDKs
	"Stripe":                     true,
	"PayPal":                     true,
	"Braintree":                  true,
	"Square":                     true,

	// UI Libraries
	"Lottie":                     true,
	"Charts":                     true,
	"SnapKit":                    true,
	"Masonry":                    true,
	"Hero":                       true,
	"Material":                   true,
	"MaterialComponents":         true,
	"MBProgressHUD":              true,
	"SVProgressHUD":              true,
	"JGProgressHUD":              true,

	// Social & Communication
	"TwitterKit":                 true,
	"TwitterCore":                true,
	"Giphy":                      true,
	"Branch":                     true,
	"Intercom":                   true,
	"ZendeskSDK":                 true,
	"Twilio":                     true,
	"SendBird":                   true,
	"Stream":                     true,

	// Advertising
	"GoogleMobileAds":            true,
	"FBAudienceNetwork":          true,
	"AdSupport":                  true,
	"AppLovin":                   true,
	"Vungle":                     true,
	"Chartboost":                 true,
	"IronSource":                 true,
	"UnityAds":                   true,

	// Database & Storage
	"Realm":                      true,
	"RealmSwift":                 true,
	"FMDB":                       true,
	"YapDatabase":                true,

	// Utility Libraries
	"RxSwift":                    true,
	"RxCocoa":                    true,
	"PromiseKit":                 true,
	"Bolts":                      true,
	"SwiftyJSON":                 true,
	"ObjectMapper":               true,
	"R.swift":                    true,

	// Media & Video
	"AVFoundation":               true,
	"AVKit":                      true,
	"YPImagePicker":              true,
	"TOCropViewController":       true,
	"JWPlayer":                   true,
	"BitmovinPlayer":             true,

	// AR/VR
	"ARKit":                      true,
	"SceneKit":                   true,
	"Vuforia":                    true,

	// Testing (sometimes included in release)
	"Reveal":                     true,
	"FLEX":                       true,
	"InjectionIII":               true,

	// Other Popular SDKs
	"Adjust":                     true,
	"AppsFlyer":                  true,
	"Kochava":                    true,
	"Localytics":                 true,
	"Urban Airship":              true,
	"OneSignal":                  true,
	"CleverTap":                  true,
	"Braze":                      true,
	"mParticle":                  true,
	"Optimizely":                 true,
	"LaunchDarkly":               true,
	"Mapbox":                     true,
	"HERE":                       true,
}

// isThirdPartySDK checks if a framework name is a known third-party SDK
func (r *ThirdPartySDKRule) isThirdPartySDK(frameworkName string) bool {
	// Check exact match
	if thirdPartySDKs[frameworkName] {
		return true
	}

	// Check prefixes for Firebase, Google, FB, etc.
	prefixes := []string{
		"Firebase",
		"Google",
		"FB",
		"FIR",
		"GUL",
		"GTM",
	}

	for _, prefix := range prefixes {
		if strings.HasPrefix(frameworkName, prefix) {
			return true
		}
	}

	return false
}

// Evaluate checks if duplicate resources are bundled within third-party SDKs
func (r *ThirdPartySDKRule) Evaluate(dup types.DuplicateSet) FilterResult {
	// Must have at least 2 files
	if len(dup.Files) < 2 {
		return FilterResult{ShouldFilter: false}
	}

	// Count how many files are in third-party SDK frameworks
	thirdPartyCount := 0
	frameworkNames := make(map[string]bool)

	for _, file := range dup.Files {
		if r.analyzer.IsFrameworkPath(file) {
			frameworkName := r.analyzer.ExtractFrameworkName(file)
			if frameworkName != "" && r.isThirdPartySDK(frameworkName) {
				thirdPartyCount++
				frameworkNames[frameworkName] = true
			}
		}
	}

	// If ALL files are in third-party SDKs, filter them out
	if thirdPartyCount == len(dup.Files) {
		// Get list of SDK names for reason message
		sdkList := make([]string, 0, len(frameworkNames))
		for name := range frameworkNames {
			sdkList = append(sdkList, name)
		}

		reason := "Resources bundled in third-party SDK frameworks (not under developer control)"
		if len(sdkList) > 0 {
			reason = "Resources in third-party SDK: " + strings.Join(sdkList, ", ")
		}

		return FilterResult{
			ShouldFilter: true,
			Reason:       reason,
			RuleID:       r.ID(),
			Priority:     "",
		}
	}

	// If MOST files (>= 50%) are in third-party SDKs, also filter
	// This handles edge cases where some duplicates span SDK and app code
	if float64(thirdPartyCount)/float64(len(dup.Files)) >= 0.5 {
		return FilterResult{
			ShouldFilter: true,
			Reason:       "Majority of duplicates in third-party SDKs",
			RuleID:       r.ID(),
			Priority:     "",
		}
	}

	// Not a third-party SDK pattern
	return FilterResult{ShouldFilter: false}
}
