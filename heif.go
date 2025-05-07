package heif

/*
#include <stdlib.h>
#include <string.h>
#include <libheif/heif.h>
*/
import "C"

import (
	"fmt"
	"image"
	"os"
	"unsafe"
)

func encodeHEIF(rgba image.RGBA, options Options) ([]byte, error) {
	width := rgba.Bounds().Dx()
	height := rgba.Bounds().Dy()

	// Create the libheif context
	ctx := C.heif_context_alloc()
	defer C.heif_context_free(ctx)

	// Create an heicImage for the output
	var heicImage *C.struct_heif_image
	errCreate := C.heif_image_create(C.int(width), C.int(height), C.heif_colorspace_RGB, C.heif_chroma_interleaved_RGBA,
		&heicImage)

	if errCreate.code != C.heif_error_Ok {
		return nil, fmt.Errorf("failed to create HEIC image: %v", C.GoString(errCreate.message))
	}

	defer C.heif_image_release(heicImage)

	// Allocate the RGBA plane (8 bits)
	errPlane := C.heif_image_add_plane(heicImage, C.heif_channel_interleaved, C.int(width), C.int(height), C.int(8))
	if errPlane.code != C.heif_error_Ok {
		return nil, fmt.Errorf("failed to add RGBA plane to HEIC image: %v", C.GoString(errPlane.message))
	}

	// Copy the pixels
	var stride C.int
	ptr := C.heif_image_get_plane(heicImage, C.heif_channel_interleaved, &stride)
	size := C.size_t(stride) * C.size_t(height)
	C.memcpy(unsafe.Pointer(ptr), unsafe.Pointer(&rgba.Pix[0]), size)

	// Pick & configure HEVC encoder
	var encoder *C.struct_heif_encoder
	errEnc := C.heif_context_get_encoder_for_format(ctx, C.heif_compression_HEVC, &encoder)
	if errEnc.code != C.heif_error_Ok {
		return nil, fmt.Errorf("failed to create HEIC encoder: %v", C.GoString(errEnc.message))
	}

	defer C.heif_encoder_release(encoder)

	errQ := C.heif_encoder_set_lossy_quality(encoder, C.int(options.ColorQuality))
	if errQ.code != C.heif_error_Ok {
		return nil, fmt.Errorf("failed to set the image quality: %v", C.GoString(errQ.message))
	}

	// Encode into the context
	var handle *C.struct_heif_image_handle
	errImg := C.heif_context_encode_image(ctx, heicImage, encoder, nil, &handle)
	if errImg.code != C.heif_error_Ok {
		return nil, fmt.Errorf("failed to encode HEIC image: %v", C.GoString(errImg.message))
	}

	defer C.heif_image_handle_release(handle)

	// Write the output to a temporary file
	// TODO: This is a temporary solution because libheif seems to be very happy write the output to a file, instead of
	//       writing it to a buffer. I will need to investigate this further on how to write the output to memory.
	tmp, err := os.CreateTemp("", "heif-go-*.heic")
	if err != nil {
		return nil, err
	}

	// Close and delete the temp file
	tmpName := tmp.Name()
	tmp.Close()
	defer os.Remove(tmpName)

	cName := C.CString(tmpName)
	defer C.free(unsafe.Pointer(cName))

	errW := C.heif_context_write_to_file(ctx, cName)
	if errW.code != C.heif_error_Ok {
		return nil, fmt.Errorf("failed to write temp file: %v", C.GoString(errW.message))
	}

	// Slurp it back into memory
	data, err := os.ReadFile(tmpName)
	if err != nil {
		return nil, err
	}

	return data, nil
}
