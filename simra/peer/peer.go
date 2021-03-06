package peer

import (
	"image"
	"log"
	"time"

	"golang.org/x/mobile/asset"
	"golang.org/x/mobile/exp/app/debug"
	"golang.org/x/mobile/exp/f32"
	"golang.org/x/mobile/exp/gl/glutil"
	"golang.org/x/mobile/exp/sprite"
	"golang.org/x/mobile/exp/sprite/clock"
	"golang.org/x/mobile/exp/sprite/glsprite"
	"golang.org/x/mobile/gl"
)

var glPeer *GLPeer

var startTime = time.Now()

// GLPeer represents gl context.
// Singleton.
type GLPeer struct {
	glctx  gl.Context
	images *glutil.Images
	fps    *debug.FPS
	eng    sprite.Engine
	scene  *sprite.Node
}

// GetGLPeer returns a instance of GLPeer.
// Since GLPeer is singleton, it is necessary to
// call this function to get GLPeer instance.
func GetGLPeer() *GLPeer {
	LogDebug("IN")
	if glPeer == nil {
		glPeer = &GLPeer{}
	}
	LogDebug("OUT")
	return glPeer
}

// Initialize initializes GLPeer.
// This function must be called inadvance of using GLPeer
func (glpeer *GLPeer) Initialize(glctx gl.Context) {
	LogDebug("IN")
	glpeer.glctx = glctx

	// transparency of png
	glpeer.glctx.Enable(gl.BLEND)
	glpeer.glctx.BlendEquation(gl.FUNC_ADD)
	glpeer.glctx.BlendFunc(gl.SRC_ALPHA, gl.ONE_MINUS_SRC_ALPHA)
	glpeer.images = glutil.NewImages(glctx)
	glpeer.fps = debug.NewFPS(glpeer.images)
	glpeer.initEng()

	LogDebug("OUT")
}

func (glpeer *GLPeer) initEng() {
	if glpeer.eng != nil {
		glpeer.eng.Release()
	}
	glpeer.eng = glsprite.Engine(glpeer.images)
	glpeer.scene = &sprite.Node{}
	glpeer.eng.Register(glpeer.scene)
	glpeer.eng.SetTransform(glpeer.scene, f32.Affine{
		{1, 0, 0},
		{0, 1, 0},
	})
}

func (glpeer *GLPeer) newNode() *sprite.Node {
	n := &sprite.Node{}
	glpeer.eng.Register(n)
	glpeer.scene.AppendChild(n)
	return n
}

func (glpeer *GLPeer) appendChild(n *sprite.Node) {
	glpeer.scene.AppendChild(n)
}

func (glpeer *GLPeer) removeChild(n *sprite.Node) {
	glpeer.scene.RemoveChild(n)
}

// LoadTexture return texture that is loaded by the information of arguments.
// Loaded texture can assign using AddSprite function.
func (glpeer *GLPeer) LoadTexture(assetName string, rect image.Rectangle) sprite.SubTex {
	LogDebug("IN")
	a, err := asset.Open(assetName)
	if err != nil {
		log.Fatal(err)
	}
	defer a.Close()

	img, _, err := image.Decode(a)
	if err != nil {
		log.Fatal(err)
	}
	t, err := glpeer.eng.LoadTexture(img)
	if err != nil {
		log.Fatal(err)
	}

	LogDebug("OUT")
	return sprite.SubTex{t, rect}
}

// LoadTextureFromImage return texture that is loaded by the information of arguments.
// Loaded texture can assign using AddSprite function.
func (glpeer *GLPeer) LoadTextureFromImage(img image.Image, rect image.Rectangle) sprite.SubTex {
	LogDebug("IN")
	t, err := glpeer.eng.LoadTexture(img)
	if err != nil {
		log.Fatal(err)
	}

	LogDebug("OUT")
	return sprite.SubTex{t, rect}
}

// Finalize finalizes GLPeer.
// This is called at termination of application.
func (glpeer *GLPeer) Finalize() {
	LogDebug("IN")
	GetSpriteContainer().RemoveSprites()
	glpeer.eng.Release()
	glpeer.fps.Release()
	glpeer.images.Release()
	glpeer.glctx = nil
	LogDebug("OUT")
}

// Update updates screen.
// This is called 60 times per 1 sec.
func (glpeer *GLPeer) Update() {
	if glpeer.glctx == nil {
		return
	}
	glpeer.glctx.ClearColor(0, 0, 0, 1) // black background
	glpeer.glctx.Clear(gl.COLOR_BUFFER_BIT)
	now := clock.Time(time.Since(startTime) * 60 / time.Second)

	glpeer.apply()

	glpeer.eng.Render(glpeer.scene, now, sz)
	glpeer.fps.Draw(sz)
}

// Reset resets current gl context.
// All sprites are also cleaned.
// This is called at changing of scene, and
// this function is for clean previous scene.
func (glpeer *GLPeer) Reset() {
	LogDebug("IN")
	GetSpriteContainer().RemoveSprites()
	glpeer.initEng()
	LogDebug("OUT")
}

func (glpeer *GLPeer) apply() {

	snpairs := GetSpriteContainer().spriteNodePairs

	for i := range snpairs {
		sc := snpairs[i]
		if sc.sprite == nil || !sc.inuse {
			continue
		}

		affine := &f32.Affine{
			{1, 0, 0},
			{0, 1, 0},
		}
		affine.Translate(affine,
			sc.sprite.X*desiredScreenSize.scale-sc.sprite.W/2*desiredScreenSize.scale+desiredScreenSize.marginWidth/2,
			(desiredScreenSize.height-sc.sprite.Y)*desiredScreenSize.scale-sc.sprite.H/2*desiredScreenSize.scale+desiredScreenSize.marginHeight/2)
		if sc.sprite.R != 0 {
			affine.Translate(affine,
				0.5*sc.sprite.W*desiredScreenSize.scale,
				0.5*sc.sprite.H*desiredScreenSize.scale)
			affine.Rotate(affine, sc.sprite.R)
			affine.Translate(affine,
				-0.5*sc.sprite.W*desiredScreenSize.scale,
				-0.5*sc.sprite.H*desiredScreenSize.scale)
		}
		affine.Scale(affine,
			sc.sprite.W*desiredScreenSize.scale,
			sc.sprite.H*desiredScreenSize.scale)
		glpeer.eng.SetTransform(sc.node, *affine)
	}
}

type arrangerFunc func(e sprite.Engine, n *sprite.Node, t clock.Time)

func (a arrangerFunc) Arrange(e sprite.Engine, n *sprite.Node, t clock.Time) { a(e, n, t) }
