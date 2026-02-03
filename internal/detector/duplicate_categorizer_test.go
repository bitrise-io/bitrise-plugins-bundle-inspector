package detector

import (
	"testing"

	"github.com/bitrise-io/bitrise-plugins-bundle-inspector/pkg/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestInfoPlistRule(t *testing.T) {
	rule := NewInfoPlistRule()

	tests := []struct {
		name             string
		files            []string
		wantShouldFilter bool
		wantReason       string
	}{
		{
			name: "Info.plist in different frameworks - should filter",
			files: []string{
				"Payload/App.app/Frameworks/GoogleMaps.framework/Info.plist",
				"Payload/App.app/Frameworks/Firebase.framework/Info.plist",
			},
			wantShouldFilter: true,
			wantReason:       "Info.plist files in different bundles (required by iOS architecture)",
		},
		{
			name: "Info.plist in framework and app - should filter",
			files: []string{
				"Payload/App.app/Info.plist",
				"Payload/App.app/Frameworks/SDK.framework/Info.plist",
			},
			wantShouldFilter: true,
			wantReason:       "Info.plist files in different bundles (required by iOS architecture)",
		},
		{
			name: "Info.plist in framework and extension - should filter",
			files: []string{
				"Payload/App.app/Frameworks/SDK.framework/Info.plist",
				"Payload/App.app/PlugIns/Share.appex/Info.plist",
			},
			wantShouldFilter: true,
			wantReason:       "Info.plist files in different bundles (required by iOS architecture)",
		},
		{
			name: "Info.plist in same directory - actionable",
			files: []string{
				"Payload/App.app/Info.plist",
				"Payload/App.app/Info.plist.backup",
			},
			wantShouldFilter: false,
		},
		{
			name: "Not Info.plist files - skip",
			files: []string{
				"Payload/App.app/config.plist",
				"Payload/App.app/Frameworks/SDK.framework/config.plist",
			},
			wantShouldFilter: false,
		},
		{
			name: "Single file - skip",
			files: []string{
				"Payload/App.app/Info.plist",
			},
			wantShouldFilter: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dup := types.DuplicateSet{
				Files: tt.files,
				Count: len(tt.files),
				Size:  1024,
			}

			result := rule.Evaluate(dup)
			assert.Equal(t, tt.wantShouldFilter, result.ShouldFilter, "ShouldFilter mismatch")
			if tt.wantShouldFilter {
				assert.Equal(t, tt.wantReason, result.Reason, "Reason mismatch")
				assert.Equal(t, rule.ID(), result.RuleID, "RuleID mismatch")
			}
		})
	}
}

func TestNIBVariantsRule(t *testing.T) {
	rule := NewNIBVariantsRule()

	tests := []struct {
		name             string
		files            []string
		wantShouldFilter bool
	}{
		{
			name: "runtime.nib and objects.nib - should filter",
			files: []string{
				"Payload/App.app/Base.lproj/Main.nib/runtime.nib",
				"Payload/App.app/Base.lproj/Main.nib/objects-8.0+.nib",
			},
			wantShouldFilter: true,
		},
		{
			name: "Multiple objects.nib variants - should filter",
			files: []string{
				"Payload/App.app/Base.lproj/Main.nib/runtime.nib",
				"Payload/App.app/Base.lproj/Main.nib/objects-11.0+.nib",
				"Payload/App.app/Base.lproj/Main.nib/objects-13.0+.nib",
			},
			wantShouldFilter: true,
		},
		{
			name: "NIB files in different .nib directories - should filter",
			files: []string{
				"Payload/App.app/Base.lproj/ViewA.nib/runtime.nib",
				"Payload/App.app/Base.lproj/ViewB.nib/runtime.nib",
			},
			wantShouldFilter: true,
		},
		{
			name: "Same NIB in same directory (true duplicate) - actionable",
			files: []string{
				"Payload/App.app/Base.lproj/Main.nib/runtime.nib",
				"Payload/App.app/Base.lproj/Main.nib/runtime.nib.backup",
			},
			wantShouldFilter: false,
		},
		{
			name: "Not NIB files - skip",
			files: []string{
				"Payload/App.app/file1.txt",
				"Payload/App.app/file2.txt",
			},
			wantShouldFilter: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dup := types.DuplicateSet{
				Files: tt.files,
				Count: len(tt.files),
				Size:  2048,
			}

			result := rule.Evaluate(dup)
			assert.Equal(t, tt.wantShouldFilter, result.ShouldFilter, "ShouldFilter mismatch")
		})
	}
}

