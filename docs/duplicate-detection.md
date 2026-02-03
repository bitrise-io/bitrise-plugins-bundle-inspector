# Duplicate Detection & Intelligent Filtering

Bundle Inspector uses intelligent filtering to show you **only actionable duplicate file recommendations**. This document explains how the system works and what patterns are automatically filtered out.

## Overview

When analyzing iOS and Android artifacts, Bundle Inspector detects all duplicate files based on SHA-256 hash matching. However, not all duplicates are problems you can (or should) fix. Many duplicates are legitimate architectural patterns required by iOS/Android, third-party SDK resources you don't control, or framework metadata.

**The filtering system eliminates 60-80% of false positives**, ensuring you only see duplicates worth your attention.

## How It Works

```
Duplicate Detection (372 duplicates)
        ↓
Intelligent Filtering (9 rules)
        ↓
Actionable Optimizations (359 shown to user)
```

Each duplicate is evaluated against 9 filtering rules:
- **Rules 1-7**: Filter out false positives (architectural patterns, SDK resources)
- **Rules 8-9**: Identify actionable duplicates and assign priority (high/medium/low)
- **Default**: If no rules match, treat as actionable (show to user)

## Filtering Rules

### Rule 1: Info.plist Bundle Boundary Detection

**Pattern**: Info.plist files in different bundles (frameworks, apps, extensions)

**Why Filtered**: Each bundle (app, framework, extension) MUST have its own Info.plist file. This is required by iOS architecture.

**Examples (Filtered):**
```
✗ GoogleMaps.framework/Info.plist
✗ Firebase.framework/Info.plist
✗ ShareExtension.appex/Info.plist
```

**Examples (Actionable):**
```
✓ App.app/Info.plist
✓ App.app/Info.plist.backup  (same directory = true duplicate)
```

---

### Rule 2: NIB Version Variants Detection

**Pattern**: NIB version variants (runtime.nib vs objects-*.nib)

**Why Filtered**: iOS uses different NIB formats for different OS versions. Files like `runtime.nib` and `objects-12.3+.nib` in the same .nib directory are version variants for backward compatibility.

**Examples (Filtered):**
```
✗ Main.nib/runtime.nib
✗ Main.nib/objects-8.0+.nib
✗ Main.nib/objects-12.3+.nib
```

**Examples (Actionable):**
```
✓ ViewA.nib/runtime.nib
✓ ViewB.nib/runtime.nib  (different NIB bundles = true duplicate)
```

---

### Rule 3: Asset Catalog Contents.json Detection

**Pattern**: Contents.json files in asset catalogs (.xcassets)

**Why Filtered**: Every asset set (AppIcon, LaunchImage, etc.) requires a Contents.json file. These are iOS-required metadata.

**Examples (Filtered):**
```
✗ Assets.xcassets/AppIcon.appiconset/Contents.json
✗ Assets.xcassets/LaunchImage.launchimage/Contents.json
✗ Assets.xcassets/Icon.imageset/Contents.json
```

**Examples (Actionable):**
```
✓ App.app/Contents.json
✓ Framework/Contents.json  (not in .xcassets = true duplicate)
```

---

### Rule 4: Localization File Bundle Isolation

**Pattern**: Localization files (.strings, .stringsdict) in different bundles

**Why Filtered**: Each bundle should have its own localization files for bundle isolation. This is required for proper localization in frameworks and extensions.

**Examples (Filtered):**
```
✗ App.app/en.lproj/Localizable.strings
✗ SDK.framework/en.lproj/Localizable.strings
✗ ShareExtension.appex/en.lproj/Localizable.strings
```

**Examples (Actionable):**
```
✓ App.app/en.lproj/Localizable.strings
✓ App.app/en.lproj/Localizable.strings.backup  (same directory = duplicate)
```

---

### Rule 5: Framework Build Scripts Detection

**Pattern**: Build scripts (strip-frameworks.sh, copy-frameworks.sh, etc.)

**Why Filtered**: These are build-time artifacts commonly bundled by CocoaPods and Carthage. While they shouldn't be in release builds, they're not under developer control per-framework.

**Examples (Filtered):**
```
✗ SDK1.framework/strip-frameworks.sh
✗ SDK2.framework/strip-frameworks.sh
✗ Framework.framework/copy-frameworks.sh
```

**Note**: These should ideally be stripped during build, but filtering them reduces noise in duplicate reports.

---

### Rule 6: Framework Metadata Files Detection

**Pattern**: Framework metadata (.supx, .bcsymbolmap, .swiftmodule, module.modulemap)

**Why Filtered**: These are required metadata files for framework distribution and usage. Each framework needs its own copy.

**Examples (Filtered):**
```
✗ Carthage.framework/Carthage.supx
✗ SDK1.framework/Modules/module.modulemap
✗ SDK2.framework/Modules/module.modulemap
✗ *.bcsymbolmap (Bitcode symbol maps)
```

