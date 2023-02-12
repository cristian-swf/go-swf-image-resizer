package main
//Future improvements:
// - add option to set the interpolation method in the POST request
// - add more image formats
// - add more image quality options
// - add more image resize options
// - create a new route to simply convert from one image format to another
// - create a new route to validate if a base64-encoded image is valid
// - split code into modules
// - add unit tests

import (
	"bytes"
	"encoding/base64"
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/nfnt/resize"
	"github.com/kolesa-team/go-webp/encoder"
	"github.com/kolesa-team/go-webp/webp"
	"github.com/joho/godotenv"
	"image"
	"image/jpeg"
	"image/png"
	"net/http"
	"strconv"
	"os"
	"time"
)

var startTime time.Time

var appVersion = "" //app version

type ResizedImages struct {
	Png         int `json:"png"`
	Jpg        	int `json:"jpg"`
	WebP 		int `json:"webp"`
	Errors 		int `json:"errors"`
}

var resizedImages = ResizedImages{ 0, 0, 0, 0}

func uptime() time.Duration {
	return time.Duration(time.Since(startTime).Minutes())
}

func init() {
	println("Starting SWF Image resize API....")
	startTime = time.Now()

	//load .env file
	godotenv.Load()

	//get app version from env
	appVersion = os.Getenv("APP_VERSION")
}

func main() {
	r := gin.Default()

	r.POST("/resize", resizeImageHandler)
	r.GET("/ping", pingHandler)
	r.GET("/appVersion", appVersionHandler)
	r.GET("/status", statusHandler)

	r.Run()
}

func pingHandler(c *gin.Context) {
	c.String(http.StatusOK, "pong")
}

func appVersionHandler(c *gin.Context) {
	c.IndentedJSON(http.StatusOK, gin.H{"appVersion": appVersion})
}

func statusHandler(c *gin.Context) {
	c.IndentedJSON(http.StatusOK, gin.H{"status": "online", "uptime": uptime(), "resizedImages": resizedImages})
}

//handle image resize requests
func resizeImageHandler(c *gin.Context) {
	imgB64 := c.PostForm("img")
	if imgB64 == "" {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "img parameter is required"})
		resizedImages.Errors++ //increment error counter
		return
	}

	imgData, err := base64.StdEncoding.DecodeString(imgB64)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "invalid base64-encoded image"})
		resizedImages.Errors++ //increment error counter
		return
	}

	widthStr := c.PostForm("width")
	heightStr := c.PostForm("height")

	width, err := strconv.Atoi(widthStr)
	if err != nil || width < 10 || width > 2048 {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "invalid image width"})
		resizedImages.Errors++ //increment error counter
		return
	}

	height, err := strconv.Atoi(heightStr)
	if err != nil || height < 10 || height > 2048 {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "invalid image height"})
		resizedImages.Errors++ //increment error counter
		return
	}

	imgReader := bytes.NewReader(imgData)

	img, format, err := image.Decode(imgReader)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "unable to decode image"})
		resizedImages.Errors++ //increment error counter
		return
	}

	//TODO: provide the option to set interpolation method in the POST request
	resizedImg := resize.Resize(uint(width), uint(height), img, resize.Lanczos3)

	buf := new(bytes.Buffer)
	resizedFormat := c.PostForm("format")

	//if the resizedFormat is not set, use the original image format
	if resizedFormat == "" {
		resizedFormat = format;
	}

	//retrieve the output quality from the POST request
	outputQuality := c.PostForm("quality");
	if outputQuality == "" {
		outputQuality = "75";
	}

	outputQualityInt, err := strconv.Atoi(outputQuality)

	if err != nil || outputQualityInt < 1 || outputQualityInt > 100 {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "invalid image quality param"})
		resizedImages.Errors++ //increment error counter
		return
	}

	//convert the image to the requested format
	switch resizedFormat {
		case "jpeg":
			err = jpeg.Encode(buf, resizedImg, &jpeg.Options{Quality: outputQualityInt})
		case "png":
			err = png.Encode(buf, resizedImg)
		case "webp":
			webpOptions, err := encoder.NewLossyEncoderOptions(encoder.PresetDefault, float32(outputQualityInt) )
			if err != nil {
				c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "unable to encode webp format"})
				resizedImages.Errors++ //increment error counter
			}
			err = webp.Encode(buf, resizedImg, webpOptions)
		default:
			err = errors.New("unsupported image format: " + resizedFormat)
	}

	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "unable to encode image"})
		resizedImages.Errors++ //increment error counter
		return
	}

	resizedImgB64 := base64.StdEncoding.EncodeToString(buf.Bytes())

	switch resizedFormat {
		case "jpeg":
			resizedImages.Jpg++
		case "png":
			resizedImages.Png++
		case "webp":
			resizedImages.WebP++
		default:
			resizedImages.Errors++ //increment error counter
	}

	//output result
	c.IndentedJSON(http.StatusOK, gin.H{"status":"ok", "resizedImage": resizedImgB64, "width": width, "height": height, "format": resizedFormat, "quality": outputQualityInt})
}

