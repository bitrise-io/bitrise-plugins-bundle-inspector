package macho

import (
	"testing"

	"github.com/bitrise-io/bitrise-plugins-bundle-inspector/pkg/types"
	"github.com/stretchr/testify/assert"
)

func TestBuildDependencyGraph(t *testing.T) {
	binaries := map[string]*types.BinaryInfo{
		"Wikipedia": {
			LinkedLibraries: []string{
				"@rpath/Wikipedia.debug.dylib",
				"/usr/lib/libSystem.B.dylib",
			},
			RPaths: []string{
				"@executable_path",
				"@executable_path/Frameworks",
			},
		},
		"Wikipedia.debug.dylib": {
			LinkedLibraries: []string{
				"@rpath/WMF.framework/WMF",
				"/usr/lib/libSystem.B.dylib",
			},
			RPaths: []string{
				"@executable_path/Frameworks",
			},
		},
		"Frameworks/WMF.framework/WMF": {
			LinkedLibraries: []string{
				"/System/Library/Frameworks/Foundation.framework/Foundation",
				"/usr/lib/libSystem.B.dylib",
			},
			RPaths: []string{},
		},
	}

	graph := BuildDependencyGraph(binaries)

	assert.NotNil(t, graph)
	assert.Contains(t, graph, "Wikipedia")
	assert.Contains(t, graph, "Wikipedia.debug.dylib")
	assert.Contains(t, graph, "Frameworks/WMF.framework/WMF")

	// Wikipedia.debug.dylib should depend on WMF framework
	deps := graph["Wikipedia.debug.dylib"]
	assert.Contains(t, deps, "Frameworks/WMF.framework/WMF")

	// WMF framework should have no custom dependencies (only system libraries)
	wmfDeps := graph["Frameworks/WMF.framework/WMF"]
	assert.Empty(t, wmfDeps, "WMF should only have system dependencies which are filtered")
}

func TestDetectUnusedFrameworks(t *testing.T) {
	// Create a graph where Framework B is not used
	graph := DependencyGraph{
		"App": []string{
			"Frameworks/A.framework/A",
		},
		"Frameworks/A.framework/A": []string{},
		"Frameworks/B.framework/B": []string{},
	}

	unused := DetectUnusedFrameworks(graph, "App")

	assert.Len(t, unused, 1)
	assert.Contains(t, unused, "Frameworks/B.framework/B")
	assert.NotContains(t, unused, "Frameworks/A.framework/A")
}

func TestDetectUnusedFrameworksWithTransitiveDeps(t *testing.T) {
	// Create a graph with transitive dependencies
	// App -> A -> B (C is unused)
	graph := DependencyGraph{
		"App": []string{
			"Frameworks/A.framework/A",
		},
		"Frameworks/A.framework/A": []string{
			"Frameworks/B.framework/B",
		},
		"Frameworks/B.framework/B": []string{},
		"Frameworks/C.framework/C": []string{},
	}

	unused := DetectUnusedFrameworks(graph, "App")

	assert.Len(t, unused, 1)
	assert.Contains(t, unused, "Frameworks/C.framework/C")
	assert.NotContains(t, unused, "Frameworks/A.framework/A")
	assert.NotContains(t, unused, "Frameworks/B.framework/B")
}

