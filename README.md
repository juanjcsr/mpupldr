[![Build Status](https://travis-ci.org/juanjcsr/mpupldr.svg?branch=master)](https://travis-ci.org/juanjcsr/mpupldr)

# mpupldr

**mpupldr** is a Mapbox tile uploader

## Requirements

- Tippecanoe

## Run it with Docker

To run **mpupldr** for multiple geojson files within a directory:

```
docker run --rm -it -v `pwd`/path/to/geojsonfiles:/mapbox juanjcsr/mpupldr uploadDir /mapbox -m sk.secret_mapbox_token -u mapbox_username
```