func TestContentsJSONRule(t *testing.T) {
	rule := NewContentsJSONRule()

	tests := []struct {
		name             string
		files            []string
		wantShouldFilter bool
	}{
		{
			name: "Contents.json in different asset sets - should filter",
			files: []string{
				"Payload/App.app/Assets.xcassets/AppIcon.appiconset/Contents.json",
				"Payload/App.app/Assets.xcassets/LaunchImage.launchimage/Contents.json",
			},
			wantShouldFilter: true,
		},
		{
			name: "Contents.json in xcassets - should filter",
			files: []string{
				"Payload/App.app/Assets.xcassets/Icon.imageset/Contents.json",
				"Payload/App.app/Assets.xcassets/Color.colorset/Contents.json",
			},
			wantShouldFilter: true,
		},
		{
			name: "Contents.json in Assets.car extraction - should filter",
			files: []string{
				"Payload/App.app/Assets.car/AppIcon/Contents.json",
				"Payload/App.app/Assets.car/LaunchImage/Contents.json",
			},
			wantShouldFilter: true,
		},
		{
			name: "Contents.json not in asset catalog - actionable",
			files: []string{
				"Payload/App.app/Contents.json",
				"Payload/App.app/Config/Contents.json",
			},
			wantShouldFilter: false,
		},
		{
			name: "Not Contents.json - skip",
			files: []string{
				"Payload/App.app/config.json",
				"Payload/App.app/data.json",
			},
			wantShouldFilter: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dup := types.DuplicateSet{
				Files: tt.files,
				Count: len(tt.files),
				Size:  512,
			}

			result := rule.Evaluate(dup)
			assert.Equal(t, tt.wantShouldFilter, result.ShouldFilter, "ShouldFilter mismatch")
		})
	}
}

func TestDuplicateCategorizer_FilterDuplicates(t *testing.T) {
	categorizer := NewDuplicateCategorizer()

	duplicates := []types.DuplicateSet{
		{
			Files: []string{
				"Payload/App.app/Frameworks/A.framework/Info.plist",
				"Payload/App.app/Frameworks/B.framework/Info.plist",
			},
			Count:       2,
			Size:        1024,
			WastedSize:  1024,
			Hash:        "abc123",
		},
		{
			Files: []string{
				"Payload/App.app/Assets.xcassets/AppIcon.appiconset/Contents.json",
				"Payload/App.app/Assets.xcassets/LaunchImage.launchimage/Contents.json",
			},
			Count:       2,
			Size:        512,
			WastedSize:  512,
			Hash:        "def456",
		},
		{
			Files: []string{
				"Payload/App.app/audio.mp3",
				"Payload/App.app/Resources/audio.mp3",
			},
			Count:       2,
			Size:        2097152, // 2 MB
			WastedSize:  2097152,
			Hash:        "ghi789",
		},
	}

	actionable, filtered := categorizer.FilterDuplicates(duplicates)

	// Should filter out Info.plist and Contents.json (2 sets)
	assert.Equal(t, 2, len(filtered), "Expected 2 filtered duplicates")

	// Should keep audio.mp3 (1 set)
	assert.Equal(t, 1, len(actionable), "Expected 1 actionable duplicate")
	assert.Contains(t, actionable[0].Files[0], "audio.mp3", "Expected audio.mp3 to be actionable")
}