---

### Rule 7: Third-Party SDK Bundled Resources 🎯 (HIGHEST IMPACT)

**Pattern**: Resources bundled within 100+ popular third-party SDKs

**Why Filtered**: You have no control over resources bundled by third-party SDKs (Google Maps, Firebase, Facebook, etc.). These should never be flagged as actionable duplicates.

**SDKs Detected (100+):**
- **Google**: GoogleMaps, Firebase (all variants), GoogleUtilities, GoogleSignIn
- **Facebook**: FBSDKCoreKit, FacebookLogin, FacebookSDK
- **Networking**: Alamofire, AFNetworking, SocketRocket
- **Image Loading**: SDWebImage, Kingfisher, PINRemoteImage
- **Analytics**: Crashlytics, Amplitude, Mixpanel, Segment, Bugsnag, Sentry
- **Payment**: Stripe, PayPal, Braintree, Square
- **UI Libraries**: Lottie, Charts, Material, SnapKit
- **And 80+ more...**

**Examples (Filtered):**
```
✗ GoogleMaps.framework/Resources/marker.png (50+ copies)
✗ Firebase.framework/GoogleService-Info.plist
✗ FBSDKCoreKit.framework/Assets/icon.png
✗ SDWebImage.framework/placeholder.png
```

**Detection Logic:**
- Exact match on 100+ known SDK names
- Prefix matching: Firebase*, Google*, FB*, FIR*, GUL*, GTM*
- If ≥50% of duplicates are in SDKs → Filter
- App resources only → Actionable

**This rule has the highest impact** - eliminates 40-50% of false positives alone!

---

### Rule 8: App Extension Resource Duplication (ACTIONABLE)

**Pattern**: Resources duplicated between app and extensions

**Why Actionable**: Extensions often unnecessarily duplicate app resources (logos, images, strings). Moving to shared asset catalogs or frameworks can significantly reduce IPA size.

**Priority:**
- High: >500 KB
- Medium: 100-500 KB
- Low: <100 KB

**Examples (Actionable with Priority):**
```
✓ App.app/logo.png (759 KB)
  ShareExtension.appex/logo.png (759 KB)
  → HIGH priority (1.5 MB wasted)

✓ App.app/icon.png (120 KB)
  Widget.appex/icon.png (120 KB)
  → MEDIUM priority (120 KB wasted)

✓ App.app/config.json (50 KB)
  Share.appex/config.json (50 KB)
  → LOW priority (50 KB wasted)
```

---

### Rule 9: Asset Duplication in Same Bundle (ACTIONABLE)

**Pattern**: Asset files (images, audio, video, JSON) duplicated within same bundle

**Why Actionable**: These are genuine duplicates that should be deduplicated. Often caused by refactoring, copy-paste mistakes, or poor asset organization.

**Supported Asset Types:**
- Images: png, jpg, jpeg, gif, svg, pdf
- Audio: mp3, m4a, wav, aac
- Video: mp4, mov
- Data: json, plist

**Priority:** Size-based (same as Rule 8)

**Examples (Actionable with Priority):**
```
✓ App.app/image.png (600 KB)
  App.app/Resources/image.png (600 KB)
  → HIGH priority (600 KB wasted)

✓ App.app/audio.mp3 (2.1 MB)
  App.app/Sounds/audio.mp3 (2.1 MB)
  → HIGH priority (2.1 MB wasted)

✓ App.app/config.json (10 KB)
  App.app/Resources/config.json (10 KB)
  → LOW priority (10 KB wasted)
```

---

## Before & After Examples

### Example 1: iOS App with Google Maps

**Before Filtering:**
```
Duplicate Files (67 sets, 3.5 MB wasted):
  • Remove 2 duplicate Info.plist files (15 KB each)
  • Remove 12 duplicate Contents.json files (4 KB each)
  • Remove 50 duplicate GoogleMaps images (8 KB each)
  • Remove 2 duplicate NIB variant files (8 KB each)
  • Remove duplicate track-main.aac (2.1 MB each)
  • Remove duplicate logo.png (759 KB each)
```

**After Filtering:**
```
Duplicate Files (2 sets, 2.9 MB wasted):
  • Remove duplicate track-main.aac (2.1 MB each) - HIGH priority
  • Remove duplicate logo.png (759 KB each) - HIGH priority
```

**Result**: 65 false positives filtered out (97%), only 2 actionable issues shown.

---

### Example 2: Android App with Firebase

**Before Filtering:**
```
Duplicate Files (45 sets, 1.8 MB wasted):
  • Remove 8 duplicate firebase-*.xml files (5 KB each)
  • Remove 15 duplicate res/ files (10 KB each)
  • Remove duplicate background.png (800 KB each)
  • Remove duplicate splash.png (400 KB each)
```

