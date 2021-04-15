package icons

import (
	"github.com/gobuffalo/packr/v2"
	"log"
)

func True() []byte {
	box := packr.New("icons", "../../assets")
	b, err := box.Find("true.png")

	if err != nil {
		log.Println("failed to find true.png")
	}

	return b
}
