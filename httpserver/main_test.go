package main

import (
	"bytes"
	"errors"
	"image/png"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

type bodyReadError int

func (bodyReadError) Read(p []byte) (n int, err error) {
	return 0, errors.New("Body read error")
}

func responseCodeTest(test *testing.T, responseCode int, expectedResponseCode int) {
	if responseCode != expectedResponseCode {
		test.Error("Incorrect HTTP response code: expected " + strconv.Itoa(expectedResponseCode) + ", received " + strconv.Itoa(responseCode))
	}
}

func responseStringTest(test *testing.T, responseString string, expectedString string) {
	if responseString != expectedString {
		test.Error("Incorrect HTTP response string: expected " + expectedString + ", received " + responseString)
	}
}

func postJsonTest(test *testing.T, jsonFile string, responseCode int, responseString string) {
	openedJsonFile, openErr := os.Open(jsonFile)

	if openErr != nil {
		test.Fatal("Error opening JSON file: " + openErr.Error())
	}

	responseRecorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/json", openedJsonFile)

	postJson(responseRecorder, request)

	responseCodeTest(test, responseRecorder.Code, responseCode)
	responseStringTest(test, responseRecorder.Body.String(), responseString)
}

func postJpegToPngTest(test *testing.T, jpegFile string, responseCode int, responseString string) {
	openedJpegFile, openErr := os.ReadFile(jpegFile)

	if openErr != nil {
		test.Fatal("Error reading JPEG file: " + openErr.Error())
	}

	responseRecorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/jpeg-to-png", bytes.NewReader(openedJpegFile))

	request.Header.Set("Content-type", "application/x-www-form-urlencoded")

	postJpegToPng(responseRecorder, request)
	responseCodeTest(test, responseRecorder.Code, responseCode)

	if responseCode == http.StatusOK {
		pngPath := strings.TrimSuffix(jpegFile, filepath.Ext(jpegFile)) + ".png"
		pngImage, decodeErr := png.Decode(bytes.NewReader(responseRecorder.Body.Bytes()))
		pngFile, createErr := os.Create(pngPath)
		encodeErr := png.Encode(pngFile, pngImage)

		if decodeErr != nil {
			test.Fatal("Error decoding PNG image: " + decodeErr.Error())
		}

		if createErr != nil {
			test.Fatal("Error creating PNG image: " + decodeErr.Error())
		}

		if encodeErr != nil {
			test.Fatal("Error encoding PNG image: " + decodeErr.Error())
		}

		_, statErr := os.Stat(pngPath)

		if errors.Is(statErr, os.ErrNotExist) {
			test.Error("Expected PNG file " + pngPath + " does not exist")
		} else if statErr != nil {
			test.Fatal("Error getting FileInfo: " + statErr.Error())
		}
	} else {
		responseStringTest(test, responseRecorder.Body.String(), responseString)
	}
}

func TestPostJson(test *testing.T) {
	postJsonTest(test, "./files/json/users.json", http.StatusOK, "[{\"User_Id\":1,\"Name\":\"Joe Smith\",\"Birth_Day_Of_Week\":12,\"Rfc_Created_On\":\"2022-01-19T12:07:14-05:00\"},{\"User_Id\":2,\"Name\":\"Jane Doe\",\"Birth_Day_Of_Week\":6,\"Rfc_Created_On\":\"2022-01-19T12:07:14-05:00\"}]")
}

func TestPostJsonIncorrectMethod(test *testing.T) {
	responseRecorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/json", nil)

	postJson(responseRecorder, request)
	responseCodeTest(test, responseRecorder.Code, http.StatusBadRequest)
	responseStringTest(test, responseRecorder.Body.String(), "400 Bad Request\nExpected POST request method, instead received GET")
}

func TestPostJsonRequestBodyReadError(test *testing.T) {
	responseRecorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/json", bodyReadError(0))

	postJson(responseRecorder, request)
	responseCodeTest(test, responseRecorder.Code, http.StatusBadRequest)
	responseStringTest(test, responseRecorder.Body.String(), "400 Bad Request\nBody read error")
}

func TestPostJsonIncorrectFileType(test *testing.T) {
	postJsonTest(test, "./files/images/cat.jpg", http.StatusBadRequest, "400 Bad Request\ninvalid character 'ÿ' looking for beginning of value")
}

func TestPostJsonUnmarshalUnexpectedEndOfJsonInput(test *testing.T) {
	postJsonTest(test, "./files/json/users_json_unmarshal_unexpected_end_of_json_input.json", http.StatusBadRequest, "400 Bad Request\nunexpected end of JSON input")
}

func TestPostJsonUnmarshalInvalidValueType(test *testing.T) {
	postJsonTest(test, "./files/json/users_json_unmarshal_invalid_value_type.json", http.StatusBadRequest, "400 Bad Request\njson: cannot unmarshal string into Go struct field User.User_Id of type int")
}

func TestPostJsonUserIdLessThanOrEqualToZero(test *testing.T) {
	expectedString := "400 Bad Request\nExpected JSON key user_id not set or set to invalid value; user_id must be set to a value greater than 0"

	postJsonTest(test, "./files/json/users_user_id_not_set.json", http.StatusBadRequest, expectedString)
	postJsonTest(test, "./files/json/users_user_id_invalid_value.json", http.StatusBadRequest, expectedString)
}

func TestPostJsonNameEmptyString(test *testing.T) {
	expectedString := "400 Bad Request\nExpected JSON key name not set or set to invalid value; name must be set to a valid string"

	postJsonTest(test, "./files/json/users_name_not_set.json", http.StatusBadRequest, expectedString)
	postJsonTest(test, "./files/json/users_name_invalid_value.json", http.StatusBadRequest, expectedString)
}

func TestPostJsonDateOfBirthTimeParseError(test *testing.T) {
	expectedString := "400 Bad Request\nparsing time \"\" as \"2006-01-02\": cannot parse \"\" as \"2006\""

	postJsonTest(test, "./files/json/users_date_of_birth_not_set.json", http.StatusBadRequest, expectedString)
	postJsonTest(test, "./files/json/users_date_of_birth_invalid_value.json", http.StatusBadRequest, expectedString)
}

func TestPostJsonCreatedOnErrorCases(test *testing.T) {
	expectedString := "400 Bad Request\nExpected JSON key created_on not set or set to invalid value; created_on must be set to a valid greater than 0"

	postJsonTest(test, "./files/json/users_created_on_not_set.json", http.StatusBadRequest, expectedString)
	postJsonTest(test, "./files/json/users_created_on_invalid_value.json", http.StatusBadRequest, expectedString)
}

func TestPostJpegToPng(test *testing.T) {
	postJpegToPngTest(test, "./files/images/cat.jpg", http.StatusOK, "")
	postJpegToPngTest(test, "./files/images/jpg-vs-jpeg.jpg", http.StatusOK, "")
	postJpegToPngTest(test, "./files/images/twitter-camera.jpeg", http.StatusOK, "")
}

func TestPostJpegToPngIncorrectMethod(test *testing.T) {
	responseRecorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/jpeg-to-png", nil)

	postJpegToPng(responseRecorder, request)
	responseCodeTest(test, responseRecorder.Code, http.StatusBadRequest)
	responseStringTest(test, responseRecorder.Body.String(), "400 Bad Request\nExpected POST request method, instead received GET")
}

func TestPostJpegToPngRequestBodyReadError(test *testing.T) {
	responseRecorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/jpeg-to-png", bodyReadError(0))

	postJpegToPng(responseRecorder, request)
	responseCodeTest(test, responseRecorder.Code, http.StatusBadRequest)
	responseStringTest(test, responseRecorder.Body.String(), "400 Bad Request\nBody read error")
}

func TestPostJpegToPngIncorrectFileType(test *testing.T) {
	postJpegToPngTest(test, "./files/json/users.json", http.StatusBadRequest, "400 Bad Request\ninvalid JPEG format: missing SOI marker")
	postJpegToPngTest(test, "./files/images/jpg-vs-jpeg.png", http.StatusBadRequest, "400 Bad Request\ninvalid JPEG format: missing SOI marker")
}
