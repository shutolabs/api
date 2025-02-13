# [2.1.0](https://github.com/lgastler/shuto-api/compare/v2.0.1...v2.1.0) (2025-02-13)


### Bug Fixes

* **docker:** pin vips and rclone package versions in Dockerfile ([3cb8d92](https://github.com/lgastler/shuto-api/commit/3cb8d9277151bd10f16545e03e41cba75c193843))


### Features

* **cors:** add CORS middleware to API endpoints ([f09cac1](https://github.com/lgastler/shuto-api/commit/f09cac110501848877ff9166db629f0ff8407bec))

## [2.0.1](https://github.com/lgastler/shuto-api/compare/v2.0.0...v2.0.1) (2025-02-12)


### Bug Fixes

* **docker:** improve container configuration and deployment ([fbaaeec](https://github.com/lgastler/shuto-api/commit/fbaaeec07c2435f205a53b67012e6e4c9e5b09ee))
* **docker:** remove unnecessary public folder copy in Dockerfile ([b64a027](https://github.com/lgastler/shuto-api/commit/b64a02771171f0cd77382dc48fa3e9166536a813))

# [2.0.0](https://github.com/lgastler/shuto-api/compare/v1.3.0...v2.0.0) (2025-02-12)


* feat(errors)!: standardize error responses across all endpoints ([924b620](https://github.com/lgastler/shuto-api/commit/924b6208f810d90718100c45cc265c6ed2c31aae))


### Bug Fixes

* **image,list:** improve error handling and API key validation ([2519598](https://github.com/lgastler/shuto-api/commit/25195983e712899ed3bbf5258a0296da33d9b7df))


### Features

* **docs:** update Swagger documentation for API endpoints ([2034b68](https://github.com/lgastler/shuto-api/commit/2034b68c2ff5bce1149e83f327990246d19a83d5))
* **security:** add API key authentication for list endpoint ([95afc36](https://github.com/lgastler/shuto-api/commit/95afc36bf68cafa130419f2e26bc6aca93abff90))


### BREAKING CHANGES

* Error responses now return a structured JSON object instead of plain text.
The new format includes an error message, error code, and optional details (in development only).

Old format:
"Invalid domain"

New format:
{
  "error": "Invalid domain",
  "code": "INVALID_DOMAIN",
  "details": "..." //
}

Error codes have been standardized across all endpoints:
- INVALID_REQUEST
- UNAUTHORIZED
- FORBIDDEN
- NOT_FOUND
- INTERNAL_ERROR
- INVALID_DOMAIN
- INVALID_PATH
- INVALID_API_KEY
- EXPIRED_TOKEN
- INVALID_SIGNATURE

# [1.3.0](https://github.com/lgastler/shuto-api/compare/v1.2.0...v1.3.0) (2025-02-11)


### Bug Fixes

* **cli:** restructure CLI with command pattern and modular design ([6c9ffb2](https://github.com/lgastler/shuto-api/commit/6c9ffb2dd29b0d14849c2f067ff06d186faa8606))
* **image:** prevent serving directories as images ([51238c7](https://github.com/lgastler/shuto-api/commit/51238c77135ed94768405de74dbd5d0ccfc29a62))
* **images:** extract image transformation logic and add utility functions ([c94af8b](https://github.com/lgastler/shuto-api/commit/c94af8b7b887a1031643085c8bc9e3fff7646978))
* **tests:** add domain config mocking for image handler tests ([7a800f8](https://github.com/lgastler/shuto-api/commit/7a800f88723676d659f30d455ddc87a39ee4e812))
* **tests:** update test cases for download and image handlers ([1900517](https://github.com/lgastler/shuto-api/commit/19005175ea21df34503199d72a9ffe72c31297a8))


### Features

* **cli:** add CLI tool for generating signed URLs ([92bd0d6](https://github.com/lgastler/shuto-api/commit/92bd0d6b9f1a6a77732e98a4b3f96510b7cee2c8))
* **download:** implement parallel file processing for folder downloads ([d506e18](https://github.com/lgastler/shuto-api/commit/d506e188c6c308e3e40617c7bd7a926accecff98))
* **image:** add automatic format selection based on browser support ([7244828](https://github.com/lgastler/shuto-api/commit/7244828ab1124da2e91bb7520cda072c5d0a419b))
* **security:** add endpoint parameter to URL signer ([d6951ee](https://github.com/lgastler/shuto-api/commit/d6951ee248aa5a13356406c24266c4af98c4a486))
* **security:** add signed URL validation for download endpoint ([bcaccc5](https://github.com/lgastler/shuto-api/commit/bcaccc5134b634c75a8e30818dc22a4c3bd9bffe))
* **security:** add signed URL validation for image endpoint ([a6143f7](https://github.com/lgastler/shuto-api/commit/a6143f78a061a5e3aca0157e8eb4e2f137c3bf7d))

# [1.2.0](https://github.com/lgastler/shuto-api/compare/v1.1.1...v1.2.0) (2025-01-09)


### Bug Fixes

* **image-metadata:** return keywords if only one ([16ac216](https://github.com/lgastler/shuto-api/commit/16ac216b4e5149e4ca2372b40b93cd514a64cf84))


### Features

* **download:** added download endpoint ([3e71c7f](https://github.com/lgastler/shuto-api/commit/3e71c7ffa1f15c819923d8fa30a314e6e1e5982b))
* **image-metadata:** added keywords to image metadata ([8724591](https://github.com/lgastler/shuto-api/commit/87245910603a92d06df4c9e0f4f7f0977c78b016))

## [1.1.1](https://github.com/lgastler/shuto-api/compare/v1.1.0...v1.1.1) (2024-12-20)


### Bug Fixes

* **cache:** added cache for image dimensions ([72cd03e](https://github.com/lgastler/shuto-api/commit/72cd03ef2a8738ab3cf87b93deb5c9c6daaa9fd7))
* **list-res:** added omitempty to omit 0 ([aee03cd](https://github.com/lgastler/shuto-api/commit/aee03cd9b38e95a031ef43f376bcb0b42286ad13))

# [1.1.0](https://github.com/lgastler/shuto-api/compare/v1.0.3...v1.1.0) (2024-12-20)


### Features

* **image-dimensions:** dimensions added for list endpoint ([cae3891](https://github.com/lgastler/shuto-api/commit/cae38918b16b6d41fc379247f830da0759c5745a))

## [1.0.3](https://github.com/lgastler/shuto-api/compare/v1.0.2...v1.0.3) (2024-12-19)


### Bug Fixes

* **ci:** removed ecr image ([2db2acc](https://github.com/lgastler/shuto-api/commit/2db2acccc6e836c24306e4d7855d78f4ade0c646))

## [1.0.2](https://github.com/lgastler/shuto-api/compare/v1.0.1...v1.0.2) (2024-12-19)


### Bug Fixes

* **tests:** with no configured logger ([274f258](https://github.com/lgastler/shuto-api/commit/274f258e1d212bb89858077df95037a9d5f25383))

## [1.0.1](https://github.com/lgastler/shuto-api/compare/v1.0.0...v1.0.1) (2024-12-12)


### Bug Fixes

* **config:** update config for all domains ([9582916](https://github.com/lgastler/shuto-api/commit/958291670916afc07da8781b9e9328c6dbdb13dd))

# 1.0.0 (2024-12-12)


### Bug Fixes

* **build:** copy public folder to image ([3f2b29c](https://github.com/lgastler/shuto-api/commit/3f2b29c0c0dc4938a51a3398dcae75f0c2ecf8c6))
* **image:** copy config to image ([2545fb1](https://github.com/lgastler/shuto-api/commit/2545fb1bbd81ebb7cc11634cf2dd61b85fdb5b5f))
* **image:** fix image fetch ([0aac30f](https://github.com/lgastler/shuto-api/commit/0aac30f7e1cab809e17215f1026dfa1530332104))
* **routes:** fixed list route config ([a8cefeb](https://github.com/lgastler/shuto-api/commit/a8cefeb3f12a1a70c6ef89faf16393006581b1ba))


### Features

* **ci:** add ci ([fb02087](https://github.com/lgastler/shuto-api/commit/fb02087332382438606b80b6d0686769af9d4df3))
* **ci:** added deploy workflow ([1574b40](https://github.com/lgastler/shuto-api/commit/1574b40d264be405b12c1dbfc436a6bffa614277))
* **ci:** setup release action ([d3fdf42](https://github.com/lgastler/shuto-api/commit/d3fdf42083348eedacfd1f1ae80c4848805d6a26))
* **config:** added rclone config options ([bf93bea](https://github.com/lgastler/shuto-api/commit/bf93beaefb68665249f655262879db2c91c20054))
* **config:** enable config per domain ([f000efb](https://github.com/lgastler/shuto-api/commit/f000efbfeb2153bc903464a8ced18c097b88361d))
* **docker:** setup docker ([ce3d462](https://github.com/lgastler/shuto-api/commit/ce3d462e7ee2658c9958b8f33b4ada6cd45522c1))
* **spec:** added specification and swaggerui ([649919f](https://github.com/lgastler/shuto-api/commit/649919fca918ec26bd46653e9f02f43ff05097d5))
* **tests:** add basic test suite ([b04f094](https://github.com/lgastler/shuto-api/commit/b04f09445db0fd4d67c887b603b86a82f8d94b78))
* **transform:** added more image transform options ([16b6260](https://github.com/lgastler/shuto-api/commit/16b6260c2452ba0ac8d412ad342d5b25bdf3143a))
