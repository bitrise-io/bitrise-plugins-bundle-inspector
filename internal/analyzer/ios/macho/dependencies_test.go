package macho

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBuildDependencyGraph(t *testing.T) {
	binaries := map[string]*BinaryInfo{
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
