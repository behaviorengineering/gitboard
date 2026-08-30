package dashboard

// Commands wraps Service for mutating operations: prune and pull.
// Read operations (Collect, ClearCaches) remain on Service.
type Commands struct {
	*Service
}

// NewCommands returns a Commands backed by the given Service.
func NewCommands(s *Service) *Commands {
	return &Commands{Service: s}
}
