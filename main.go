package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"regexp"
	"strings"
)

type ImageSource struct {
	ImageUri string `json:"imageUri"`
}

type Image struct {
	Source ImageSource `json:"source"`
}

type Feature struct {
	Type string `json:"type"`
}

type Request struct {
	Image    Image     `json:"image"`
	Features []Feature `json:"features"`
}

type VisionRequest struct {
	Requests []Request `json:"requests"`
}

type OCRRequest struct {
	ImageUri string `json:"imageUri"`
}

type ParsedData struct {
	IDCardNumber string
	Name         string
	LastName     string
	DateOfBirth  string
	Address      string
	DateOfIssue  string
	DateOfExpiry string
}

func extractIDCardNumber(line string) string {
	re := regexp.MustCompile(`\d{1} \d{4} \d{5} \d{2} \d{1}`)
	match := re.FindString(line)
	return strings.ReplaceAll(match, " ", "")
}

func extractName(line string) (string, string) {
	parts := strings.Split(line, " ")
	if len(parts) >= 5 {
		return parts[3], parts[4]
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

func parseOCRDescription(description string) ParsedData {
	lines := strings.Split(description, "\n")
	var parsed ParsedData

	for _, line := range lines {
		line = strings.TrimSpace(line)

		if strings.Contains(line, "เลขประจำตัวประชาชน") || strings.Contains(line, "Identification Number") {
			parsed.IDCardNumber = extractIDCardNumber(line)
		} else if strings.Contains(line, "ชื่อตัวและชื่อสกุล") || strings.Contains(line, "Name") {
			parsed.Name, parsed.LastName = extractName(line)
		} else if strings.Contains(line, "เกิดวันที่") || strings.Contains(line, "Date of Birth") {
			parsed.DateOfBirth = extractDate(line)
		} else if strings.Contains(line, "ที่อยู่") {
			parsed.Address = extractAddress(lines)
		} else if strings.Contains(line, "วันออกบัตร") || strings.Contains(line, "Date of issue") {
			parsed.DateOfIssue = extractDate(line)
		} else if strings.Contains(line, "วันบัตรหมดอายุ") || strings.Contains(line, "Date of Expiry") {
			parsed.DateOfExpiry = extractDate(line)
		}
	}

	return parsed
}

type ThaiIDCard struct {
	IDNumber         string
	Name             string
	LastName         string
	DateOfBirth      string
	Address          string
	DateOfIssue      string
	OfficerName      string
	DateOfExpiry     string
	OtherInformation string
}

type TextAnnotation struct {
	Locale      string `json:"locale"`
	Description string `json:"description"`
}

type Response struct {
	TextAnnotations []TextAnnotation `json:"textAnnotations"`
}

type OCRResponse struct {
	Responses []Response `json:"responses"`
}

func ocrHandler(w http.ResponseWriter, r *http.Request) {
	var ocrReq OCRRequest
	err := json.NewDecoder(r.Body).Decode(&ocrReq)
	if err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	apiKey := os.Getenv("API_KEY")
	visionURL := "https://vision.googleapis.com/v1/images:annotate?key=" + apiKey

	visionReq := VisionRequest{
		Requests: []Request{
			{
				Image: Image{
					Source: ImageSource{
						ImageUri: ocrReq.ImageUri,
					},
				},
				Features: []Feature{
					{Type: "TEXT_DETECTION"},
				},
			},
		},
	}

	reqBody, err := json.Marshal(visionReq)
	if err != nil {
		http.Error(w, "Error creating request", http.StatusInternalServerError)
		return
	}

	resp, err := http.Post(visionURL, "application/json", bytes.NewBuffer(reqBody))
	if err != nil {
		http.Error(w, "Error making request to Vision API", http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		http.Error(w, "Error reading response", http.StatusInternalServerError)
		return
	}

	var ocr OCRResponse
	err = json.Unmarshal([]byte(body), &ocr)
	if err != nil {
		log.Fatalf("Error unmarshaling OCR response: %v", err)
	}

	if len(ocr.Responses) > 0 && len(ocr.Responses[0].TextAnnotations) > 0 {
		description := ocr.Responses[0].TextAnnotations[0].Description
		parsedData := parseOCRDescription(description)

		fmt.Printf("ID Card Number: %s\n", parsedData.IDCardNumber)
		fmt.Printf("Name: %s\n", parsedData.Name)
		fmt.Printf("Last Name: %s\n", parsedData.LastName)
		fmt.Printf("Date of Birth: %s\n", parsedData.DateOfBirth)
		fmt.Printf("Address: %s\n", parsedData.Address)
		fmt.Printf("Date of Issue: %s\n", parsedData.DateOfIssue)
		fmt.Printf("Date of Expiry: %s\n", parsedData.DateOfExpiry)

		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"idCardNumber":"` + parsedData.IDCardNumber + `","name":"` + parsedData.Name + `","lastName":"` + parsedData.LastName + `","dateOfBirth":"` + parsedData.DateOfBirth + `","address":"` + parsedData.Address + `","dateOfIssue":"` + parsedData.DateOfIssue + `","dateOfExpiry":"` + parsedData.DateOfExpiry + `"}`))

	}
}

func main() {
	http.HandleFunc("/ocr", ocrHandler)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	fmt.Printf("Starting server on port %s\n", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
