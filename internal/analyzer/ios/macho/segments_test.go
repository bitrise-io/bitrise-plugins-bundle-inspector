package macho

import (
	"os"
	"testing"
)

func TestExpandSegmentsAsChildren(t *testing.T) {
	tests := []struct {
		name           string
		segments       *MachOSegments
		binaryPath     string
		expectedCount  int
		checkFirstSeg  string
		checkFirstSize int64
	}{
		{
			name:          "nil segments",
			segments:      nil,
			binaryPath:    "TestApp",
			expectedCount: 0,
		},
		{
			name: "empty segments",
			segments: &MachOSegments{
				Path:     "TestApp",
				Segments: []SegmentInfo{},
			},
			binaryPath:    "TestApp",
			expectedCount: 0,
		},
		{
			name: "single segment with sections",
			segments: &MachOSegments{
				Path:         "TestApp",
				Architecture: "arm64",
				Segments: []SegmentInfo{
					{
						Name:     "__TEXT",
						FileSize: 16384,
						Sections: []SectionInfo{
							{Name: "__text", Segment: "__TEXT", Size: 8000},
							{Name: "__stubs", Segment: "__TEXT", Size: 500},
						},
					},
				},
			},
			binaryPath:     "TestApp",
			expectedCount:  1,
			checkFirstSeg:  "__TEXT",
			checkFirstSize: 16384,
		},
		{
			name: "linkedit segment without sections",
			segments: &MachOSegments{
				Path:         "TestApp",
				Architecture: "arm64",
				Segments: []SegmentInfo{
					{
						Name:     "__LINKEDIT",
						FileSize: 32768,
						Sections: []SectionInfo{},
					},
				},
			},
			binaryPath:    "TestApp",
			expectedCount: 1,
		},
		{
			name: "multiple segments",
			segments: &MachOSegments{
				Path:         "TestApp",
				Architecture: "arm64",
				Segments: []SegmentInfo{
					{
						Name:     "__TEXT",
						FileSize: 16384,
						Sections: []SectionInfo{
							{Name: "__text", Segment: "__TEXT", Size: 8000},
						},
					},
					{
						Name:     "__DATA",
						FileSize: 8192,
						Sections: []SectionInfo{
							{Name: "__data", Segment: "__DATA", Size: 4000},
						},
					},
					{
						Name:     "__LINKEDIT",
						FileSize: 4096,
						Sections: []SectionInfo{},
					},
				},
			},
			binaryPath:    "TestApp",
			expectedCount: 3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			children := ExpandSegmentsAsChildren(tt.segments, tt.binaryPath)

			if len(children) != tt.expectedCount {
				t.Errorf("expected %d children, got %d", tt.expectedCount, len(children))
			}

			if tt.expectedCount > 0 && tt.checkFirstSeg != "" {
				if children[0].Name != tt.checkFirstSeg {
					t.Errorf("expected first segment name %s, got %s", tt.checkFirstSeg, children[0].Name)
				}
				if children[0].Size != tt.checkFirstSize {
					t.Errorf("expected first segment size %d, got %d", tt.checkFirstSize, children[0].Size)
				}
			}

			// Verify all children are virtual and have proper source file
			for _, child := range children {
				if !child.IsVirtual {
					t.Errorf("segment node should be virtual")
				}
				if child.SourceFile != tt.binaryPath {
					t.Errorf("expected source file %s, got %s", tt.binaryPath, child.SourceFile)
				}
				if !child.IsDir {
					t.Errorf("segment node should be a directory")
				}
			}
		})
	}
}

func TestExpandSegmentsAsChildren_UnmappedData(t *testing.T) {
	segments := &MachOSegments{
		Path:         "TestApp",
		Architecture: "arm64",
		Segments: []SegmentInfo{
			{
				Name:     "__TEXT",
				FileSize: 16384,
				Sections: []SectionInfo{
					{Name: "__text", Segment: "__TEXT", Size: 8000},
					{Name: "__stubs", Segment: "__TEXT", Size: 500},
				},
			},
		},
	}

	children := ExpandSegmentsAsChildren(segments, "TestApp")

	if len(children) != 1 {
		t.Fatalf("expected 1 segment, got %d", len(children))
	}

	textSeg := children[0]
	if textSeg.Name != "__TEXT" {
		t.Fatalf("expected __TEXT segment, got %s", textSeg.Name)
	}

	// Should have 3 children: __text, __stubs, __unmapped
	if len(textSeg.Children) != 3 {
		t.Errorf("expected 3 section children (including __unmapped), got %d", len(textSeg.Children))
	}

	// Find unmapped entry
	var hasUnmapped bool
	var unmappedSize int64
	for _, child := range textSeg.Children {
		if child.Name == "__unmapped" {
			hasUnmapped = true
			unmappedSize = child.Size
		}
	}

	if !hasUnmapped {
		t.Error("expected __unmapped entry for unaccounted segment space")
	}

	// Expected unmapped: 16384 - 8000 - 500 = 7884
	expectedUnmapped := int64(16384 - 8000 - 500)
	if unmappedSize != expectedUnmapped {
		t.Errorf("expected unmapped size %d, got %d", expectedUnmapped, unmappedSize)
	}
}