func TestDuplicateCategorizer_Categorize(t *testing.T) {
	categorizer := NewDuplicateCategorizer()

	duplicates := []types.DuplicateSet{
		{
			Files: []string{
				"Payload/App.app/Frameworks/A.framework/Info.plist",
				"Payload/App.app/Frameworks/B.framework/Info.plist",
			},
			Count: 2,
			Size:  1024,
		},
		{
			Files: []string{
				"Payload/App.app/logo.png",
				"Payload/App.app/Resources/logo.png",
			},
			Count: 2,
			Size:  102400, // 100 KB
		},
	}

	result := categorizer.Categorize(duplicates)

	assert.Equal(t, 2, result.TotalCount, "Total count should be 2")
	assert.Equal(t, 1, len(result.Actionable), "Should have 1 actionable duplicate (logo.png)")
	assert.Equal(t, 1, len(result.Filtered), "Should have 1 filtered duplicate (Info.plist)")
}

func TestRuleRegistry_Evaluate(t *testing.T) {
	registry := NewRuleRegistry()

	tests := []struct {
		name             string
		dup              types.DuplicateSet
		wantShouldFilter bool
		wantRuleID       string
	}{
		{
			name: "Info.plist rule matches",
			dup: types.DuplicateSet{
				Files: []string{
					"Payload/App.app/Frameworks/A.framework/Info.plist",
					"Payload/App.app/Frameworks/B.framework/Info.plist",
				},
				Count: 2,
			},
			wantShouldFilter: true,
			wantRuleID:       "rule-1-info-plist",
		},
		{
			name: "NIB variants rule matches",
			dup: types.DuplicateSet{
				Files: []string{
					"Payload/App.app/Base.lproj/Main.nib/runtime.nib",
					"Payload/App.app/Base.lproj/Main.nib/objects-8.0+.nib",
				},
				Count: 2,
			},
			wantShouldFilter: true,
			wantRuleID:       "rule-2-nib-variants",
		},
		{
			name: "Contents.json rule matches",
			dup: types.DuplicateSet{
				Files: []string{
					"Payload/App.app/Assets.xcassets/AppIcon.appiconset/Contents.json",
					"Payload/App.app/Assets.xcassets/LaunchImage.launchimage/Contents.json",
				},
				Count: 2,
			},
			wantShouldFilter: true,
			wantRuleID:       "rule-3-contents-json",
		},
		{
			name: "No rule matches - default actionable",
			dup: types.DuplicateSet{
				Files: []string{
					"Payload/App.app/image.png",
					"Payload/App.app/Resources/image.png",
				},
				Count: 2,
			},
			wantShouldFilter: false,
			wantRuleID:       "default",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := registry.Evaluate(tt.dup)
			assert.Equal(t, tt.wantShouldFilter, result.ShouldFilter, "ShouldFilter mismatch")
			assert.Equal(t, tt.wantRuleID, result.RuleID, "RuleID mismatch")
		})
	}
}

func TestDuplicateCategorizer_EvaluateDuplicate(t *testing.T) {
	categorizer := NewDuplicateCategorizer()

	// Test architectural pattern (should be filtered)
	archDup := types.DuplicateSet{
		Files: []string{
			"Payload/App.app/Frameworks/A.framework/Info.plist",
			"Payload/App.app/Frameworks/B.framework/Info.plist",
		},
		Count: 2,
		Size:  1024,
	}

	result := categorizer.EvaluateDuplicate(archDup)
	assert.True(t, result.ShouldFilter, "Architectural pattern should be filtered")
	assert.NotEmpty(t, result.Reason, "Reason should be provided")
	assert.NotEmpty(t, result.RuleID, "RuleID should be provided")

	// Test actionable duplicate (should not be filtered)
	actionableDup := types.DuplicateSet{
		Files: []string{
			"Payload/App.app/logo.png",
			"Payload/App.app/Resources/logo.png",
		},
		Count: 2,
		Size:  102400,
	}

	result = categorizer.EvaluateDuplicate(actionableDup)
	assert.False(t, result.ShouldFilter, "Actionable duplicate should not be filtered")
}

