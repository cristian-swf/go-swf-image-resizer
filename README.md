# go-swf-image-resizer
A simple API that provides an endpoint to resize images to a given size and format


# Routes
* GET **/ping** - replies with "pong" text - useful for external monitoring apps like Zabbix
* GET **/status** - returns a JSON with the API status, up time and number of converted images and errors
* GET **/appVersions** - returns a JOSN with app version 
* POST **/resize** - resizes an image transmited via img POST param, together with width, height, format and quality parameters

## /resize parameters
* **img** (mandatory) - base64 encoded image
* **width** (mandatory) - output image width
* **height** (mandatory) - output image height
* **quality** (optional) - output image quality.\
Possible values: between 1 and 100.\
Default value: 75

* **format** (optional) - output image format.\
Possible values: png, jpg or webp.\
Default value: same as input image

## /resize output
A JSON with output image params and image encoded in base64
Sample output:
```
{
    "status":       "ok",
    "format":       "webp",
    "width":        100,
    "height":       100,
    "quality":      10,
    "resizedImage": "UklGRi..."
}
```

# Postman
You can find a Postman collection in this Repo