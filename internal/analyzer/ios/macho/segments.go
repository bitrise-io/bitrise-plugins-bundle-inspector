package macho

import (
	"debug/macho"
	"path/filepath"
	"sort"

	"github.com/bitrise-io/bitrise-plugins-bundle-inspector/pkg/types"
)

// SectionInfo contains metadata about a single Mach-O section.
type SectionInfo struct {
	Name    string `json:"name"`
	Segment string `json:"segment"`
	Size    int64  `json:"size"`
}

// SegmentInfo contains metadata about a Mach-O segment and its sections.
type SegmentInfo struct {
	Name     string        `json:"name"`
	FileSize int64         `json:"file_size"`
	Sections []SectionInfo `json:"sections,omitempty"`
}

// MachOSegments contains parsed segment/section information from a Mach-O binary.
type MachOSegments struct {
	Path         string        `json:"path"`
	Architecture string        `json:"architecture"`
	IsFat        bool          `json:"is_fat"`
	Segments     []SegmentInfo `json:"segments"`
}

// ParseSegments extracts segment and section information from a Mach-O binary.
// For fat binaries, it parses the first architecture only.
func ParseSegments(path string) (*MachOSegments, error) {
	result := &MachOSegments{
		Path:     path,
		Segments: make([]SegmentInfo, 0),
	}

	// Try to open as fat binary first
	fatFile, err := macho.OpenFat(path)
	if err == nil {
		defer fatFile.Close()
		result.IsFat = true

		if len(fatFile.Arches) > 0 {
			// Use first architecture
			arch := fatFile.Arches[0]
			result.Architecture = GetCPUTypeName(arch.Cpu)
			result.Segments = extractSegments(arch.File)
		}
		return result, nil
	}

	// Not a fat binary, try regular Mach-O
	file, err := macho.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	result.IsFat = false
	result.Architecture = GetCPUTypeName(file.Cpu)
	result.Segments = extractSegments(file)

	return result, nil
}

// extractSegments extracts segment and section information from a Mach-O file.
func extractSegments(file *macho.File) []SegmentInfo {
	segmentMap := make(map[string]*SegmentInfo)
	segmentOrder := make([]string, 0)

	// First pass: collect segments from load commands
	for _, load := range file.Loads {
		if seg, ok := load.(*macho.Segment); ok {
			// Skip __PAGEZERO (no file content, only virtual memory)
			if seg.Name == "__PAGEZERO" {
				continue
			}

			if _, exists := segmentMap[seg.Name]; !exists {
				segmentMap[seg.Name] = &SegmentInfo{
					Name:     seg.Name,
					FileSize: int64(seg.Filesz),
					Sections: make([]SectionInfo, 0),
				}
				segmentOrder = append(segmentOrder, seg.Name)
			}
		}
	}

	// Second pass: collect sections and associate with segments
	for _, section := range file.Sections {
		if segInfo, exists := segmentMap[section.Seg]; exists {
			segInfo.Sections = append(segInfo.Sections, SectionInfo{
				Name:    section.Name,
				Segment: section.Seg,
				Size:    int64(section.Size),
			})
		}
	}

	// Build result in segment order, sorted by file size descending
	segments := make([]SegmentInfo, 0, len(segmentOrder))
	for _, name := range segmentOrder {
		seg := segmentMap[name]

		// Sort sections by size descending
		sort.Slice(seg.Sections, func(i, j int) bool {
			return seg.Sections[i].Size > seg.Sections[j].Size
		})

		segments = append(segments, *seg)
	}

	// Sort segments by file size descending
	sort.Slice(segments, func(i, j int) bool {
		return segments[i].FileSize > segments[j].FileSize
	})

	return segments
}

// ExpandSegmentsAsChildren creates virtual FileNode children representing
// segments and their sections for a Mach-O binary.
func ExpandSegmentsAsChildren(segments *MachOSegments, binaryRelativePath string) []*types.FileNode {
	if segments == nil || len(segments.Segments) == 0 {
		return nil
	}

	children := make([]*types.FileNode, 0, len(segments.Segments))

	for _, seg := range segments.Segments {
		segPath := filepath.Join(binaryRelativePath, seg.Name)

		segNode := &types.FileNode{
			Path:       segPath,
			Name:       seg.Name,
			Size:       seg.FileSize,
			IsDir:      true,
			IsVirtual:  true,
			SourceFile: binaryRelativePath,
			Children:   make([]*types.FileNode, 0),
		}

		// Calculate sum of section sizes
		var sectionSizeSum int64
		for _, section := range seg.Sections {
			sectionSizeSum += section.Size
		}

		// Add section children
		for _, section := range seg.Sections {
			sectionPath := filepath.Join(segPath, section.Name)
			sectionNode := &types.FileNode{
				Path:       sectionPath,
				Name:       section.Name,
				Size:       section.Size,
				IsDir:      false,
				IsVirtual:  true,
				SourceFile: binaryRelativePath,
			}
			segNode.Children = append(segNode.Children, sectionNode)
		}

		// Handle segments without sections (like __LINKEDIT)
		// Create a single child representing the segment's content
		if len(seg.Sections) == 0 && seg.FileSize > 0 {
			linkerDataNode := &types.FileNode{
				Path:       filepath.Join(segPath, "linker_data"),
				Name:       "linker_data",
				Size:       seg.FileSize,
				IsDir:      false,
				IsVirtual:  true,
				SourceFile: binaryRelativePath,
			}
			segNode.Children = append(segNode.Children, linkerDataNode)
		}

		// Add __unmapped entry if section sizes don't account for all segment data
		if len(seg.Sections) > 0 && sectionSizeSum < seg.FileSize {
			unmappedSize := seg.FileSize - sectionSizeSum
			unmappedNode := &types.FileNode{
				Path:       filepath.Join(segPath, "__unmapped"),
				Name:       "__unmapped",
				Size:       unmappedSize,
				IsDir:      false,
				IsVirtual:  true,
				SourceFile: binaryRelativePath,
			}
			segNode.Children = append(segNode.Children, unmappedNode)

			// Re-sort children after adding unmapped
			sort.Slice(segNode.Children, func(i, j int) bool {
				return segNode.Children[i].Size > segNode.Children[j].Size
			})
		}

		children = append(children, segNode)
	}

	return children
}