func TestRuleRegistry_Register(t *testing.T) {
	registry := &RuleRegistry{
		rules: make([]Rule, 0),
	}

	// Start with no rules
	assert.Equal(t, 0, len(registry.GetRules()))

	// Register a rule
	rule := NewInfoPlistRule()
	registry.Register(rule)

	// Should have 1 rule
	assert.Equal(t, 1, len(registry.GetRules()))
}

func TestNewRuleRegistry(t *testing.T) {
	registry := NewRuleRegistry()

	// Should have 7 default rules (PR 1: Rules 1-3, PR 2: Rules 4-7)
	rules := registry.GetRules()
	require.Equal(t, 7, len(rules), "Should have 7 default rules")

	// Verify rule IDs
	ruleIDs := make(map[string]bool)
	for _, rule := range rules {
		ruleIDs[rule.ID()] = true
	}

	assert.True(t, ruleIDs["rule-1-info-plist"], "Should have Info.plist rule")
	assert.True(t, ruleIDs["rule-2-nib-variants"], "Should have NIB variants rule")
	assert.True(t, ruleIDs["rule-3-contents-json"], "Should have Contents.json rule")
	assert.True(t, ruleIDs["rule-4-localization"], "Should have Localization rule")
	assert.True(t, ruleIDs["rule-5-framework-scripts"], "Should have Framework scripts rule")
	assert.True(t, ruleIDs["rule-6-framework-metadata"], "Should have Framework metadata rule")
	assert.True(t, ruleIDs["rule-7-third-party-sdk"], "Should have Third-party SDK rule")
}

func TestLocalizationRule(t *testing.T) {
	rule := NewLocalizationRule()

	tests := []struct {
		name             string
		files            []string
		wantShouldFilter bool
	}{
		{
			name: "Localization files in different .lproj bundles - should filter",
			files: []string{
				"Payload/App.app/en.lproj/Localizable.strings",
				"Payload/App.app/de.lproj/Localizable.strings",
			},
			wantShouldFilter: true,
		},
		{
			name: "Localization files in different frameworks with same locale - should filter",
			files: []string{
				"Payload/App.app/Frameworks/SDK.framework/en.lproj/Localizable.strings",
				"Payload/App.app/Frameworks/OtherSDK.framework/en.lproj/Localizable.strings",
			},
			wantShouldFilter: true,
		},
		{
			name: "Localization files in framework and app - should filter",
			files: []string{
				"Payload/App.app/en.lproj/Localizable.strings",
				"Payload/App.app/Frameworks/SDK.framework/en.lproj/Localizable.strings",
			},
			wantShouldFilter: true,
		},
		{
			name: "stringsdict files in different bundles - should filter",
			files: []string{
				"Payload/App.app/en.lproj/Plurals.stringsdict",
				"Payload/App.app/Frameworks/SDK.framework/en.lproj/Plurals.stringsdict",
			},
			wantShouldFilter: true,
		},
		{
			name: "Localization files in same directory - actionable",
			files: []string{
				"Payload/App.app/en.lproj/Localizable.strings",
				"Payload/App.app/en.lproj/Localizable.strings.backup",
			},
			wantShouldFilter: false,
		},
		{
			name: "Not localization files - skip",
			files: []string{
				"Payload/App.app/config.json",
				"Payload/App.app/Frameworks/SDK.framework/config.json",
			},
			wantShouldFilter: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dup := types.DuplicateSet{
				Files: tt.files,
				Count: len(tt.files),
				Size:  2048,
			}

			result := rule.Evaluate(dup)
			assert.Equal(t, tt.wantShouldFilter, result.ShouldFilter, "ShouldFilter mismatch")
		})
	}
}

