#version 330

uniform mat4 projection;
uniform mat4 camera;
uniform mat4 model;

in vec3 vert;
in vec3 color;

out vec3 vertColor;

void main() {
	vertColor = color;
    gl_Position = projection * camera * model * vec4(vert, 1);
}