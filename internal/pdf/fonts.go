package pdf

import _ "embed"

//go:embed fonts/DejaVuSansCondensed.ttf
var dejaVuRegular []byte

//go:embed fonts/DejaVuSansCondensed-Bold.ttf
var dejaVuBold []byte
