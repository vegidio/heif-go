# heif-go

A Go library and CLI tool to encode/decode HEIF/HEIC images without system dependencies (CGO).

## 💡 Motivation

There are a couple of libraries to encode/decode HEIF images in Go, and even though they do the job well, they have one limitation that don't satisfy my needs: they either depend on libraries to be installed on the system to be built and/or later be executed.

**heif-go** uses CGO to create a static implementation of HEIF/HEIC, so you don't need to have `libheif` (or any of its sub-dependencies) installed to build or run your Go application.

It also runs on native code (supports `darwin/amd64`, `darwin/arm64`, `linux/amd64`, `linux/arm64`, `windows/amd64`), so it achieves the best performance possible.

## ⬇️ Installation

### Library

This library can be installed using Go modules. To do that, run the following command in your project's root directory:

```bash
$ go get github.com/vegidio/heif-go
```

### CLI

The binaries are available for Windows, macOS, and Linux. Download the [latest release](https://github.com/vegidio/heif-go/releases) that matches your computer architecture and operating system.

## 🤖 Usage

### Library

This is a CGO library, so to use it, you _must_ enable CGO while building your application. You can do that by setting the `CGO_ENABLED` environment variable to `1`:

```bash
$ CGO_ENABLED=1 go build /path/to/your/app.go
```

Here are some examples of how to encode and decode HEIC images using this library. These snippets don't have any error handling for the sake of simplicity, but you should always check for errors in production code.

#### Encoding

```go
var originalImage image.Image = ... // an image.Image to be encoded
heicFile, err := os.Create("/path/to/image.heic") // create the file to save the HEIC
err = heif.Encode(heicFile, originalImage, nil) // encode the image & save it to the file
```

#### Decoding

```go
import _ "github.com/vegidio/heif-go" // do a blank import to register the HEIC decoder
heicFile, err := os.Open("/path/to/image.heic") // open the HEIC file to be decoded
heicImage, _, err := image.Decode(heicFile) // decode the image
```

### CLI

If you want to decode an HEIC image, run the following command:

```bash
$ heic decode /path/to/image.heic /path/to/image.png
```

---

To encode an image to HEIC, run the following command:

```bash
$ heic encode /path/to/image.png /path/to/image.heic
```

For the full list of parameters, type `heic encode --help` in the terminal.

## 💣 Troubleshooting

### I cannot build my app after importing this library

If you cannot build your app after importing **heif-go**, it is probably because you didn't set the `CGO_ENABLED` environment variable to `1`.

You must either set a global environment variable with `export CGO_ENABLED=1` or set it in the command line when building your app with `CGO_ENABLED=1 go build /path/to/your/app.go`.

### "App Is Damaged..." (Unidentified Developer — macOS only)

For a couple of years now, Apple has required developers to join their "Developer Program" to gain the pretentious status of an _identified developer_ 😛.

Translating to non-BS language, this means that if you’re not registered with Apple (i.e., paying the fee), you can’t freely distribute macOS software. Apps from unidentified developers will display a message saying the app is damaged and can’t be opened.

To bypass this, open the Terminal and run the command below, replacing `<path-to-app>` with the correct path to where you’ve installed the app:

```bash
$ xattr -d com.apple.quarantine <path-to-app>
```

## 📝 License

**heif-go** is released under the MIT License. See [LICENSE](LICENSE) for details.

## 👨🏾‍💻 Author

Vinicius Egidio ([vinicius.io](http://vinicius.io))