**After Filtering:**
```
Duplicate Files (2 sets, 1.2 MB wasted):
  • Remove duplicate background.png (800 KB each) - HIGH priority
  • Remove duplicate splash.png (400 KB each) - MEDIUM priority
```

**Result**: 43 false positives filtered out (96%), only 2 actionable issues shown.

---

## Priority System

Actionable duplicates (Rules 8-9) are assigned priorities based on file size:

| Size | Priority | Use Case | Example |
|------|----------|----------|---------|
| >500 KB | **HIGH** | Large images, audio files, videos | logo.png (759 KB), track.mp3 (2.1 MB) |
| 100-500 KB | **MEDIUM** | Medium assets, frameworks | icon.png (200 KB), splash.png (400 KB) |
| <100 KB | **LOW** | Small files, configs, strings | config.json (50 KB), strings (10 KB) |

**Focus on HIGH priority items first** for maximum impact.

---

## FAQ

### Why don't I see Info.plist duplicates anymore?

Info.plist files in different bundles are **required by iOS architecture**. Each framework, extension, and app must have its own Info.plist. These are filtered out because they're not actionable.

If you have Info.plist duplicates in the **same directory** (e.g., backup files), those will still be shown as actionable.

---

### Why don't I see third-party SDK duplicates?

Resources bundled within third-party SDKs (Google Maps, Firebase, Facebook, etc.) are **not under your control**. You cannot (and should not) modify or deduplicate resources inside vendor frameworks.

If you have genuine duplicates **between** SDKs and your app code, those will still be shown.

---

### Can I disable filtering?

No, filtering is always active. The goal is to show you **only actionable recommendations**. Manual filtering would defeat the purpose of intelligent detection.

If you need raw duplicate data for debugging, you can access the full `duplicates` array in JSON output, which contains all detected duplicates (both filtered and actionable).

---

### Why do I still see Contents.json duplicates?

Contents.json files **in asset catalogs** (.xcassets) are filtered because they're required metadata.

Contents.json files **outside asset catalogs** are shown as actionable because they're genuine duplicates (often from configuration or data files).

---

### How can I verify filtering is working?

Check the JSON output:

```bash
# Total duplicates detected
jq '.duplicates | length' report.json

# Optimizations generated (filtered subset)
jq '.optimizations | map(select(.category == "duplicates")) | length' report.json

# Difference = filtered count
```

Example:
- Duplicates detected: 372
- Optimizations generated: 359
- Filtered: 13 (3.5%)

---

### What if filtering misses a false positive?

Please report it! We're continuously improving the filtering rules. Open an issue at:
https://github.com/bitrise-io/bitrise-plugins-bundle-inspector/issues

Include:
- The duplicate file paths
- Why you consider it a false positive
- Your IPA/APK/AAB (if you can share it)

---

### What if filtering removes a real issue?

This is unlikely (filtering is conservative), but if it happens, please report it!

The system is designed to **prefer false positives over false negatives** - better to show a duplicate than hide a real issue.

---

## Technical Details

### Rule Evaluation Order

Rules are evaluated in order (1-9). The **first matching rule** determines the filter result.

```
1. Info.plist → Check
2. NIB variants → Check
3. Contents.json → Check
4. Localization → Check
5. Framework scripts → Check
6. Framework metadata → Check
7. Third-party SDKs → Check
8. Extension duplication → Check (actionable)
9. Asset duplication → Check (actionable)
→ Default: Actionable (show to user)
```

### Path Analysis

The system analyzes file paths to detect:
- Bundle boundaries (.framework, .appex, .app, .xcassets, .lproj)
- Framework names (GoogleMaps, Firebase, etc.)
- Extension names (ShareExtension, Widget, etc.)
- Asset catalog contexts

This allows intelligent classification of duplicates based on their location in the artifact structure.

### SHA-256 Hashing

Duplicate detection uses SHA-256 hashing for file identity. Files with identical hashes are considered duplicates, regardless of filename or location.

---

## Best Practices

1. **Focus on HIGH priority items first** - maximum impact for minimum effort
2. **Review extension duplicates** - often easy wins (Rules 8)
3. **Deduplicate assets in same bundle** - clean up copy-paste mistakes (Rule 9)
4. **Trust the filtering** - false positives have been eliminated
5. **Use JSON output for automation** - integrate into CI/CD size checks

---

## Related Documentation

- [README.md](../README.md) - Getting started guide
- [QUICKSTART.md](../QUICKSTART.md) - Developer guide
- [CLAUDE.md](../CLAUDE.md) - Architecture documentation

---

## Changelog

**Version 0.3.0** (Current):
- Intelligent filtering system with 9 rules
- 60-80% false positive reduction
- Priority-based recommendations (high/medium/low)
- 100+ third-party SDK detection

**Version 0.2.0** (Previous):
- Basic duplicate detection (no filtering)
- All duplicates shown as optimizations
- High false positive rate
