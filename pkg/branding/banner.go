// Package branding contains terminal identity strings shared by CLI entrypoints.
package branding

const (
	bannerWhite = "\x1b[1;38;5;255m"
	bannerName  = "\x1b[1;38;5;205m"
	bannerProxy = "\x1b[1;38;5;51m"
	bannerReset = "\x1b[0m"
)

// Banner is intentionally compact so it fits above help, setup, and startup summaries.
const Banner = "" +
	bannerWhite + "    / \\__\n" +
	"   (    @\\___\n" +
	"   /         O\n" +
	"  /   (_____/\n" +
	" /_____/   U\n" + bannerReset +
	bannerName + "chicha-ip-proxy\n" + bannerReset +
	bannerProxy + "TCP / UDP proxy\n" + bannerReset
