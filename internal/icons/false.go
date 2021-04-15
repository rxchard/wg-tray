package icons

import (
	"github.com/gobuffalo/packr/v2"
	"log"
)

func False() []byte {
	box := packr.New("icons", "../../assets")
	b, err := box.Find("false.png")

	if err != nil {
		log.Println("failed to find false.png")
	}

	return b
}
