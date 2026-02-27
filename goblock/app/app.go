package app

import (
	"log"

	"github.com/go-gl/gl/v4.1-core/gl"
	"github.com/go-gl/glfw/v3.3/glfw"
	"github.com/go-gl/mathgl/mgl32"
)

type App struct {
	window *glfw.Window
	//shaders       *ShaderManager
	camera *Camera
	//chunkRenderer *RhunkRenderer
}

func Start() {
	log.Println("Starting application...")

	a := &App{}
	a.Init()
	defer a.Terminate()

	//a.shaders = newShaderManager("./shaders")
	//a.shaders.Add("main")

	a.camera = newCamera(mgl32.Vec3{0, 0, 2}, a.window)

	//a.chunkRenderer = newChunkRenderer(a.shaders, a.camera)

	//chunk := a.chunkRenderer.CreateChunk(mgl32.Vec3{0, 0, 0})

	a.camera.SetLookHandler()

	isPressed := false
	a.window.SetMouseButtonCallback(func(w *glfw.Window, button glfw.MouseButton, action glfw.Action, mods glfw.ModifierKey) {
		if button != glfw.MouseButtonLeft {
			return
		}

		if action == glfw.Press && !isPressed {
			isPressed = true
			//if chunk.target != nil {
			//a.chunkRenderer.BreakBlock(chunk.target)
			//fmt.Println(chunk.target.WorldPos())
			//}
		} else if action == glfw.Release {
			isPressed = false
		}
	})

	for !a.window.ShouldClose() && a.window.GetKey(glfw.KeyQ) != glfw.Press {
		gl.Clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT)

		a.camera.HandleMove()

		//a.chunkRenderer.SetTargetBlock(chunk)
		//a.chunkRenderer.Draw(chunk, mgl32.Vec3{0, 0, 0})

		a.window.SwapBuffers()
		glfw.PollEvents()
	}
}

func (a *App) Init() {
	a.window = initWindow()

	gl.Enable(gl.DEPTH_TEST)
	gl.ClearColor(1.0, 1.0, 1.0, 1.0)
}

func (a *App) Terminate() {
	glfw.Terminate()
}
