package main

import (
	"fmt"
	"image/color"
	_ "image/png"
	"log"
	"runtime"
	"strings"

	"io/ioutil"

	"github.com/Logiraptor/palette"
	"github.com/go-gl/gl/v4.5-core/gl"
	"github.com/go-gl/glfw/v3.2/glfw"
	"github.com/go-gl/mathgl/mgl32"
)

const windowWidth = 1920
const windowHeight = 1080

func init() {
	// GLFW event handling must run on the main OS thread
	runtime.LockOSThread()
}

func main() {
	if err := glfw.Init(); err != nil {
		log.Fatalln("failed to initialize glfw:", err)
	}
	defer glfw.Terminate()

	glfw.WindowHint(glfw.Resizable, glfw.False)
	glfw.WindowHint(glfw.ContextVersionMajor, 4)
	glfw.WindowHint(glfw.ContextVersionMinor, 5)
	glfw.WindowHint(glfw.OpenGLProfile, glfw.OpenGLCoreProfile)
	glfw.WindowHint(glfw.OpenGLForwardCompatible, glfw.True)
	window, err := glfw.CreateWindow(windowWidth, windowHeight, "Cube", nil, nil)
	if err != nil {
		panic(err)
	}
	window.MakeContextCurrent()

	// Initialize Glow
	if err := gl.Init(); err != nil {
		panic(err)
	}

	version := gl.GoStr(gl.GetString(gl.VERSION))
	fmt.Println("OpenGL version", version)

	// Configure the vertex and fragment shaders
	program, err := newProgramFromFiles("flat_shaded.vs", "flat_shaded.fs")
	if err != nil {
		panic(err)
	}

	gl.UseProgram(program)

	projection := mgl32.Perspective(mgl32.DegToRad(45.0), float32(windowWidth)/windowHeight, 0.1, 10.0)
	projectionUniform := gl.GetUniformLocation(program, gl.Str("projection\x00"))
	gl.UniformMatrix4fv(projectionUniform, 1, false, &projection[0])

	camera := mgl32.LookAtV(mgl32.Vec3{3, 3, 3}, mgl32.Vec3{0, 0, 0}, mgl32.Vec3{0, 1, 0})
	cameraUniform := gl.GetUniformLocation(program, gl.Str("camera\x00"))
	gl.UniformMatrix4fv(cameraUniform, 1, false, &camera[0])

	model := mgl32.Ident4()
	modelUniform := gl.GetUniformLocation(program, gl.Str("model\x00"))
	gl.UniformMatrix4fv(modelUniform, 1, false, &model[0])

	gl.BindFragDataLocation(program, 0, gl.Str("outputColor\x00"))

	const numDivisions = 6

	var subDivisions = make([]*mesh, 0, numDivisions)

	img := loadImage("/tmp/wallpaper.png")

	c := palette.ColorCut{}
	c.Quantize(make(color.Palette, 0, numDivisions), img)

	for _, box := range c.Boxes {
		avg := box.Avg()
		cube := makeCube(point{
			float32(box.Min.R) / 255,
			float32(box.Min.G) / 255,
			float32(box.Min.B) / 255,
		}, point{
			float32(box.Rng.R) / 255,
			float32(box.Rng.G) / 255,
			float32(box.Rng.B) / 255,
		}, point{
			float32(avg.R) / 255,
			float32(avg.G) / 255,
			float32(avg.B) / 255,
		})
		cube.initGL(program)

		subDivisions = append(subDivisions, cube)
	}

	meshYCbCr := loadImageMeshYCbCr(img)
	meshYCbCr.initGL(program)

	meshRGB := loadImageMeshRGB(img)
	meshRGB.initGL(program)

	sampleMeshRGB := loadSampleMeshRGB()
	sampleMeshYCbCr := loadSampleMeshYCbCr()
	sampleMeshRGB.initGL(program)
	sampleMeshYCbCr.initGL(program)

	// Configure global settings
	gl.Enable(gl.DEPTH_TEST)
	gl.DepthFunc(gl.LESS)
	gl.ClearColor(0.4, 0.4, 0.4, 0.4)

	angle := 0.0
	previousTime := glfw.GetTime()

	gl.PointSize(2)

	for !window.ShouldClose() {
		gl.Clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT)

		// Update
		time := glfw.GetTime()
		elapsed := time - previousTime
		previousTime = time

		angle += elapsed
		model = mgl32.HomogRotate3D(float32(angle), mgl32.Vec3{0, 1, 0})

		// Render
		gl.UseProgram(program)
		// meshYCbCr.render(modelUniform, model, gl.POINTS)
		// sampleMeshYCbCr.render(modelUniform, model, gl.POINTS)

		meshRGB.render(modelUniform, model, gl.POINTS)
		// sampleMeshRGB.render(modelUniform, model, gl.POINTS)

		gl.PolygonMode(gl.FRONT_AND_BACK, gl.LINE)
		for _, sub := range subDivisions {
			sub.render(modelUniform, model, gl.TRIANGLES)
		}

		// Maintenance
		window.SwapBuffers()
		glfw.PollEvents()
	}
}

func newProgramFromFiles(vsFile, fsFile string) (uint32, error) {
	vsBuf, err := ioutil.ReadFile(vsFile)
	if err != nil {
		return 0, err
	}
	fsBuf, err := ioutil.ReadFile(fsFile)
	if err != nil {
		return 0, err
	}

	return newProgram(string(vsBuf), string(fsBuf))
}

func newProgram(vertexShaderSource, fragmentShaderSource string) (uint32, error) {
	vertexShader, err := compileShader(vertexShaderSource, gl.VERTEX_SHADER)
	if err != nil {
		return 0, err
	}

	fragmentShader, err := compileShader(fragmentShaderSource, gl.FRAGMENT_SHADER)
	if err != nil {
		return 0, err
	}

	program := gl.CreateProgram()

	gl.AttachShader(program, vertexShader)
	gl.AttachShader(program, fragmentShader)
	gl.LinkProgram(program)

	var status int32
	gl.GetProgramiv(program, gl.LINK_STATUS, &status)
	if status == gl.FALSE {
		var logLength int32
		gl.GetProgramiv(program, gl.INFO_LOG_LENGTH, &logLength)

		log := strings.Repeat("\x00", int(logLength+1))
		gl.GetProgramInfoLog(program, logLength, nil, gl.Str(log))

		return 0, fmt.Errorf("failed to link program: %v", log)
	}

	gl.DeleteShader(vertexShader)
	gl.DeleteShader(fragmentShader)

	return program, nil
}

func compileShader(source string, shaderType uint32) (uint32, error) {
	shader := gl.CreateShader(shaderType)

	csources, free := gl.Strs(source)
	gl.ShaderSource(shader, 1, csources, nil)
	free()
	gl.CompileShader(shader)

	var status int32
	gl.GetShaderiv(shader, gl.COMPILE_STATUS, &status)
	if status == gl.FALSE {
		var logLength int32
		gl.GetShaderiv(shader, gl.INFO_LOG_LENGTH, &logLength)

		log := strings.Repeat("\x00", int(logLength+1))
		gl.GetShaderInfoLog(shader, logLength, nil, gl.Str(log))

		return 0, fmt.Errorf("failed to compile %v: %v", source, log)
	}

	return shader, nil
}
