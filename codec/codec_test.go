package codec_test

import "testing"

func TestCodec(t *testing.T) {
	t.Run("testEncoderOpus", testEncoderOpus)
	t.Run("testDecoderOpus", testDecoderOpus)
}
