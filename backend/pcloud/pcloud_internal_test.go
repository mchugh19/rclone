package pcloud

import (
	"testing"

	"github.com/rclone/rclone/fs/hash"
	"github.com/stretchr/testify/assert"
)

// TestCallbackHostname checks that only hosts within pCloud's domain
// are accepted.
func TestCallbackHostname(t *testing.T) {
	for _, test := range []struct {
		in   string
		want string // "" means an error is expected
	}{
		{"", defaultHostname},
		{"api.pcloud.com", "api.pcloud.com"},
		{"eapi.pcloud.com", "eapi.pcloud.com"},
		{"EAPI.pCloud.COM", "eapi.pcloud.com"},
		{"new-region.api.pcloud.com", "new-region.api.pcloud.com"},
		{"pcloud.com", ""},
		{".pcloud.com", ""},
		{"-.pcloud.com", ""},
		{"a-.pcloud.com", ""},
		{"a..pcloud.com", ""},
		{"api.pcloud.com.", ""},
		{"api.pclоud.com", ""}, // Cyrillic о
		{"127.0.0.1:18443", ""},
		{"evil.example.com", ""},
		{"api.pcloud.com.evil.example.com", ""},
		{"api.pcloud.com:443", ""},
		{"api.pcloud.com/path", ""},
		{"user@evil.example.com#.pcloud.com", ""},
		{"evil.example.com?.pcloud.com", ""},
	} {
		t.Run("["+test.in+"]", func(t *testing.T) {
			got, err := callbackHostname(test.in)
			if test.want == "" {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, test.want, got)
		})
	}
}

// TestCheckUploadedObject checks the comparison at the heart of
// upload verification.
func TestCheckUploadedObject(t *testing.T) {
	for _, test := range []struct {
		name       string
		size       int64
		storedSize int64
		wantHash   string
		storedHash string
		want       bool
	}{
		{"match", 100, 100, "abc123", "abc123", true},
		{"match case insensitive", 100, 100, "abc123", "ABC123", true},
		{"size differs", 100, 99, "abc123", "abc123", false},
		{"hash differs", 100, 100, "abc123", "def456", false},
		{"hash not reported yet", 100, 100, "abc123", "", false},
		{"empty file match", 0, 0, "e3b0c4", "e3b0c4", true},
	} {
		t.Run(test.name, func(t *testing.T) {
			got := checkUploadedObject(test.size, test.storedSize, test.wantHash, test.storedHash)
			assert.Equal(t, test.want, got)
		})
	}
}

// TestStrongestHash checks the verification hash selection.
func TestStrongestHash(t *testing.T) {
	hashType, sum, err := strongestHash(map[hash.Type]string{
		hash.MD5:    "md5sum",
		hash.SHA1:   "sha1sum",
		hash.SHA256: "sha256sum",
	})
	assert.NoError(t, err)
	assert.Equal(t, hash.SHA256, hashType)
	assert.Equal(t, "sha256sum", sum)

	hashType, sum, err = strongestHash(map[hash.Type]string{
		hash.MD5:  "md5sum",
		hash.SHA1: "sha1sum",
	})
	assert.NoError(t, err)
	assert.Equal(t, hash.SHA1, hashType)
	assert.Equal(t, "sha1sum", sum)

	hashType, sum, err = strongestHash(map[hash.Type]string{
		hash.MD5: "md5sum",
	})
	assert.NoError(t, err)
	assert.Equal(t, hash.MD5, hashType)
	assert.Equal(t, "md5sum", sum)

	_, _, err = strongestHash(map[hash.Type]string{})
	assert.Error(t, err)

	_, _, err = strongestHash(map[hash.Type]string{hash.SHA256: ""})
	assert.Error(t, err)
}
