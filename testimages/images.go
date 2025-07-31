package testimages

import _ "embed"

//go:embed fox.jpg
var FoxJPG []byte // Embed the .jpg file as a byte slice

//go:embed fox.png
var FoxPNG []byte // Embed the .jpg file as a byte slice

//go:embed fox.avif
var FoxAVIF []byte // Embed the .jpg file as a byte slice
