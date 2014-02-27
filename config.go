// Compilation Time Configuration.

package tagbbs

const (
	SuperUser = "sysop"
)

var (
	version string
)

func init() {
	if len(version) == 0 {
		version = "dev"
	}
}
