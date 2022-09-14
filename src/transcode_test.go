package png_test

import (
	"os"
	"testing"

	png "github.com/accrazed/png/src"
	"github.com/stretchr/testify/assert"
)

func TestNewTranscoder(t *testing.T) {

	t.Run("valid basic png file", func(t *testing.T) {
		f, err := os.Open("testlib/validbasic.png")
		assert.NoError(t, err)
		tc, err := png.NewTranscoder(f)
		assert.NoError(t, err)
		assert.Equal(t, uint32(800), tc.Width)
		assert.Equal(t, uint32(600), tc.Height)
		assert.Equal(t, png.BitDepth(8), tc.BitDepth)
		assert.Equal(t, png.ColorType(6), tc.ColorType)
		assert.Equal(t, map[string][]uint32{"IDAT": {33}, "IEND": {226921}, "IHDR": {8}}, tc.SeenSet)
	})
}
