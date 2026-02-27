package app

import (
	"github.com/go-gl/glfw/v3.3/glfw"
	"github.com/go-gl/mathgl/mgl32"
)

type Camera struct {
	window                   *glfw.Window
	eye                      mgl32.Vec3
	view                     mgl32.Vec3
	up                       mgl32.Vec3
	projection               mgl32.Mat4
	prevScreenX, prevScreenY float32
}

func newCamera(initialPos mgl32.Vec3, window *glfw.Window) *Camera {
	c := &Camera{}
	c.eye = initialPos
	c.view = mgl32.Vec3{0, 0, -1}
	c.up = mgl32.Vec3{0, 1, 0}
	c.projection = mgl32.Perspective(mgl32.DegToRad(45.0), float32(windowWidth)/windowHeight, 0.1, 100.0)
	c.window = window
	return c
}

func (c *Camera) SetLookHandler() {
	c.window.SetCursorPosCallback(func(w *glfw.Window, xpos, ypos float64) {
		c.Look(float32(xpos), float32(ypos))
	})
}

func (c *Camera) HandleMove() {
	var rightMove float32
	var forwardMove float32

	if c.window.GetKey(glfw.KeyA) == glfw.Press {
		rightMove--
	}
	if c.window.GetKey(glfw.KeyD) == glfw.Press {
		rightMove++
	}
	if c.window.GetKey(glfw.KeyW) == glfw.Press {
		forwardMove++
	}
	if c.window.GetKey(glfw.KeyS) == glfw.Press {
		forwardMove--
	}
	c.Move(forwardMove, rightMove)
}

func (c *Camera) Move(forward, right float32) {
	rightVec := c.view.Cross(c.up)

	movement := c.view.Mul(forward).Add(rightVec.Mul(right))
	if movement.Len() > 0 {
		movement = movement.Normalize()
	}

	var speed float32 = 0.1
	movement = movement.Mul(speed)
	c.eye = c.eye.Add(movement)
}

func (c *Camera) Look(screenX, screenY float32) {
	deltaX := -screenX + c.prevScreenX
	deltaY := -screenY + c.prevScreenY
	c.prevScreenX = screenX
	c.prevScreenY = screenY

	var speedX float32 = 0.1
	var speedY float32 = 0.05

	rotationX := mgl32.HomogRotate3D(speedX*mgl32.DegToRad(deltaX), c.up)
	dir := mgl32.Vec4{
		c.view.X(),
		c.view.Y(),
		c.view.Z(),
		0,
	}
	yaxis := c.up.Cross(c.view)
	rotationY := mgl32.HomogRotate3D(speedY*mgl32.DegToRad(deltaY), yaxis)
	rotation := rotationX.Mul4(rotationY)

	c.view = rotation.Mul4x1(dir).Vec3().Normalize()
}
