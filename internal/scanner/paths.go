package scanner

// DefaultSystemPaths returns standard Linux system binary paths
func DefaultSystemPaths() []string {
	return []string{
		"/bin",
		"/sbin",
		"/usr/bin",
		"/usr/sbin",
		"/usr/local/bin",
		"/usr/local/sbin",
	}
}
