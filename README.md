# Pixlserv

A Go server for processing and serving images.

Images are requested from the server by accessing a URL of the following format: `http://server/parameters/filename`. Parameters are strings like `transformation_value` connected with commas, e.g. `w_400,h_300`. A full URL could look like this: `http://pixlserv.com/w_400,h_300/logo.jpg`. Authorisation can be easily set up to require an API key between `server` and `parameters` in the example URL above.


## Installation

TODO - test this out

```
git clone https://github.com/ReshNesh/pixlserv
go build
```

## Usage

Start redis (see the Requirements section below for details). Create a directory `local-images` with some JPEG or PNG images in the same directory where you installed pixlserv. Then run:

```
./pixlserv run config/example.yaml
```

This will run the server using a simple configuration defined by [config/example.yaml](config/example.yaml). You are encouraged to look at the Configuration section below, create a copy of the sample configuration file and customise it to suit your needs.

Assuming you copied a file `cat.jpg` to the `local-images` directory you can now access [http://localhost:3000/image/t_square/cat.jpg](http://localhost:3000/image/t_square/cat.jpg) using your browser.

## Configuration

Pixlserv supports 2 types of underlying storage: local file system and Amazon S3. If environment variables `AWS_ACCESS_KEY_ID`, `AWS_SECRET_ACCESS_KEY` and `PIXLSERV_S3_BUCKET` are detected the server will try to connect to S3 given the given credentials. Otherwise, local storage will be used. The path at which images will be stored locally can be specified using the `local-path` configuration option.

[//]: # (TODO: more info)
Other configuration options include `throttling-rate`, `allow-custom-transformations`, `allow-custom-scale`, `async-uploads`, `authorisation`, `cache`, `jpeg-quality`, `transformations` and `upload-max-file-size`. See [config/example.yaml](config/example.yaml) for an example.


## Supported transformations

### Resizing

| Parameter | Meaning                       |
| --------- | ----------------------------- |
| h_X       | sets height of the image to X |
| w_X       | sets width of the image to X  |


### Cropping

| Parameter | Meaning                                                                                                       |
| --------- | ------------------------------------------------------------------------------------------------------------- |
| c_e       | exact, image scaled exactly to given dimensions (default)                                                     |
| c_a       | all, the whole image will be visible in a frame of given dimensions, retains proportions                      |
| c_p       | part, part of the image will be visible in a frame of given dimensions, retains proportions, optional gravity |
| c_k       | keep scale, original scale of the image preserved, optional gravity                                           |


### Gravity

For some cropping modes gravity determines which part of the image will be shown.

| Parameter | Meaning                         |
| --------- | ------------------------------- |
| g_n       | north, top edge                 |
| g_ne      | north east, top-right corner    |
| g_e       | east, right edge                |
| g_se      | south east, bottom-right corner |
| g_s       | south, bottom edge              |
| g_sw      | south west, bottom-left corner  |
| g_w       | west, left edge                 |
| g_nw      | north west, top-left corner     |
| g_c       | center                          |


### Filter/colouring

| Parameter   | Meaning   |
| ----------- | --------- |
| f_grayscale | grayscale |


### Scaling

Scales the image up to support retina devices. For example to generate a thumbnail of an image (`image.jpg`) at twice the size request `image@2x.jpg`. Only positive integers are accepted as valid scaling factors.


### Named transformations

In your configuration file you can specify transformations using parameters described above and then give each transformation a name. The transformation can then be invoked using a `t_mytransformation` URL parameter.

Named transformations can also be set to be `eager`. Such transformations will be run for all images uploaded using the server straight after the upload happens.


## Requirements

A running [redis](http://redis.io/) instance is required for the server to be able to maintain a cache of images. Check the redis website to find out how to download and install redis. If you run redis on a different port than the default 6379 please make sure to set up a `PIXLSERV_REDIS_PORT` environment variable with the port you are using.
