package main

import (
	"image/png"
	"os"
	"unsafe"

	"image/color"

	"github.com/go-gl/gl/v4.5-core/gl"
	"github.com/go-gl/mathgl/mgl32"

	"image"
	"math/rand"
)

const vertLen = int(unsafe.Sizeof(vertex{}))

type point mgl32.Vec3

type vertex struct {
	pos   point
	color point
}

type mesh struct {
	points []vertex
	index  []uint32
	vao    uint32
}

func (m *mesh) add(pos, color point) uint32 {
	m.points = append(m.points, vertex{
		pos:   pos,
		color: color,
	})
	return uint32(len(m.points) - 1)
}

func (m *mesh) triangle(a, b, c uint32) {
	m.index = append(m.index, a, b, c)
}

func (m *mesh) initGL(program uint32) {
	// Configure the vertex data
	var vao uint32
	gl.GenVertexArrays(1, &vao)
	gl.BindVertexArray(vao)

	var vbo uint32
	gl.GenBuffers(1, &vbo)
	gl.BindBuffer(gl.ARRAY_BUFFER, vbo)
	gl.BufferData(gl.ARRAY_BUFFER, len(m.points)*vertLen, gl.Ptr(m.points), gl.STATIC_DRAW)

	vertAttrib := uint32(gl.GetAttribLocation(program, gl.Str("vert\x00")))
	gl.EnableVertexAttribArray(vertAttrib)
	gl.VertexAttribPointer(vertAttrib, 3, gl.FLOAT, false, int32(vertLen), gl.PtrOffset(0))

	colorAttrib := uint32(gl.GetAttribLocation(program, gl.Str("color\x00")))
	gl.EnableVertexAttribArray(colorAttrib)
	gl.VertexAttribPointer(colorAttrib, 3, gl.FLOAT, false, int32(vertLen), gl.PtrOffset(3*4))

	if len(m.index) > 0 {
		var indices uint32
		gl.GenBuffers(1, &indices)
		gl.BindBuffer(gl.ELEMENT_ARRAY_BUFFER, indices)
		gl.BufferData(gl.ELEMENT_ARRAY_BUFFER, len(m.index)*int(unsafe.Sizeof(m.index[0])), gl.Ptr(m.index), gl.STATIC_DRAW)
	}

	m.vao = vao
}

func (m *mesh) render(modelUniform int32, model mgl32.Mat4, mode uint32) {
	gl.UniformMatrix4fv(modelUniform, 1, false, &model[0])

	gl.BindVertexArray(m.vao)

	if len(m.index) > 0 {
		gl.DrawElements(mode, int32(len(m.index)), gl.UNSIGNED_INT, nil)
	} else {
		gl.DrawArrays(mode, 0, int32(len(m.points)))
	}
}

func loadImage(filename string) image.Image {
	f, err := os.Open(filename)
	if err != nil {
		panic(err)
	}
	img, err := png.Decode(f)
	if err != nil {
		panic(err)
	}
	return img
}

func loadImageMeshYCbCr(img image.Image) *mesh {
	m := new(mesh)
	bounds := img.Bounds()
	for i := bounds.Min.X; i < bounds.Max.X; i++ {
		for j := bounds.Min.Y; j < bounds.Max.Y; j++ {
			c := img.At(i, j)
			cr, cg, cb, _ := c.RGBA()
			humanColor := color.YCbCrModel.Convert(c).(color.YCbCr)
			x := float32(humanColor.Y) / 255
			y := float32(humanColor.Cb) / 255
			z := float32(humanColor.Cr) / 255

			r := float32(cr) / 65535
			g := float32(cg) / 65535
			b := float32(cb) / 65535

			m.add(point{x, y, z}, point{r, g, b})
		}
	}
	return m
}

func loadImageMeshRGB(img image.Image) *mesh {
	m := new(mesh)
	bounds := img.Bounds()
	for i := bounds.Min.X; i < bounds.Max.X; i++ {
		for j := bounds.Min.Y; j < bounds.Max.Y; j++ {
			c := img.At(i, j)
			cr, cg, cb, _ := c.RGBA()
			r := float32(cr) / 65535
			g := float32(cg) / 65535
			b := float32(cb) / 65535

			m.add(point{r, g, b}, point{r, g, b})
		}
	}
	return m
}

func disturb(m *mesh) *mesh {
	for _, point := range m.points {
		point.pos[0] += rand.Float32() / 20
		point.pos[1] += rand.Float32() / 20
		point.pos[2] += rand.Float32() / 20
	}
	return m
}

func loadSampleMeshRGB() *mesh {
	const limit = 10
	m := new(mesh)
	for i := 0; i < limit; i++ {
		for j := 0; j < limit; j++ {
			for k := 0; k < limit; k++ {
				r := (float32(i) / (limit - 1))
				g := (float32(j) / (limit - 1))
				b := (float32(k) / (limit - 1))
				m.add(point{r, g, b}, point{r, g, b})
			}
		}
	}
	return m
}

func loadSampleMeshYCbCr() *mesh {
	const limit = 10
	m := new(mesh)
	for i := 0; i < limit; i++ {
		for j := 0; j < limit; j++ {
			for k := 0; k < limit; k++ {
				r := (float32(i) / (limit - 1))
				g := (float32(j) / (limit - 1))
				b := (float32(k) / (limit - 1))

				y := color.YCbCrModel.Convert(color.RGBA{R: uint8(r * 255), G: uint8(g * 255), B: uint8(b * 255)}).(color.YCbCr)

				m.add(point{float32(y.Y) / 255, float32(y.Cb) / 255, float32(y.Cr) / 255}, point{r, g, b})
			}
		}
	}
	return m
}
