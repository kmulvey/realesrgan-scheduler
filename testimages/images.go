package testimages

import _ "embed"

//go:embed fox.jpg
var FoxJPG []byte

//go:embed fox.png
var FoxPNG []byte

//go:embed fox.avif
var FoxAVIF []byte

//go:embed not_an_image.jpg
var NotAnImage []byte
