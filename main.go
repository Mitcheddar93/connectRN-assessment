package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"image/jpeg"
	"image/png"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/disintegration/imaging"
)

// Data that is expected to be marshalled from the JSON input file
type User struct {
	User_Id       int
	Name          string
	Date_Of_Birth string
	Created_On    int64
}

// Data that will be marshalled for the JSON output
type UserInfo struct {
	User_Id           int
	Name              string
	Birth_Day_Of_Week int
	Rfc_Created_On    string
}

// 400 Bad Request - used in many error handling messages, put in variable for all methods to access
var badRequestString string = fmt.Sprint(http.StatusBadRequest) + " " + http.StatusText(http.StatusBadRequest) + "\n"

// Method to handle JSON post requests, called using http://localhost:8080/json
// Tested with CURL using the following cURL command: curl -X POST -d "@./path/to/json/file.json" http://localhost:8080/json
func postJson(responseWriter http.ResponseWriter, request *http.Request) {
	fmt.Println("Received request: /json")

	// Check to ensure the POST method is called to upload and output JSON
	// When an error occurs, error will output to client detailing what went wrong before continuing
	if request.Method != "POST" {
		requestMethodErr := "Expected POST request method, instead received " + request.Method

		fmt.Println("Error reading JSON: ", requestMethodErr)

		responseWriter.WriteHeader(http.StatusBadRequest)
		responseWriter.Header().Set("Content-Type", "text/plain")
		responseWriter.Write([]byte(badRequestString + requestMethodErr))

		return
	}

	// Read the body of the request and output an io.Reader value
	body, err := io.ReadAll(request.Body)

	// Output an error to the client if the body could not be read for any reason
	if err != nil {
		fmt.Printf("Request body could not be read: %s\n", err)

		responseWriter.Header().Set("Content-Type", "text/plain")
		responseWriter.WriteHeader(http.StatusBadRequest)
		responseWriter.Write([]byte(badRequestString + err.Error()))

		return
	}

	var users []User
	var usersInfo []UserInfo

	// Unmarshal the data of the JSON file and output to a User struct variable
	jsonUnmarshalErr := json.Unmarshal([]byte(body), &users)

	// If the value in a JSON field is of the wrong type, this error will be shown
	if jsonUnmarshalErr != nil {
		fmt.Println("Error unmarshalling JSON string: ", jsonUnmarshalErr)

		responseWriter.WriteHeader(http.StatusBadRequest)
		responseWriter.Header().Set("Content-Type", "text/plain")
		responseWriter.Write([]byte(badRequestString + jsonUnmarshalErr.Error()))

		return
	}

	// Loop through all of the users in the JSON file
	for _, user := range users {
		// Check that each of the JSON fields is set and has a valid value; if not, show the client an error
		if user.User_Id <= 0 {
			userIdErr := "Expected JSON key user_id not set or set to invalid value; user_id must be set to a value greater than 0"

			fmt.Println("Error reading JSON: ", userIdErr)

			responseWriter.WriteHeader(http.StatusBadRequest)
			responseWriter.Header().Set("Content-Type", "text/plain")
			responseWriter.Write([]byte(badRequestString + userIdErr))

			return
		}

		if user.Name == "" {
			userNameErr := "Expected JSON key name not set or set to invalid value; name must be set to a valid string"

			fmt.Println("Error reading JSON: ", userNameErr)

			responseWriter.WriteHeader(http.StatusBadRequest)
			responseWriter.Header().Set("Content-Type", "text/plain")
			responseWriter.Write([]byte(badRequestString + userNameErr))

			return
		}

		// To parse a date in Go, layouts must use the reference time Mon Jan 2 15:04:05 MST 2006; only the year, month and day are expected from the JSON output
		dateOfBirth, timeParseErr := time.Parse("2006-01-02", user.Date_Of_Birth)

		if timeParseErr != nil {
			fmt.Println("Error parsing date: ", timeParseErr)

			responseWriter.WriteHeader(http.StatusBadRequest)
			responseWriter.Header().Set("Content-Type", "text/plain")
			responseWriter.Write([]byte(badRequestString + timeParseErr.Error()))

			return
		}

		createdOn := time.Unix(user.Created_On, 0)

		// The date returned from the Unix time will be set to the earliest possible value if the JSON field is not set
		// The value of the Unix time must also be zero or greater to be valid
		if createdOn.Before(time.Unix(0, 0)) || createdOn.Equal(time.Unix(0, 0)) {
			createdOnErr := "Expected JSON key created_on not set or set to invalid value; created_on must be set to a valid greater than 0"

			fmt.Println("Error reading JSON: ", createdOnErr)

			responseWriter.WriteHeader(http.StatusBadRequest)
			responseWriter.Header().Set("Content-Type", "text/plain")
			responseWriter.Write([]byte(badRequestString + createdOnErr))

			return
		}

		// Set the RFC time output to show Eastern Standard Time (time will be set to Eastern Daylight Time depending on region and current date)
		location, loadLocationErr := time.LoadLocation("EST")

		if loadLocationErr != nil {
			fmt.Println("Error getting EST location: ", loadLocationErr)

			responseWriter.WriteHeader(http.StatusBadRequest)
			responseWriter.Header().Set("Content-Type", "text/plain")
			responseWriter.Write([]byte(badRequestString + loadLocationErr.Error()))

			return
		}

		createdOn = createdOn.In(location)

		// Prepare the UserInfo struct variable to output to the body
		userInfo := UserInfo{
			User_Id:           user.User_Id,
			Name:              user.Name,
			Birth_Day_Of_Week: dateOfBirth.Day(),
			Rfc_Created_On:    createdOn.Format("2006-01-02T15:04:05Z07:00"),
		}

		// Append each user into a UserInfo array
		usersInfo = append(usersInfo, userInfo)
	}

	// Marshal the UserInfo array into a byte array for processing
	userInfoData, jsonMarshalErr := json.Marshal(usersInfo)

	if jsonMarshalErr != nil {
		fmt.Println("Error marshalling JSON struct data: ", jsonMarshalErr)

		responseWriter.WriteHeader(http.StatusBadRequest)
		responseWriter.Header().Set("Content-Type", "text/plain")
		responseWriter.Write([]byte(badRequestString + jsonMarshalErr.Error()))

		return
	}

	// Setting the Content-type to JSON will output the byte array as a JSON string that will be outputted in the client body
	responseWriter.Header().Set("Content-Type", "application/json")
	responseWriter.Write(userInfoData)
}