func TestExpandSegmentsAsChildren_LinkeditCreatesLinkerData(t *testing.T) {
	segments := &MachOSegments{
		Path:         "TestApp",
		Architecture: "arm64",
		Segments: []SegmentInfo{
			{
				Name:     "__LINKEDIT",
				FileSize: 32768,
				Sections: []SectionInfo{}, // LINKEDIT has no sections
			},
		},
	}

	children := ExpandSegmentsAsChildren(segments, "TestApp")

	if len(children) != 1 {
		t.Fatalf("expected 1 segment, got %d", len(children))
	}

	linkedit := children[0]
	if linkedit.Name != "__LINKEDIT" {
		t.Fatalf("expected __LINKEDIT segment, got %s", linkedit.Name)
	}

	// Should have 1 child: linker_data
	if len(linkedit.Children) != 1 {
		t.Errorf("expected 1 child (linker_data), got %d", len(linkedit.Children))
	}

	if linkedit.Children[0].Name != "linker_data" {
		t.Errorf("expected child named linker_data, got %s", linkedit.Children[0].Name)
	}

	if linkedit.Children[0].Size != 32768 {
		t.Errorf("expected linker_data size 32768, got %d", linkedit.Children[0].Size)
	}
}

func TestParseSegments(t *testing.T) {
	wikipediaBinary := "../../../../test-artifacts/ios/Wikipedia.app/Wikipedia"

	if _, err := os.Stat(wikipediaBinary); os.IsNotExist(err) {
		t.Skip("Wikipedia test artifact not found")
	}

	segments, err := ParseSegments(wikipediaBinary)
	if err != nil {
		t.Fatalf("ParseSegments failed: %v", err)
	}

	if segments == nil {
		t.Fatal("ParseSegments returned nil")
	}

	if segments.Architecture == "" {
		t.Error("expected non-empty architecture")
	}

	if len(segments.Segments) == 0 {
		t.Fatal("expected at least one segment")
	}

	// Verify common segments exist
	segmentNames := make(map[string]bool)
	for _, seg := range segments.Segments {
		segmentNames[seg.Name] = true
		t.Logf("Segment %s: %d bytes, %d sections", seg.Name, seg.FileSize, len(seg.Sections))

		// Log some sections for debugging
		for _, section := range seg.Sections {
			t.Logf("  Section %s: %d bytes", section.Name, section.Size)
		}
	}

	// Most Mach-O binaries have __TEXT and __LINKEDIT
	if !segmentNames["__TEXT"] {
		t.Error("expected __TEXT segment")
	}
	if !segmentNames["__LINKEDIT"] {
		t.Error("expected __LINKEDIT segment")
	}

	// __TEXT should have sections
	for _, seg := range segments.Segments {
		if seg.Name == "__TEXT" {
			if len(seg.Sections) == 0 {
				t.Error("__TEXT segment should have sections")
			}
			// Should have __text section (code)
			hasTextSection := false
			for _, section := range seg.Sections {
				if section.Name == "__text" {
					hasTextSection = true
					break
				}
			}
			if !hasTextSection {
				t.Error("__TEXT segment should have __text section")
			}
		}
	}
}

func TestParseSegments_Framework(t *testing.T) {
	frameworkBinary := "../../../../test-artifacts/ios/Wikipedia.app/Frameworks/WMF.framework/WMF"

	if _, err := os.Stat(frameworkBinary); os.IsNotExist(err) {
		t.Skip("WMF framework test artifact not found")
	}

	segments, err := ParseSegments(frameworkBinary)
	if err != nil {
		t.Fatalf("ParseSegments failed: %v", err)
	}

	if segments == nil {
		t.Fatal("ParseSegments returned nil")
	}

	if len(segments.Segments) == 0 {
		t.Error("expected segments in framework binary")
	}

	t.Logf("Framework architecture: %s, is fat: %v", segments.Architecture, segments.IsFat)
	for _, seg := range segments.Segments {
		t.Logf("Segment %s: %d bytes", seg.Name, seg.FileSize)
	}
}
