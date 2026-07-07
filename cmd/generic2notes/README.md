# Enrich FreeBSD sources NOTES files with GENERIC comments

It also provides filtered GENERIC and NOTES files for further processing. It takes into account DEFAULT files as well.

I'm too tired of keeping my kernel config up to date with each major release. Obviously, the GENERIC doesn't contain all options and devices. But the NOTES files don't contain everything either. Plus, DEFAULTS introduced some time ago can change. To make my life easier, I created this helper. Later, it can be improved by adding various "bells and whistles" such as command-line argument processing, semantic versioning, adding the commit hash in version output, using goreleaser, cosign, etc., but I'm happy with the result it gives me right now.
