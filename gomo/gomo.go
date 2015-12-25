// +build darwin linux

package gomo

import (
	"github.com/pankona/gomo-simra/peer"
	"github.com/pankona/gomo-simra/scene"

	_ "image/jpeg"
	_ "image/png"

	"golang.org/x/mobile/app"
	"golang.org/x/mobile/event/lifecycle"
	"golang.org/x/mobile/event/paint"
	"golang.org/x/mobile/event/size"
	"golang.org/x/mobile/event/touch"
	"golang.org/x/mobile/gl"
)

func main() {
	app.Main(func(a app.App) {
		glPeer := peer.GetGLPeer()
		touchPeer := peer.GetTouchPeer()
		sceneCtrl := scene.GetControllerInstance()

		for e := range a.Events() {
			switch e := a.Filter(e).(type) {
			case lifecycle.Event:
				switch e.Crosses(lifecycle.StageVisible) {
				case lifecycle.CrossOn:
					glctx, _ := e.DrawContext.(gl.Context)

					// initialize gl peer
					glPeer.Initialize(glctx)

					// initialize scene controller
					//sceneCtrl.Initialize()

					// start scene controller
					//sceneCtrl.Start()

					a.Send(paint.Event{})
				case lifecycle.CrossOff:
					sceneCtrl.Stop()
				}
			case size.Event:
				peer.SetScreenSize(e)
			case paint.Event:
				if e.External {
					continue
				}

				sceneCtrl.Update()

				a.Publish()
				a.Send(paint.Event{}) // keep animating
			case touch.Event:
				switch e.Type {
				case touch.TypeBegin:
					touchPeer.OnTouchBegin(e.X, e.Y)
				case touch.TypeMove:
					touchPeer.OnTouchMove(e.X, e.Y)
				case touch.TypeEnd:
					touchPeer.OnTouchEnd(e.X, e.Y)
				default:
				}
			}
		}
	})
}