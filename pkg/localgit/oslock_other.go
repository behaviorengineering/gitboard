//go:build !unix && !windows

package localgit

// defaultOSLocker is nil on platforms without a cross-process lock implementation.
// In-process MutationCoordinator serialization still applies inside one process.
func defaultOSLocker() OSLocker {
	return nil
}