// Method to handle post requests to run JPEG files in to PNG files, while scaling to a maximum of 256x256, called using http://localhost:8080/jpeg-to-png
// Tested with CURL using the following cURL command: curl -X POST --data-binary "@./path/to/jpeg/file.jpeg(or .jpg)" http://localhost:8080/json > path\to\png\file.png
// The --data-binary argument is necessary as using -d (--data) removes newline characters that jpeg.Decode needs to properly read the JPEG file
func postJpegToPng(responseWriter http.ResponseWriter, request *http.Request) {
	fmt.Println("Received request: /jpeg-to-png")

	// Check to ensure the POST method is called to upload and output JSON
	// When an error occurs, error will output to client detailing what went wrong before continuing
	if request.Method != "POST" {
		requestMethodErr := "Expected POST request method, instead received " + request.Method

		fmt.Println("Error reading JSON: ", requestMethodErr)

		responseWriter.WriteHeader(http.StatusBadRequest)
		responseWriter.Header().Set("Content-Type", "text/plain")
		responseWriter.Write([]byte(badRequestString + requestMethodErr))

		return
	}

	// Read the body of the request and output an io.Reader value
	body, err := io.ReadAll(request.Body)

	// Output an error to the client if the body could not be read for any reason
	if err != nil {
		fmt.Printf("Request body could not be read: %s\n", err)

		responseWriter.Header().Set("Content-Type", "text/plain")
		responseWriter.WriteHeader(http.StatusBadRequest)
		responseWriter.Write([]byte(badRequestString + err.Error()))

		return
	}

	// Decode the body in a bytes.Reader format to an image.Image variable to manipulate and convert
	jpeg, jpegDecodeErr := jpeg.Decode(bytes.NewReader([]byte(body)))

	if jpegDecodeErr != nil {
		fmt.Println("Error decoding JPEG: ", jpegDecodeErr)

		responseWriter.WriteHeader(http.StatusBadRequest)
		responseWriter.Header().Set("Content-Type", "text/plain")
		responseWriter.Write([]byte(badRequestString + jpegDecodeErr.Error()))

		return
	}

	// Determine the size of the JPEG provided and set the target width and height to 256px
	bounds := jpeg.Bounds()
	width := float64(bounds.Dx())
	height := float64(bounds.Dy())
	var scaledWidth float64 = 256
	var scaledHeight float64 = 256

	// Determine the target width and height by determining whether the height or width is larger and changing the size accordingly
	// If the height is greater than the width, set the target width to the scale of the width:height ratio
	// If the width is greater than the height, set the target height to the scale of the height:width ratio
	// If both are equal, keep the width and height at 256px
	if height > width {
		scaledWidth = scaledWidth * (width / height)
	} else if width > height {
		scaledHeight = scaledHeight * (height / width)
	}

	// Resize the image using the github.com/disintegration/imaging library
	// Attempts were made to use built-in Go commands to resize the image, but time and complexity led to using an outside library
	jpeg = imaging.Resize(jpeg, int(scaledWidth), int(scaledHeight), imaging.Lanczos)

	// Create a PNG buffer to store the PNG data in
	pngBuffer := new(bytes.Buffer)

	// Encode the data from the JPEG file into the PNG buffer, throwing an error if one occurs
	pngEncodeErr := png.Encode(pngBuffer, jpeg)

	if pngEncodeErr != nil {
		fmt.Println("Error encoding PNG from JPEG file: ", pngEncodeErr)

		responseWriter.WriteHeader(http.StatusBadRequest)
		responseWriter.Header().Set("Content-Type", "text/plain")
		responseWriter.Write([]byte(badRequestString + pngEncodeErr.Error()))

		return
	}

	// Output the Content-type as a PNG image
	// The file output is handled by the client
	responseWriter.Header().Set("Content-Type", "image/png")
	responseWriter.Write(pngBuffer.Bytes())
}

// Initiates the HTTP server using the base ListenAndServe method
// Server is called using http://localhost:8080
func main() {
	http.HandleFunc("/json", postJson)
	http.HandleFunc("/jpeg-to-png", postJpegToPng)

	fmt.Printf("Starting server...\n")

	err := http.ListenAndServe(":8080", nil)

	// Output a message depending on if the server is closed legitimately or not
	if errors.Is(err, http.ErrServerClosed) {
		fmt.Println("Server closed")
	} else if err != nil {
		fmt.Printf("Error starting server: %s\n", err)

		os.Exit(1)
	}
}
