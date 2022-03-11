package main

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"os"
	"strings"
)

/*
	Connect to TCP on port 8080 and start handling input
*/
func main() {
	//Connect TCP on port 8080
	ln, err := net.Listen("tcp", ":8080")
	handleError(err)
	fmt.Println("Started Server.")

	for { //ever
		//accept connection
		conn, err := ln.Accept()
		handleError(err)
		//parse request and send response
		handleConnection(conn)
		conn.Close()
	}
}

/*
	Parse request message, form response message from request, send response and data if needed
	Parameters: accepted connection
*/
func handleConnection(conn net.Conn) {
	//parse response, code is GET/HEAD, uri is requested file
	code, uri := parseRequest(conn)
	//response is the headers of the response message, data is the content of the file, byte array; data is nil if it was a HEAD request
	response, data := generateResponse(code, uri)
	//if it is a head request, just send the response headers
	if data != nil {
		sendResponse(response, data, conn)
	} else {
		fmt.Fprint(conn, response)
	}
}

/*
	print out headers of request message and return the HTTP request code and the uri requested
	Parameters: accepted connection
	Output: HTTP request code, uri requested
*/
func parseRequest(conn net.Conn) (string, string) {
	//read on connection
	reader := bufio.NewReader(conn)
	//read up to \n and trim of CRLF at the end
	msg, err := reader.ReadString('\n')
	handleError(err)
	msg = strings.TrimSuffix(msg, "\r\n")
	fmt.Println(msg)
	//print out all headers after first line
	for reader.Buffered() > 0 {
		header, err := reader.ReadString('\n')
		handleError(err)
		header = strings.TrimSuffix(header, "\r\n")
		fmt.Println(header)
	}

	//parse first line by separating by space to get the request code and uri. EX: ["GET", "/index.html", "HTTP/1.1"]
	msgSplit := strings.Split(msg, " ")
	//if for some reason there is no uri or no request code, exit
	if len(msgSplit) < 2 {
		fmt.Println("ERROR: TOO FEW HTML ARGUMENTS")
		os.Exit(1)
		//if it is not a head or a get request then we cannot process it
	} else if msgSplit[0] != "HEAD" && msgSplit[0] != "GET" {
		fmt.Println("ERROR: NOT GET OR HEAD REQUEST")
		os.Exit(1)
	}
	//return request code, uri
	return msgSplit[0], msgSplit[1]
}

/*
	Creates a response message for a given HTTP request
	Parameters: HTTP request code (GET/HEAD), uri of file needed
	Output: Response message as string, data of requested file as byte array or nil if it was a HEAD request
*/
func generateResponse(code string, uri string) (string, []byte) {
	responseMessage := "HTTP/1.1"
	fileName := strings.TrimPrefix(uri, "/")
	//try to open contents of file
	data, err := ioutil.ReadFile("www/" + fileName)

	//if file does not exist, send a 404 not found response
	if err != nil {
		responseMessage += " 404 Not found\r\n"
		//get data from 404.html
		data, err = ioutil.ReadFile("www/404.html")
		handleError(err)
	} else { //otherwise generate a 200 response and add content length, last modified, and server headers
		responseMessage += " 200 Okay\r\n"

		//get content length
		contentLength := len(data)
		responseMessage += "Content-length: " + fmt.Sprintf("%d", contentLength) + "\r\n"

		//get last modified date of file and convert to GMT
		file, _ := os.Stat("www/" + uri)
		modifiedTime := file.ModTime()

		dayName := modifiedTime.Weekday().String()
		day := modifiedTime.Day()
		month := modifiedTime.Month().String()
		year := modifiedTime.Year()
		hour := modifiedTime.Hour()
		minutes := modifiedTime.Minute()
		seconds := modifiedTime.Second()
		responseMessage += fmt.Sprintf("Last-Modified: %s, %02d %s %d %02d:%02d:%02d GMT\r\n", dayName, day, month, year, hour, minutes, seconds)

		//add on server header
		responseMessage += "Server: cihttp\r\n\r\n"
	}

	//if the request was a HEAD request, give nil as the data
	if code == "GET" {
		return responseMessage, data
	} else {
		return responseMessage, nil
	}
}

/*
	Sends a response message on the given connection
	Parameters: response headers in string format, data in byte array, accepted connection
*/
func sendResponse(response string, data []byte, conn net.Conn) {
	fmt.Fprint(conn, response)
	conn.Write(data)
}

/*
	Simple error handling helper method, if an error occurs, print out the error message and quit
*/
func handleError(err error) {
	if err != nil {
		fmt.Println("ERROR")
		log.Fatal(err.Error())
		os.Exit(1)
	}
}
