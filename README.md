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
        "id_card_number": "1341800066789",
        "name": "Naruchet",
        "last_name": "Phicharattanachai",
        "date_of_birth": "21 Aug 1991",
        "address": "999/80 หมู่ที่ 3 ต.ตลาด อ.เมืองนครราชสีมา จ.นครราชสีมา",
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
  - `image` (string): Base64-encoded image data.
- **Response**:
  - `id_card_number` (string): ID card number.
  - `name` (string): First name.
  - `last_name` (string): Last name.
  - `date_of_birth` (string): Date of birth.
  - `address` (string): Address.
  - `date_of_issue` (string): Date of issue.
  - `date_of_expiry` (string): Date of expiry.

### Example Code

Here is a basic implementation of the main.go file:

```go
package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"regexp"
	"strings"
	"io/ioutil"
)

type OCRRequest struct {
	Image string `json:"image"`
}

type OCRResponse struct {
	IDCardNumber  string `json:"id_card_number"`
	Name          string `json:"name"`
	LastName      string `json:"last_name"`
	DateOfBirth   string `json:"date_of_birth"`
	Address       string `json:"address"`
	DateOfIssue   string `json:"date_of_issue"`
	DateOfExpiry  string `json:"date_of_expiry"`
}

func extractIDCardNumber(line string) string {
	re := regexp.MustCompile(`\d{1} \d{4} \d{5} \d{2} \d{1}`)
	match := re.FindString(line)
	return strings.ReplaceAll(match, " ", "")
}

func extractName(line string) (string, string) {
	re := regexp.MustCompile(`ชื่อตัวและชื่อสกุล\s*(นาย|นาง|นางสาว)\s*([^\s]+)\s*([^\s]+)`)
	matches := re.FindStringSubmatch(line)
	if len(matches) == 4 {
		return matches[2], matches[3]
	}
	return "", ""
}

func extractDate(line string) string {
	re := regexp.MustCompile(`\d{1,2} \w+ \d{4}`)
	return re.FindString(line)
}

func extractAddress(lines []string) string {
	startIndex := -1
	for i, line := range lines {
		if strings.Contains(line, "ที่อยู่") {
			startIndex = i
			break
		}
	}
	if startIndex != -1 {
		return strings.Join(lines[startIndex:startIndex+2], " ")
	}
	return ""
}

func parseOCRDescription(description string) OCRResponse {
	lines := strings.Split(description, "\n")
	var parsed OCRResponse
	var buffer string

	for _, line := range lines {
		line = strings.TrimSpace(line)
		buffer += line + " "

		if strings.Contains(line, "เลขประจำตัวประชาชน") || strings.Contains(line, "Identification Number") {
			parsed.IDCardNumber = extractIDCardNumber(buffer)
			buffer = ""
		} else if strings.Contains(line, "ชื่อตัวและชื่อสกุล") || strings.Contains(line, "Name") {
			parsed.Name, parsed.LastName = extractName(buffer)
			buffer = ""
		} else if strings.Contains(line, "เกิดวันที่") || strings.Contains(line, "Date of Birth") {
			parsed.DateOfBirth = extractDate(buffer)
			buffer = ""
		} else if strings.Contains(line, "ที่อยู่") {
			parsed.Address = extractAddress(lines)
		} else if strings.Contains(line, "วันออกบัตร") || strings.Contains(line, "Date of issue") {
			parsed.DateOfIssue = extractDate(buffer)
			buffer = ""
		} else if strings.Contains(line, "วันบัตรหมดอายุ") || strings.Contains(line, "Date of Expiry") {
			parsed.DateOfExpiry = extractDate(buffer)
			buffer = ""
		}
	}

	return parsed
}

func ocrHandler(w http.ResponseWriter, r *http.Request) {
	var req OCRRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	// Mocked response from Google Vision API
	mockedResponse := `{
		"responses": [
			{
				"textAnnotations": [
					{
						"locale": "th",
						"description": "บัตรประจำตัวประชาชน Thai National\nID Card\nเลขประจำตัวประชาชน\nIdentification Number\n1 3418 00066 78 9\nชื่อตัวและชื่อสกุล นาย นฤเชษ พิชารัตนชัย\nName Mr. Naruchet\nLast name Phicharattanachai\nเกิดวันที่ 21 ส.ค. 2534\n160.\nDate of Birth 21 Aug. 1991\nที่อยู่ 999/80 หมู่ที่ 3 ต.ตลาด อ.เมืองนครราชสีมา\nจ.นครราชสีมา\n13 มี.ค. 2566\nวันออกบัตร\n13 Mar. 2023\nDate of issue\n(นายแมนรัตน์ รัตนสุคนธ์)\nเจ้าพนักงานออกบัตร\n20 ส.ค. 2574\nวันบัตรหมดอายุ\n20 Aug, 2031\nDate of Expiry\n222\n160\n150\n150\n140\n140\n3001-05-03131655"
					}
				]
			}
		]
	}`

	var visionResponse map[string]interface{}
	err = json.Unmarshal([]byte(mockedResponse), &visionResponse)
	if err != nil {
		http.Error(w, "Error processing image", http.StatusInternalServerError)
		return
	}

	if responses, ok := visionResponse["responses"].([]interface{}); ok {
		if len(responses) > 0 {
			if annotations, ok := responses[0].(map[string]interface{})["textAnnotations"].([]interface{}); ok {
				if len(annotations) > 0 {
					description := annotations[0].(map[string]interface{})["description"].(string)
					parsedData := parseOCRDescription(description)
					w.Header().Set("Content-Type", "application/json")
					json.NewEncoder(w).Encode(parsedData)
					return
				}
			}
		}
	}

	http.Error(w, "No text annotations found", http.StatusNotFound)
}

func main() {
	http.HandleFunc("/ocr", ocrHandler)
	fmt.Println("Server started at :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
