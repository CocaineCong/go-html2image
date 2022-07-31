package html2image

import (
	"context"
	"fmt"
	"io/ioutil"
	"testing"
)

func Test_html2image(t *testing.T) {
	result := DoctronConfig{}
	requestDTO := newDefaultHtml2ImageRequestDTO()
	result.Url = requestDTO.Url
	ins := new(html2image)
	ins.Conver = Conver{
		ctx: context.Background(),
		cc: ConvertConfig{
			Url:    "https://www.baidu.com",
			Params: convertToHtml2ImageParams(requestDTO),
		},
	}

	img, err := ins.Convert()
	if err != nil {
		fmt.Println("err", err)
		return
	}
	err = ioutil.WriteFile("baidu.png", img, 0666)
	if err != nil {
		fmt.Println("err", err)
		return
	}
}
