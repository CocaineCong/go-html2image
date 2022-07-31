package html2image

import (
	"context"
	"errors"
	"github.com/chromedp/cdproto/emulation"
	"github.com/chromedp/cdproto/page"
	"github.com/chromedp/chromedp"
	uuid "github.com/iris-contrib/go.uuid"
	"time"
)

//FormatPng Image compression format (defaults to png).
const FormatPng = page.CaptureScreenshotFormatPng

//DefaultQuality Compression quality from range [0..100] (jpeg only).
const DefaultQuality = 0

//DefaultFromSurface Capture the screenshot from the surface, rather than the view. Defaults to true.
const DefaultFromSurface = true

//DefaultViewportX Capture the screenshot of a given region only.
// X offset in device independent pixels (dip).
const DefaultViewportX = 0

//DefaultViewportY Y offset in device independent pixels (dip).
const DefaultViewportY = 0

//DefaultViewportWidth Rectangle width in device independent pixels (dip).
const DefaultViewportWidth = 996

//DefaultViewportHeight Rectangle height in device independent pixels (dip).
const DefaultViewportHeight = 996

//DefaultViewportScale Page scale factor.
const DefaultViewportScale = 1

//DefaultWaitingTime Waiting time after the page loaded. Default 0 means not wait. unit:Millisecond
const DefaultWaitingTime = 0

type Html2ImageParams struct {
	page.CaptureScreenshotParams
	CustomClip  bool
	WaitingTime int // Waiting time after the page loaded. Default 0 means not wait. unit:Millisecond
}

type CommonRequestDTO struct {
	Url       string `schema:"url,omitempty" validate:"required,url"`
	UploadKey string `schema:"uploadKey,omitempty" validate:"omitempty"`
	Username  string `schema:"u,omitempty" validate:"required"`
	Password  string `schema:"p,omitempty" validate:"required"`
}

type Html2ImageRequestDTO struct {
	CommonRequestDTO
	Format      page.CaptureScreenshotFormat `schema:"format,omitempty" validate:"omitempty"`      // Image compression format (defaults to png).
	Quality     int64                        `schema:"quality,omitempty" validate:"omitempty"`     // Compression quality from range [0..100] (jpeg only).
	CustomClip  bool                         `schema:"customClip,omitempty" validate:"omitempty"`  //if set this value, the below clip will work,otherwise not work!
	ClipX       float64                      `schema:"clipX,omitempty" validate:"omitempty"`       // Capture the screenshot of a given region only.X offset in device independent pixels (dip).
	ClipY       float64                      `schema:"clipY,omitempty" validate:"omitempty"`       // Capture the screenshot of a given region only.Y offset in device independent pixels (dip).
	ClipWidth   float64                      `schema:"clipWidth,omitempty" validate:"omitempty"`   // Capture the screenshot of a given region only.Rectangle width in device independent pixels (dip).
	ClipHeight  float64                      `schema:"clipHeight,omitempty" validate:"omitempty"`  // Capture the screenshot of a given region only.Rectangle height in device independent pixels (dip).
	ClipScale   float64                      `schema:"clipScale,omitempty" validate:"omitempty"`   // Capture the screenshot of a given region only.Page scale factor.
	FromSurface bool                         `schema:"fromSurface,omitempty" validate:"omitempty"` // Capture the screenshot from the surface, rather than the view. Defaults to true.
	WaitingTime int                          `schema:"waitingTime,omitempty" validate:"omitempty"` // Waiting time after the page loaded. Default 0 means not wait. unit:Millisecond
}

//NewDefaultHtml2ImageParams default html convert to image params
func NewDefaultHtml2ImageParams() Html2ImageParams {
	return Html2ImageParams{
		CustomClip: false,
		CaptureScreenshotParams: page.CaptureScreenshotParams{
			Format:  FormatPng,
			Quality: DefaultQuality,
			Clip: &page.Viewport{
				X:      DefaultViewportX,
				Y:      DefaultViewportY,
				Width:  DefaultViewportWidth,
				Height: DefaultViewportHeight,
				Scale:  DefaultViewportScale,
			},
			FromSurface: DefaultFromSurface,
		},
		WaitingTime: DefaultWaitingTime,
	}
}

func newDefaultHtml2ImageRequestDTO() *Html2ImageRequestDTO {
	return &Html2ImageRequestDTO{
		Format:      FormatPng,
		Quality:     DefaultQuality,
		CustomClip:  false,
		ClipX:       DefaultViewportX,
		ClipY:       DefaultViewportY,
		ClipWidth:   DefaultViewportWidth,
		ClipHeight:  DefaultViewportHeight,
		ClipScale:   DefaultViewportScale,
		FromSurface: DefaultFromSurface,
		WaitingTime: DefaultWaitingTime,
	}
}

func convertToHtml2ImageParams(requestDTO *Html2ImageRequestDTO) Html2ImageParams {
	params := NewDefaultHtml2ImageParams()
	params.Format = requestDTO.Format
	params.Quality = requestDTO.Quality
	params.CustomClip = requestDTO.CustomClip
	params.Clip.X = requestDTO.ClipX
	params.Clip.Y = requestDTO.ClipY
	params.Clip.Width = requestDTO.ClipWidth
	params.Clip.Height = requestDTO.ClipHeight
	params.Clip.Scale = requestDTO.ClipScale
	params.FromSurface = requestDTO.FromSurface
	params.WaitingTime = requestDTO.WaitingTime
	return params
}

type ConvertConfig struct {
	Url    string `validate:"required,url"`
	Params Params
}

type DoctronConfig struct {
	TraceId     uuid.UUID
	Ctx         context.Context
	DoctronType int
	ConvertConfig
}

type Params interface {
}

type Converter interface {
	Convert() ([]byte, error)
	GetConvertElapsed() time.Duration
}

type Conver struct {
	ctx            context.Context
	cc             ConvertConfig
	buf            []byte
	convertElapsed time.Duration
}

type html2image struct {
	Conver
}

func (ins *html2image) GetConvertElapsed() time.Duration {
	return ins.convertElapsed
}

func (ins *html2image) Convert() ([]byte, error) {
	start := time.Now()
	defer func() {
		ins.convertElapsed = time.Since(start)
	}()
	var params Html2ImageParams
	params, ok := ins.cc.Params.(Html2ImageParams)
	if !ok {
		return nil, errors.New("wrong html2image params given")
	}
	ctx, cancel := chromedp.NewContext(ins.ctx)
	defer cancel()

	if err := chromedp.Run(ctx,
		chromedp.Navigate(ins.cc.Url),
		chromedp.Sleep(time.Duration(params.WaitingTime)*time.Millisecond),
		chromedp.ActionFunc(func(ctx context.Context) error {

			if !params.CustomClip {
				// get layout metrics
				_, _, contentSize, _, _, _, err := page.GetLayoutMetrics().Do(ctx)
				if err != nil {
					return err
				}
				params.Clip.X = contentSize.X
				params.Clip.Y = contentSize.Y
				params.Clip.Width = contentSize.Width
				params.Clip.Height = contentSize.Height
			}

			// force viewport emulation
			err := emulation.SetDeviceMetricsOverride(int64(params.Clip.Width), int64(params.Clip.Height), 1, false).
				WithScreenOrientation(&emulation.ScreenOrientation{
					Type:  emulation.OrientationTypePortraitPrimary,
					Angle: 0,
				}).
				Do(ctx)
			if err != nil {
				return err
			}

			// capture screenshot
			ins.buf, err = params.Do(ctx)
			return err
		}),
	); err != nil {
		return nil, err
	}

	return ins.buf, nil
}