func TestResolveDependencyPath(t *testing.T) {
	tests := []struct {
		name       string
		libPath    string
		binaryPath string
		rpaths     []string
		expected   string
	}{
		{
			name:       "rpath framework",
			libPath:    "@rpath/Foo.framework/Foo",
			binaryPath: "App",
			rpaths:     []string{"@executable_path/Frameworks"},
			expected:   "Frameworks/Foo.framework/Foo",
		},
		{
			name:       "system library",
			libPath:    "/usr/lib/libSystem.B.dylib",
			binaryPath: "App",
			rpaths:     []string{},
			expected:   "",
		},
		{
			name:       "executable_path",
			libPath:    "@executable_path/Frameworks/Foo.framework/Foo",
			binaryPath: "App",
			rpaths:     []string{},
			expected:   "Frameworks/Foo.framework/Foo",
		},
		{
			name:       "loader_path",
			libPath:    "@loader_path/Foo.dylib",
			binaryPath: "PlugIns/Extension.appex/Extension",
			rpaths:     []string{},
			expected:   "PlugIns/Extension.appex/Foo.dylib",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := resolveDependencyPath(tt.libPath, tt.binaryPath, tt.rpaths)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestIsSystemLibrary(t *testing.T) {
	tests := []struct {
		path     string
		expected bool
	}{
		{"/System/Library/Frameworks/Foundation.framework/Foundation", true},
		{"/usr/lib/libSystem.B.dylib", true},
		{"/usr/local/lib/libfoo.dylib", true},
		{"@rpath/Foo.framework/Foo", false},
		{"Frameworks/Bar.framework/Bar", false},
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			result := isSystemLibrary(tt.path)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestDetectUnusedFrameworks_EmptyMainBinary(t *testing.T) {
	graph := DependencyGraph{
		"Frameworks/A.framework/A": []string{},
		"Frameworks/B.framework/B": []string{},
	}

	// Empty main binary path should return nil
	unused := DetectUnusedFrameworks(graph, "")
	assert.Nil(t, unused)
}

func TestDetectUnusedFrameworks_MainBinaryNotInGraph(t *testing.T) {
	graph := DependencyGraph{
		"Frameworks/A.framework/A": []string{},
		"Frameworks/B.framework/B": []string{},
	}

	// Main binary not in graph should return nil instead of false positives
	unused := DetectUnusedFrameworks(graph, "PkgInfo")
	assert.Nil(t, unused, "should return nil when main binary not found in graph")

	unused = DetectUnusedFrameworks(graph, "NonExistent")
	assert.Nil(t, unused, "should return nil when main binary doesn't exist")
}

func TestDetectUnusedFrameworks_MainBinaryExists(t *testing.T) {
	graph := DependencyGraph{
		"Runner": []string{
			"Frameworks/A.framework/A",
		},
		"Frameworks/A.framework/A": []string{},
		"Frameworks/B.framework/B": []string{},
	}

	// When main binary exists in graph, should work correctly
	unused := DetectUnusedFrameworks(graph, "Runner")
	assert.Len(t, unused, 1)
	assert.Contains(t, unused, "Frameworks/B.framework/B")
	assert.NotContains(t, unused, "Frameworks/A.framework/A")
}

func TestDetectUnusedFrameworks_FlutterApp(t *testing.T) {
	// Flutter app with App.framework that's dynamically loaded
	graph := DependencyGraph{
		"Runner": []string{
			"Frameworks/Flutter.framework/Flutter",
		},
		"Frameworks/Flutter.framework/Flutter": []string{},
		"Frameworks/App.framework/App":         []string{},
		"Frameworks/Unused.framework/Unused":   []string{},
	}

	unused := DetectUnusedFrameworks(graph, "Runner")

	// App.framework should NOT be in unused list (dynamically loaded by Flutter)
	assert.NotContains(t, unused, "Frameworks/App.framework/App",
		"App.framework should not be flagged as unused in Flutter apps")

	// But genuinely unused frameworks should still be detected
	assert.Contains(t, unused, "Frameworks/Unused.framework/Unused")
	assert.Len(t, unused, 1)
}

func TestDetectUnusedFrameworks_AppFrameworkWithoutFlutter(t *testing.T) {
	// App.framework in a non-Flutter app should be flagged if unused
	graph := DependencyGraph{
		"MainApp": []string{
			"Frameworks/SomeOther.framework/SomeOther",
		},
		"Frameworks/SomeOther.framework/SomeOther": []string{},
		"Frameworks/App.framework/App":             []string{},
	}

	unused := DetectUnusedFrameworks(graph, "MainApp")

	// Without Flutter.framework present, App.framework CAN be flagged as unused
	assert.Contains(t, unused, "Frameworks/App.framework/App")
}

func TestIsDynamicallyLoadedFramework(t *testing.T) {
	tests := []struct {
		name          string
		frameworkPath string
		graph         DependencyGraph
		want          bool
	}{
		{
			name:          "App.framework in Flutter app",
			frameworkPath: "Frameworks/App.framework/App",
			graph: DependencyGraph{
				"Frameworks/Flutter.framework/Flutter": []string{},
			},
			want: true,
		},
		{
			name:          "App.framework without Flutter",
			frameworkPath: "Frameworks/App.framework/App",
			graph: DependencyGraph{
				"Frameworks/Other.framework/Other": []string{},
			},
			want: false,
		},
		{
			name:          "Other framework in Flutter app",
			frameworkPath: "Frameworks/Other.framework/Other",
			graph: DependencyGraph{
				"Frameworks/Flutter.framework/Flutter": []string{},
			},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isDynamicallyLoadedFramework(tt.frameworkPath, tt.graph)
			assert.Equal(t, tt.want, got)
		})
	}
}
