# OCR API for Thai National ID Card

This API allows you to extract information from Thai National ID cards using Google Vision API. The extracted information includes ID card number, name, surname, date of birth, address, date of issue, and date of expiry.

## Getting Started

### Prerequisites

- Go 1.18 or later
- Google Cloud Vision API Key

### Installation

1. Clone the repository:

    ```sh
    git clone https://github.com/yourusername/ocr-id-card.git
    cd ocr-id-card
    ```

2. Install dependencies:

    ```sh
    go mod tidy
    ```

3. Set up your Google Cloud Vision API key:

    ```sh
    export GOOGLE_VISION_API_KEY=your_api_key
    ```

### Usage

1. Build and run the application:

    ```sh
    go run main.go
    ```

2. Send a POST request to the `/ocr` endpoint with a JSON payload containing the base64-encoded image data. For example:

    ```sh
    curl -X POST -H "Content-Type: application/json" -d '{"image": "base64_image_data"}' http://localhost:8080/ocr
    ```

3. The API will respond with the extracted information:

    ```json
    {
        "id_card_number": "1111111111111",
        "name": "ssss",
        "last_name": "sssss",
        "date_of_birth": "22 Aug 1992",
        "address": "xxxxxxxxxxxxx",
        "date_of_issue": "13 Mar 2023",
        "date_of_expiry": "20 Aug 2031"
    }
    ```

### API Endpoints

#### POST /ocr

Extract information from a Thai National ID card.

- **URL**: `/ocr`
- **Method**: `POST`
- **Request Body**:
  - `image` (string)
- **Response**:
  - `id_card_number` (string): ID card number.
  - `name` (string): First name.
  - `last_name` (string): Last name.
  - `date_of_birth` (string): Date of birth.
  - `address` (string): Address.
  - `date_of_issue` (string): Date of issue.
  - `date_of_expiry` (string): Date of expiry.
x
