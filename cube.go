package main

func makeCube(origin, size, color point) *mesh {
	m := new(mesh)

	extent := point{
		origin[0] + size[0],
		origin[1] + size[1],
		origin[2] + size[2],
	}

	verts := []uint32{
		m.add(point{origin[0], origin[1], extent[2]}, color),
		m.add(point{extent[0], origin[1], extent[2]}, color),
		m.add(point{extent[0], extent[1], extent[2]}, color),
		m.add(point{origin[0], extent[1], extent[2]}, color),
		m.add(point{origin[0], origin[1], origin[2]}, color),
		m.add(point{extent[0], origin[1], origin[2]}, color),
		m.add(point{extent[0], extent[1], origin[2]}, color),
		m.add(point{origin[0], extent[1], origin[2]}, color),
	}

	// front
	m.triangle(verts[0], verts[1], verts[2])
	m.triangle(verts[2], verts[3], verts[0])
	// top
	m.triangle(verts[1], verts[5], verts[6])
	m.triangle(verts[6], verts[2], verts[1])
	// back
	m.triangle(verts[7], verts[6], verts[5])
	m.triangle(verts[5], verts[4], verts[7])
	// bottom
	m.triangle(verts[4], verts[0], verts[3])
	m.triangle(verts[3], verts[7], verts[4])
	// left
	m.triangle(verts[4], verts[5], verts[1])
	m.triangle(verts[1], verts[0], verts[4])
	// right
	m.triangle(verts[3], verts[2], verts[6])
	m.triangle(verts[6], verts[7], verts[3])

	return m
}