func TestFrameworkScriptsRule(t *testing.T) {
	rule := NewFrameworkScriptsRule()

	tests := []struct {
		name             string
		files            []string
		wantShouldFilter bool
	}{
		{
			name: "strip-frameworks.sh in different frameworks - should filter",
			files: []string{
				"Payload/App.app/Frameworks/SDK1.framework/strip-frameworks.sh",
				"Payload/App.app/Frameworks/SDK2.framework/strip-frameworks.sh",
			},
			wantShouldFilter: true,
		},
		{
			name: "copy-frameworks.sh in frameworks - should filter",
			files: []string{
				"Payload/App.app/Frameworks/A.framework/copy-frameworks.sh",
				"Payload/App.app/Frameworks/B.framework/copy-frameworks.sh",
			},
			wantShouldFilter: true,
		},
		{
			name: "embed-frameworks.sh in different bundles - should filter",
			files: []string{
				"Payload/App.app/Frameworks/SDK.framework/embed-frameworks.sh",
				"Payload/App.app/PlugIns/Share.appex/embed-frameworks.sh",
			},
			wantShouldFilter: true,
		},
		{
			name: "Not a framework script - skip",
			files: []string{
				"Payload/App.app/Frameworks/SDK.framework/custom-script.sh",
				"Payload/App.app/Frameworks/OtherSDK.framework/custom-script.sh",
			},
			wantShouldFilter: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dup := types.DuplicateSet{
				Files: tt.files,
				Count: len(tt.files),
				Size:  4096,
			}

			result := rule.Evaluate(dup)
			assert.Equal(t, tt.wantShouldFilter, result.ShouldFilter, "ShouldFilter mismatch")
		})
	}
}

func TestFrameworkMetadataRule(t *testing.T) {
	rule := NewFrameworkMetadataRule()

	tests := []struct {
		name             string
		files            []string
		wantShouldFilter bool
	}{
		{
			name: ".supx files in different frameworks - should filter",
			files: []string{
				"Payload/App.app/Frameworks/Carthage.framework/Carthage.supx",
				"Payload/App.app/Frameworks/OtherLib.framework/OtherLib.supx",
			},
			wantShouldFilter: true,
		},
		{
			name: ".bcsymbolmap files - should filter",
			files: []string{
				"Payload/App.app/BCSymbolMaps/ABC123.bcsymbolmap",
				"Payload/App.app/BCSymbolMaps/DEF456.bcsymbolmap",
			},
			wantShouldFilter: true,
		},
		{
			name: "module.modulemap in different frameworks - should filter",
			files: []string{
				"Payload/App.app/Frameworks/SDK1.framework/Modules/module.modulemap",
				"Payload/App.app/Frameworks/SDK2.framework/Modules/module.modulemap",
			},
			wantShouldFilter: true,
		},
		{
			name: ".swiftmodule in frameworks - should filter",
			files: []string{
				"Payload/App.app/Frameworks/A.framework/A.swiftmodule",
				"Payload/App.app/Frameworks/B.framework/B.swiftmodule",
			},
			wantShouldFilter: true,
		},
		{
			name: "Not metadata file - skip",
			files: []string{
				"Payload/App.app/Frameworks/SDK.framework/config.json",
				"Payload/App.app/Frameworks/OtherSDK.framework/config.json",
			},
			wantShouldFilter: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dup := types.DuplicateSet{
				Files: tt.files,
				Count: len(tt.files),
				Size:  1024,
			}

			result := rule.Evaluate(dup)
			assert.Equal(t, tt.wantShouldFilter, result.ShouldFilter, "ShouldFilter mismatch")
		})
	}
}

