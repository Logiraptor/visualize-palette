#version 330

uniform sampler2D tex;

in vec3 vertColor;

out vec4 outputColor;

void main() {
    outputColor = vec4(vertColor.rgb, 1.0);
}