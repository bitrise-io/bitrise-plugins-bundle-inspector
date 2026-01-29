# iOS Compression Support

This package provides support for various compression methods used in iOS artifacts.

## LZFSE Compression

LZFSE (Lempel-Ziv Finite State Entropy) is Apple's compression algorithm optimized for Apple platforms. Modern iOS IPAs may use LZFSE compression (ZIP compression method 99) instead of standard DEFLATE compression.

### Features

- **Automatic Detection**: The bundle-inspector automatically detects LZFSE-compressed files (method 99) when extracting IPAs
- **Transparent Decompression**: Files are automatically decompressed during extraction
- **Graceful Fallback**: If LZFSE support is not available, a helpful error message is shown

### Implementation

The LZFSE decompression is provided by `github.com/blacktop/lzfse-cgo`, which wraps Apple's official LZFSE library.

### Magic Bytes

LZFSE compressed data starts with one of these magic bytes:
- `bvx-` - LZFSE compressed block
- `bvx$` - LZFSE end of stream
- `bvxn` - LZFSE uncompressed block
- `bvx2` - LZVN compressed block

### Usage

LZFSE support is automatically enabled when analyzing IPAs:

```bash
bundle-inspector analyze app.ipa
```

If an IPA uses LZFSE compression, it will be automatically decompressed during analysis.

### Testing

Since the test artifacts don't include LZFSE-compressed IPAs, the LZFSE decompression is tested with:
1. Magic byte detection tests
2. Empty/invalid data error handling
3. Integration with ZIP extraction (will be triggered when a real LZFSE-compressed IPA is analyzed)

To test with a real LZFSE-compressed IPA, simply analyze any modern iOS IPA that uses this compression method.