func TestThirdPartySDKRule(t *testing.T) {
	rule := NewThirdPartySDKRule()

	tests := []struct {
		name             string
		files            []string
		wantShouldFilter bool
	}{
		{
			name: "GoogleMaps resources in different locations - should filter",
			files: []string{
				"Payload/App.app/Frameworks/GoogleMaps.framework/Resources/icon.png",
				"Payload/App.app/Frameworks/GoogleMaps.framework/Resources/tile.png",
				"Payload/App.app/Frameworks/GoogleMapsBase.framework/Resources/icon.png",
			},
			wantShouldFilter: true,
		},
		{
			name: "Firebase SDK resources - should filter",
			files: []string{
				"Payload/App.app/Frameworks/FirebaseCore.framework/GoogleService-Info.plist",
				"Payload/App.app/Frameworks/FirebaseAuth.framework/GoogleService-Info.plist",
			},
			wantShouldFilter: true,
		},
		{
			name: "Facebook SDK resources - should filter",
			files: []string{
				"Payload/App.app/Frameworks/FBSDKCoreKit.framework/Assets/icon.png",
				"Payload/App.app/Frameworks/FBSDKLoginKit.framework/Assets/icon.png",
			},
			wantShouldFilter: true,
		},
		{
			name: "Alamofire and AFNetworking (networking SDKs) - should filter",
			files: []string{
				"Payload/App.app/Frameworks/Alamofire.framework/logo.png",
				"Payload/App.app/Frameworks/AFNetworking.framework/logo.png",
			},
			wantShouldFilter: true,
		},
		{
			name: "SDWebImage resources - should filter",
			files: []string{
				"Payload/App.app/Frameworks/SDWebImage.framework/placeholder.png",
				"Payload/App.app/Frameworks/SDWebImageWebPCoder.framework/placeholder.png",
			},
			wantShouldFilter: true,
		},
		{
			name: "Stripe SDK resources - should filter",
			files: []string{
				"Payload/App.app/Frameworks/Stripe.framework/Assets/card.png",
				"Payload/App.app/Frameworks/StripeUICore.framework/Assets/card.png",
			},
			wantShouldFilter: true,
		},
		{
			name: "Mixed: SDK and app code - should filter (>50% SDK)",
			files: []string{
				"Payload/App.app/Frameworks/GoogleMaps.framework/icon.png",
				"Payload/App.app/Frameworks/Firebase.framework/icon.png",
				"Payload/App.app/icon.png", // Only 1 app file out of 3
			},
			wantShouldFilter: true,
		},
		{
			name: "App resources only - actionable",
			files: []string{
				"Payload/App.app/logo.png",
				"Payload/App.app/Resources/logo.png",
			},
			wantShouldFilter: false,
		},
		{
			name: "Custom framework (not third-party) - actionable",
			files: []string{
				"Payload/App.app/Frameworks/MyAppFramework.framework/icon.png",
				"Payload/App.app/Frameworks/MyOtherFramework.framework/icon.png",
			},
			wantShouldFilter: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dup := types.DuplicateSet{
				Files: tt.files,
				Count: len(tt.files),
				Size:  8192,
			}

			result := rule.Evaluate(dup)
			assert.Equal(t, tt.wantShouldFilter, result.ShouldFilter, "ShouldFilter mismatch")
		})
	}
}

func TestThirdPartySDKRule_Integration(t *testing.T) {
	categorizer := NewDuplicateCategorizer()

	duplicates := []types.DuplicateSet{
		{
			Files: []string{
				"Payload/App.app/Frameworks/GoogleMaps.framework/Resources/marker.png",
				"Payload/App.app/Frameworks/GoogleMapsBase.framework/Resources/marker.png",
			},
			Count: 2,
			Size:  10240,
		},
		{
			Files: []string{
				"Payload/App.app/Frameworks/Firebase.framework/Info.plist",
				"Payload/App.app/Frameworks/FirebaseCore.framework/Info.plist",
			},
			Count: 2,
			Size:  1024,
		},
		{
			Files: []string{
				"Payload/App.app/logo.png",
				"Payload/App.app/Resources/logo.png",
			},
			Count: 2,
			Size:  102400,
		},
	}

	actionable, filtered := categorizer.FilterDuplicates(duplicates)

	// Should filter out GoogleMaps and Firebase duplicates (2 sets)
	assert.Equal(t, 2, len(filtered), "Expected 2 filtered duplicates")

	// Should keep app logo.png (1 set)
	assert.Equal(t, 1, len(actionable), "Expected 1 actionable duplicate")
	assert.Contains(t, actionable[0].Files[0], "logo.png", "Expected logo.png to be actionable")
}
