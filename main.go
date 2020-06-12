package main

import (
	"fmt"
	"log"
	"runtime"

	"github.com/xlab/android-go/android"
	"github.com/xlab/android-go/app"
	"github.com/xlab/android-go/egl"
	gl "github.com/xlab/android-go/gles3"
)

const vertexShaderSource = `
	uniform mat4 ProjMtx;
	
	in vec2 Position;
	in vec2 TexCoord;
	in vec4 Color;
	
	out vec2 Frag_UV;
	out vec4 Frag_Color;

	void main() {
		Frag_UV = TexCoord;
		Frag_Color = Color;
		gl_Position = ProjMtx * vec4(Position.xy, 0, 1);
	}
`
const fragmentShaderSource = `
	precision mediump float;
	uniform sampler2D Texture;
	
	in vec2 Frag_UV;
	in vec4 Frag_Color;
	
	out vec4 Out_Color;

	void main(){
		Out_Color = Frag_Color * texture(Texture, Frag_UV.st);
	}
`

var vbo = make([]uint32, 1)
var vao = make([]uint32, 1)

func init() {
	app.SetLogTag("THDWB")
}

func main() {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	nativeWindowEvents := make(chan app.NativeWindowEvent, 1)
	windowFocusEvents := make(chan app.WindowFocusEvent, 1)
	inputQueueEvents := make(chan app.InputQueueEvent, 1)
	inputQueueChan := make(chan *android.InputQueue, 1)

	var displayHandle *egl.DisplayHandle
	var windowFocused bool

	app.Main(func(a app.NativeActivity) {
		a.HandleNativeWindowEvents(nativeWindowEvents)
		a.HandleWindowFocusEvents(windowFocusEvents)
		a.HandleInputQueueEvents(inputQueueEvents)

		go app.HandleInputQueues(inputQueueChan, func() {
			a.InputQueueHandled()
		}, app.LogInputEvents)

		a.InitDone()
		for {
			select {
			case <-a.LifecycleEvents():
			case event := <-windowFocusEvents:
				if event.HasFocus && !windowFocused {
					windowFocused = true
				}
				if !event.HasFocus && windowFocused {
					windowFocused = false
				}

				draw(displayHandle)
			case event := <-inputQueueEvents:
				switch event.Kind {
				case app.QueueCreated:
					inputQueueChan <- event.Queue
				case app.QueueDestroyed:
					inputQueueChan <- nil
				}
			case event := <-nativeWindowEvents:
				switch event.Kind {
				case app.NativeWindowRedrawNeeded:
					draw(displayHandle)
					a.NativeWindowRedrawDone()

				case app.NativeWindowCreated:
					expectedSurface := map[int32]int32{
						egl.SurfaceType:          egl.WindowBit,
						egl.ContextClientVersion: 3.0,

						egl.RedSize:   8,
						egl.GreenSize: 8,
						egl.BlueSize:  8,
						egl.AlphaSize: 8,
						egl.DepthSize: 24,
					}

					if handle, err := egl.NewDisplayHandle(event.Window, expectedSurface); err != nil {
						log.Fatalln("EGL error:", err)
					} else {
						displayHandle = handle
						log.Printf("EGL display res: %dx%d", handle.Width, handle.Height)
					}
					initGL()

				case app.NativeWindowDestroyed:
					displayHandle.Destroy()
				}
			}
		}
	})
}

func initGL() {
	gl.Enable(gl.BLEND)
	gl.BlendEquation(gl.FUNC_ADD)
	gl.BlendFunc(gl.SRC_ALPHA, gl.ONE_MINUS_SRC_ALPHA)
	gl.Disable(gl.CULL_FACE)
	gl.Disable(gl.DEPTH_TEST)
	gl.Enable(gl.SCISSOR_TEST)

	fmt.Println(vertexShaderSource)
	fmt.Println(fragmentShaderSource)

	program := gl.CreateProgram()

	vertexShader := gl.CreateShader(gl.VERTEX_SHADER)
	fragmentShader := gl.CreateShader(gl.FRAGMENT_SHADER)

	assignShader(vertexShader, vertexShaderSource)
	assignShader(fragmentShader, fragmentShaderSource)

	var status int32
	gl.GetShaderiv(vertexShader, gl.COMPILE_STATUS, &status)
	if status != gl.TRUE {
		panic("vert_shdr failed to compile")
	}
	gl.GetShaderiv(fragmentShader, gl.COMPILE_STATUS, &status)
	if status != gl.TRUE {
		panic("frag_shdr failed to compile")
	}

	gl.AttachShader(program, vertexShader)
	gl.AttachShader(program, fragmentShader)

	gl.LinkProgram(program)

	gl.GetProgramiv(program, gl.LINK_STATUS, &status)
	if status != gl.TRUE {
		panic("gl program failed to link")
	}

	gl.GenBuffers(1, vbo)
	gl.GenVertexArrays(1, vao)

	gl.BindVertexArray(vao[0])
	gl.BindBuffer(gl.ARRAY_BUFFER, vbo[0])

	gl.BindTexture(gl.TEXTURE_2D, 0)
	gl.BindBuffer(gl.ARRAY_BUFFER, 0)
	gl.BindVertexArray(0)
	fmt.Println("gl enabled")
}

func draw(handle *egl.DisplayHandle) {
	gl.ClearColor(.2, .8, 1, 1)
	gl.Clear(gl.COLOR_BUFFER_BIT)
	fmt.Println("drawing")

	handle.SwapBuffers()
}

func assignShader(shaderHandle uint32, shaderSource string) {
	gl.ShaderSource(shaderHandle, 2, []string{"#version 300 es\x00", shaderSource + "\x00"}, nil)
	gl.CompileShader(shaderHandle)
}
