# Image Processing API Specification

This specification outlines the details for an image processing API, including resizing, formatting, effects, and image delivery.

## API Overview

The API processes images via URL query parameters, supporting operations like resizing, format conversion, quality adjustments, and effects such as blurring.

### Endpoint

Base URL:

https://api.example.com/v1/image

---

## Query Parameters

### 1. **Image Sizing**

- **`w` (Width)**:

  - **Description**: Sets the output image width in pixels.
  - **Type**: Integer.
  - **Default**: Original image width.
  - **Example**: `w=300`.
  - **Reference**: [Image Width](https://docs.imgix.com/en-US/apis/rendering/size/image-width).

- **`h` (Height)**:

  - **Description**: Sets the output image height in pixels.
  - **Type**: Integer.
  - **Default**: Original image height.
  - **Example**: `h=200`.
  - **Reference**: [Image Height](https://docs.imgix.com/en-US/apis/rendering/size/image-height).

- **`fit` (Resize Mode)**:

  - **Description**: Specifies how the image should be resized to fit the given dimensions.
  - **Type**: Enum (`clip`, `crop`, `fill`).
  - **Default**: `clip`.
  - **Example**: `fit=crop`.
  - **Reference**: [Resize Fit Mode](https://docs.imgix.com/en-US/apis/rendering/size/resize-fit-mode).

- **`dpr` (Device Pixel Ratio)**:
  - **Description**: Adjusts the resolution for high-density displays (e.g., Retina screens).
  - **Type**: Float (1.0–3.0).
  - **Default**: 1.0.
  - **Example**: `dpr=2`.
  - **Reference**: [Device Pixel Ratio](https://docs.imgix.com/en-US/apis/rendering/device-pixel-ratio).

---

### 2. **Image Formatting**

- **`fm` (Format)**:

  - **Description**: Specifies the output format of the image.
  - **Type**: Enum (`jpg`, `png`, `webp`).
  - **Default**: Retains the original format.
  - **Example**: `fm=webp`.
  - **Reference**: [Output Format](https://docs.imgix.com/en-US/apis/rendering/format/output-format).

- **`q` (Quality)**:
  - **Description**: Adjusts the compression quality for supported formats.
  - **Type**: Integer (0–100).
  - **Default**: 80.
  - **Example**: `q=80`.
  - **Reference**: [Output Quality](https://docs.imgix.com/en-US/apis/rendering/format/output-quality).

---

### 3. **Effects**

- **`blur` (Gaussian Blur)**:
  - **Description**: Applies a blur effect to the image.
  - **Type**: Integer (0–100).
  - **Default**: 0 (no blur).
  - **Example**: `blur=15`.
  - **Reference**: [Gaussian Blur](https://docs.imgix.com/en-US/apis/rendering/stylize/gaussian-blur).

---

### 4. **Image Delivery**

- **`dl` (Force Download)**:
  - **Description**: Forces the image to be downloaded instead of displayed.
  - **Type**: Boolean (`1` for true).
  - **Default**: `false`.
  - **Example**: `dl=1`.

---

## Usage Examples

### 1. Resize to Specific Dimensions

https://api.example.com/v1/image?url=https://example.com/image.jpg?w=300&h=200&fit=crop

### 2. Convert Format and Adjust Quality

https://api.example.com/v1/image?url=https://example.com/image.jpg?fm=webp&q=80

### 3. Apply Blur Effect with High Resolution

https://api.example.com/v1/image?url=https://example.com/image.jpg?blur=15&dpr=2

### 4. Force Download

https://api.example.com/v1/image?url=https://example.com/image.jpg?dl=1

---

## API Response

- **Success**: Returns the processed image.
  - **HTTP Status**: `200 OK`.
  - **Content-Type**: Depends on the `fm` parameter (`image/jpeg`, `image/png`, `image/webp`).
- **Error**:
  - **HTTP Status**: `400 Bad Request` (invalid parameters).
  - **HTTP Status**: `404 Not Found` (image not found).

---
