package main


import (
	"bufio"
	"fmt"
	"net"
	"os"
	"io"
	"strings"
	"errors"
)

const (
	EXIT = "exit"
	MAXLEN = 1024
)

func main(){

	// need to put the port number as a cmd-line arg
	args := os.Args

	if len(args) <=1 {
		fmt.Println("Port Number Not Given")
		os.Exit(0)
	}

	// making the port argument from the cmd line argument given by the user
	l, err := net.Listen("tcp", ":"+args[1])

	if err != nil{
		fmt.Println("Error while creating the TCP server: ", err.Error())
		os.Exit(1)
	}

	// close the server at the end of main
	defer l.Close()

	fmt.Println("Listening to ", ":"+args[1])

	for {

		conn, err := l.Accept()
		if err != nil{
			fmt.Println("Error in accepting connection reuqest", err.Error())
		}

		// added go routines to support spawnning of multiple tcp servers together conncurrently
		go handleConnection(conn)
	}
}



func readFromConn(conn net.Conn) (string, error){

	// to limit the length of data being read
	maxLen := &io.LimitedReader{conn, MAXLEN }
	reader := bufio.NewScanner(maxLen)
	nstr := ""
	for reader.Scan(){
		nstr += reader.Text()
	}

	if nstr == EXIT{
		return "", nil
	}
	// to create an output string that shows how much of the campaign was succesful
	// Example Read String
	// "{Impressions:1000, CTR:1200, Budget:1200}"
	result := make(map[string]string)
	if strings.Contains(nstr, "Impressions") && strings.Contains(nstr, "CTR") && strings.Contains(nstr, "Budget"){
		sep := strings.Split(nstr,",")
		firstVal := strings.Split(sep[0],":")
		result["Impressions"] = firstVal[len(firstVal)-1]
		secondVal := strings.Split(sep[1], ":")
		result["CTR"] = secondVal[len(secondVal)-1]
		thirdVal := strings.Split(sep[2], ":")
		anotherVal := thirdVal[1]
		result["Budget"] = anotherVal[:len(anotherVal)-1]

	} else {
		return "", nil
	}
	return fmt.Sprintf("Achieved: Impressions %s per fetch with a CTR of %s%% using 1/4th of given budget $%s", result["Impressions"], result["CTR"], result["Budget"]), nil
}
// returns number of bytes written to the file
func writeToConn(conn net.Conn, written string, response chan<- bool) (int, error){

	// writes to the file


	f, err := os.Create("data.txt")
	if err != nil{
		response<-false
		return 0, err
	}
	defer f.Close()

	writer := bufio.NewWriter(f)
	num, err := writer.WriteString(written)
	if err!=nil{
		response<-false
		return 0, err
	}
	fmt.Printf("Wrote %d bytes\n", num)

	writer.Flush()

	// writing to the conn
	if _, err:= conn.Write([]byte(written)); err!=nil{
		
		response<-false
		return 0, err
	}

	response<-true
	return num, nil

}

func handleConnection(conn net.Conn){
	// reading from the tcp connection

	// added channel to ensure synchronization of internal operations of the concurrent go-routines which
	// are being spawned
	responseWritten := make(chan bool)

	buff, err := readFromConn(conn)

	if err != nil{
		fmt.Println(err.Error())
	}

	if buff == EXIT{
		return
	}
	// adding newline delim

	buff = buff+"\n"
	if _, err := writeToConn(conn, buff, responseWritten); err!=nil{
		panic(err)
	}

	completed := <- responseWritten

	if completed{
		conn.Close()
	} else{
		errors.New("Synchronization Issues")
	}

}