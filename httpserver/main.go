package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"image/jpeg"
	"image/png"
	"io"
	"net"
	"net/http"
	"os"
	"time"

	"github.com/disintegration/imaging"
)

type User struct {
	User_Id       int
	Name          string
	Date_Of_Birth string
	Created_On    int64
}

type UserInfo struct {
	User_Id           int
	Name              string
	Birth_Day_Of_Week int
	Rfc_Created_On    string
}

var badRequestString string = fmt.Sprint(http.StatusBadRequest) + " " + http.StatusText(http.StatusBadRequest) + "\n"

func postJSON(responseWriter http.ResponseWriter, request *http.Request) {
	fmt.Println("Received request: /json")

	body, err := io.ReadAll(request.Body)

	if err != nil {
		fmt.Printf("Request body could not be read: %s\n", err)

		responseWriter.Header().Set("Content-Type", "text/plain")
		responseWriter.WriteHeader(http.StatusBadRequest)
		responseWriter.Write([]byte(badRequestString + err.Error()))

		return
	}

	var users []User
	var usersInfo []UserInfo

	jsonUnmarshallErr := json.Unmarshal([]byte(body), &users)

	if jsonUnmarshallErr != nil {
		fmt.Println("Error unmarshalling JSON string: ", jsonUnmarshallErr)

		responseWriter.WriteHeader(http.StatusBadRequest)
		responseWriter.Header().Set("Content-Type", "text/plain")
		responseWriter.Write([]byte(badRequestString + jsonUnmarshallErr.Error()))

		return
	}

	for _, user := range users {
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

		dateOfBirth, timeParseErr := time.Parse("2006-01-02", user.Date_Of_Birth)

		if timeParseErr != nil {
			fmt.Println("Error parsing date: ", timeParseErr)

			responseWriter.WriteHeader(http.StatusBadRequest)
			responseWriter.Header().Set("Content-Type", "text/plain")
			responseWriter.Write([]byte(badRequestString + timeParseErr.Error()))

			return
		}

		createdOn := time.Unix(user.Created_On, 0)
		location, loadLocationErr := time.LoadLocation("EST")

		if loadLocationErr != nil {
			fmt.Println("Error getting EST location: ", loadLocationErr)

			responseWriter.WriteHeader(http.StatusBadRequest)
			responseWriter.Header().Set("Content-Type", "text/plain")
			responseWriter.Write([]byte(badRequestString + loadLocationErr.Error()))

			return
		}

		createdOn = createdOn.In(location)

		userInfo := UserInfo{
			User_Id:           user.User_Id,
			Name:              user.Name,
			Birth_Day_Of_Week: dateOfBirth.Day(),
			Rfc_Created_On:    createdOn.Format("2006-01-02T15:04:05Z07:00"),
		}

		usersInfo = append(usersInfo, userInfo)
	}

	userInfoData, jsonMarshalErr := json.Marshal(usersInfo)

	if jsonMarshalErr != nil {
		fmt.Println("Error marshalling JSON struct data: ", jsonMarshalErr)

		responseWriter.WriteHeader(http.StatusBadRequest)
		responseWriter.Header().Set("Content-Type", "text/plain")
		responseWriter.Write([]byte(badRequestString + jsonMarshalErr.Error()))

		return
	}

	responseWriter.Header().Set("Content-Type", "application/json")
	responseWriter.Write(userInfoData)
}

func postJpegToPng(responseWriter http.ResponseWriter, request *http.Request) {
	fmt.Println("Received request: /jpeg-to-png")

	body, err := io.ReadAll(request.Body)

	if err != nil {
		fmt.Printf("Request body could not be read: %s\n", err)

		os.Exit(6)
	}

	jpeg, jpegDecodeErr := jpeg.Decode(bytes.NewReader([]byte(body)))

	if jpegDecodeErr != nil {
		fmt.Println("Error decoding JPEG: ", jpegDecodeErr)

		os.Exit(7)
	}

	bounds := jpeg.Bounds()
	width := float64(bounds.Dx())
	height := float64(bounds.Dy())
	var scaledWidth float64 = 256
	var scaledHeight float64 = 256

	if height > width {
		scaledWidth = scaledWidth * (width / height)
	} else if width > height {
		scaledHeight = scaledWidth * (height / width)
	}

	jpeg = imaging.Resize(jpeg, int(scaledWidth), int(scaledHeight), imaging.Lanczos)

	pngBuffer := new(bytes.Buffer)

	pngEncodeErr := png.Encode(pngBuffer, jpeg)

	if pngEncodeErr != nil {
		fmt.Println("Error encoding PNG from JPEG file: ", pngEncodeErr)

		os.Exit(8)
	}

	responseWriter.Header().Set("Content-Type", "image/png")
	responseWriter.Write(pngBuffer.Bytes())
}

func main() {
	const keyServerAddr = "serverAddr"
	mux := http.NewServeMux()
	ctx := context.Background()
	server := &http.Server{
		Addr:    ":8080",
		Handler: mux,
		BaseContext: func(l net.Listener) context.Context {
			ctx = context.WithValue(ctx, keyServerAddr, l.Addr().String())
			return ctx
		},
	}

	mux.HandleFunc("/json", postJSON)
	mux.HandleFunc("/jpeg-to-png", postJpegToPng)

	fmt.Printf("Starting server...\n")

	err := server.ListenAndServe()

	if errors.Is(err, http.ErrServerClosed) {
		fmt.Println("Server closed")
	} else if err != nil {
		fmt.Printf("Error starting server: %s\n", err)

		os.Exit(1)
	}
}
