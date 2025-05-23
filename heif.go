// Package heif is a Go library and CLI tool to encode/decode HEIF/HEIC images without system dependencies (CGO).
package heif

/*
#include <stdlib.h>
#include <string.h>
#include <libheif/heif.h>

// Full decode: reads HEIF data from memory, gets the primary image,
// decodes it into an interleaved RGBA plane, and returns the heif_image*.
// Also returns the heif_context* and heif_image_handle* for cleanup.
struct heif_image* decode_heif_image(const uint8_t *data, size_t size,
                              struct heif_context **outCtx,
                              struct heif_image_handle **outHandle) {
    struct heif_context* ctx = heif_context_alloc();
    if (!ctx) return NULL;

    struct heif_error err = heif_context_read_from_memory(ctx, data, size, NULL);
    if (err.code != heif_error_Ok) {
        heif_context_free(ctx);
        return NULL;
    }

    struct heif_image_handle* handle = NULL;
    err = heif_context_get_primary_image_handle(ctx, &handle);
    if (err.code != heif_error_Ok) {
        heif_context_free(ctx);
        return NULL;
    }

    struct heif_image* img = NULL;
    // ask for interleaved RGBA
    err = heif_decode_image(handle, &img,
                            heif_colorspace_RGB,
                            heif_chroma_interleaved_RGBA,
                            NULL);
    if (err.code != heif_error_Ok) {
        heif_image_handle_release(handle);
        heif_context_free(ctx);
        return NULL;
    }

    if (outCtx)    *outCtx    = ctx;
    if (outHandle) *outHandle = handle;
    return img;
}

// get_heif_config: reads just enough of the HEIF file to extract width/height.
void get_heif_config(const uint8_t *data, size_t size,
                     uint32_t *width, uint32_t *height) {
    struct heif_context* ctx = heif_context_alloc();
    if (!ctx) {
        *width = 0;
        *height = 0;
        return;
    }

    struct heif_error err = heif_context_read_from_memory(ctx, data, size, NULL);
    if (err.code != heif_error_Ok) {
        *width = 0;
        *height = 0;
        heif_context_free(ctx);
        return;
    }

    struct heif_image_handle* handle = NULL;
    err = heif_context_get_primary_image_handle(ctx, &handle);
    if (err.code != heif_error_Ok) {
        *width = 0;
        *height = 0;
        heif_context_free(ctx);
        return;
    }

    *width  = (uint32_t)heif_image_handle_get_width(handle);
    *height = (uint32_t)heif_image_handle_get_height(handle);

    heif_image_handle_release(handle);
    heif_context_free(ctx);
}
*/
import "C"

import (
	"fmt"
	"image"
	"image/color"
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

	// Set the image quality
	var errQ C.struct_heif_error
	if options.Quality < 100 {
		errQ = C.heif_encoder_set_lossy_quality(encoder, C.int(options.Quality))
	} else {
		errQ = C.heif_encoder_set_lossless(encoder, C.int(1))
	}

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

func decodeHEIFToRGBA(data []byte) (*image.RGBA, error) {
	// Copy the Go slice into C memory
	cData := C.CBytes(data)
	defer C.free(cData)

	// Call our C helper
	var ctx *C.struct_heif_context
	var handle *C.struct_heif_image_handle
	img := C.decode_heif_image((*C.uint8_t)(cData), C.size_t(len(data)), &ctx, &handle)
	if img == nil {
		return nil, fmt.Errorf("failed to decode HEIF image")
	}

	// Query width/height from the interleaved plane
	width := int(C.heif_image_get_width(img, C.heif_channel_interleaved))
	height := int(C.heif_image_get_height(img, C.heif_channel_interleaved))

	// Grab a pointer to the RGBA data and its stride
	var cStride C.int
	ptr := C.heif_image_get_plane_readonly(img, C.heif_channel_interleaved, &cStride)
	rowBytes := int(cStride)

	// Allocate our Go RGBA
	goImg := image.NewRGBA(image.Rect(0, 0, width, height))

	// Copy row by row (width*4 bytes per row)
	for y := 0; y < height; y++ {
		// compute an *unsafe.Pointer* to the start of row y
		rowPtr := unsafe.Pointer(uintptr(unsafe.Pointer(ptr)) + uintptr(y*rowBytes))

		// now pass that directly into C.GoBytes
		chunk := C.GoBytes(rowPtr, C.int(width*4))

		dstOff := y * goImg.Stride
		copy(goImg.Pix[dstOff:dstOff+width*4], chunk)
	}

	// Cleanup C resources
	C.heif_image_release(img)
	C.heif_image_handle_release(handle)
	C.heif_context_free(ctx)

	return goImg, nil
}

// DecodeConfig reads enough of data to determine the image's configuration (dimensions, etc.).
// Here we read the entire data and call a lightweight C function that only parses the header.
func decodeConfig(data []byte) (image.Config, error) {
	if len(data) == 0 {
		return image.Config{}, fmt.Errorf("empty data buffer")
	}

	var w, h C.uint32_t
	C.get_heif_config(
		(*C.uint8_t)(unsafe.Pointer(&data[0])),
		C.size_t(len(data)),
		&w,
		&h,
	)

	if w == 0 || h == 0 {
		return image.Config{}, fmt.Errorf("failed to get HEIF image config")
	}

	return image.Config{
		ColorModel: color.RGBAModel,
		Width:      int(w),
		Height:     int(h),
	}, nil
}
