package icon

import (
	"bytes"
	"encoding/binary"
	"image"
	"image/png"
	"io/fs"

	"go-downloader/web"
)

// LoadLogoICO 从嵌入资源加载 logo.png 并转换为 ICO 格式
func LoadLogoICO() []byte {
	pngData, err := web.Assets.ReadFile("assets/logo.png")
	if err != nil {
		return fallbackIcon()
	}

	img, err := png.Decode(bytes.NewReader(pngData))
	if err != nil {
		return fallbackIcon()
	}

	scaled := scaleImage(img, 32)

	var buf bytes.Buffer
	png.Encode(&buf, scaled)
	scaledPNG := buf.Bytes()

	return pngToIco(scaledPNG, 32, 32)
}

// LoadLogoPNG 从嵌入资源加载 logo.png 并缩放为指定尺寸的 PNG
func LoadLogoPNG(size int) []byte {
	pngData, err := web.Assets.ReadFile("assets/logo.png")
	if err != nil {
		return nil
	}

	img, err := png.Decode(bytes.NewReader(pngData))
	if err != nil {
		return nil
	}

	scaled := scaleImage(img, size)
	var buf bytes.Buffer
	png.Encode(&buf, scaled)
	return buf.Bytes()
}

// scaleImage 将图片缩放到指定尺寸
func scaleImage(src image.Image, size int) image.Image {
	dst := image.NewRGBA(image.Rect(0, 0, size, size))
	srcBounds := src.Bounds()
	sw, sh := srcBounds.Dx(), srcBounds.Dy()
	for y := 0; y < size; y++ {
		for x := 0; x < size; x++ {
			sx := x * sw / size
			sy := y * sh / size
			dst.Set(x, y, src.At(srcBounds.Min.X+sx, srcBounds.Min.Y+sy))
		}
	}
	return dst
}

// pngToIco 将 PNG 数据包装为 ICO 文件格式
func pngToIco(pngData []byte, width, height int) []byte {
	size := len(pngData)
	ico := make([]byte, 6+16+size)

	binary.LittleEndian.PutUint16(ico[0:], 0)
	binary.LittleEndian.PutUint16(ico[2:], 1)
	binary.LittleEndian.PutUint16(ico[4:], 1)

	ico[6] = byte(width)
	ico[7] = byte(height)
	ico[8] = 0
	ico[9] = 0
	binary.LittleEndian.PutUint16(ico[10:], 1)
	binary.LittleEndian.PutUint16(ico[12:], 32)
	binary.LittleEndian.PutUint32(ico[14:], uint32(size))
	binary.LittleEndian.PutUint32(ico[18:], 22)

	copy(ico[22:], pngData)
	return ico
}

// fallbackIcon 生成兜底图标：16x16 黑色方块 ICO
func fallbackIcon() []byte {
	const s = 16
	img := image.NewRGBA(image.Rect(0, 0, s, s))
	for y := 2; y < s-2; y++ {
		for x := 2; x < s-2; x++ {
			img.Set(x, y, image.Black)
		}
	}
	var buf bytes.Buffer
	png.Encode(&buf, img)
	return pngToIco(buf.Bytes(), s, s)
}

// 确保 fs 包被引用
var _ fs.FileInfo
