package main

import (
 "fmt"
 "net"
 "os"
 "strings"
 "math/rand"
 "errors"
)

const (
 CONN_HOST = "localhost"
 CONN_PORT = "5001"
 CONN_TYPE = "tcp"
 //---------------------
 RANDOM_TEXT_SIZE = 16
 TOKEN_SIZE = 8
)

var ServerFunctions map[string]func(net.Conn, string)error
var tokens map[string]string
var p = fmt.Println

func main() {
  initServerFunctions()
  sock, err := net.Listen(CONN_TYPE, CONN_HOST + ":" + CONN_PORT)
  if err != nil {
    fmt.Println("Error listening:", err.Error())
    os.Exit(1)
  }

  defer sock.Close()

  fmt.Println("Listening on " + CONN_HOST + ":" + CONN_PORT)

  for {
    conn, err := sock.Accept()
    if err != nil {
      fmt.Println("Error accepting: ", err.Error())
      continue
    }
    go workWithClient(conn)
  }
}

func initServerFunctions() {
  ServerFunctions = make( map[string]func(net.Conn, string)error )
  tokens  = make(map[string]string)
  ServerFunctions["GET_TKN"] = getToken
  ServerFunctions["QUIT"] = unlogin
}

func workWithClient(conn net.Conn) {
  defer conn.Close()
  fmt.Println("connected")
  for {
    data := make([]byte, 1024)
    data_len, err := conn.Read(data)

    if err != nil {
      fmt.Println("Error reading:", err.Error())
      break
    }
    if data_len == 0 {
      fmt.Println("No data")
      break
    }

    text := string(data[:data_len])
    cmd := strings.Split(text, ":")[0]
    args := text[len(cmd)+1:]
    err = ServerFunctions[cmd](conn, args)
    p(tokens)
    if err != nil {
      break
    }


  }
}


func getToken(conn net.Conn, args string) error {

  inText := getRandomText(RANDOM_TEXT_SIZE)
  inData := []byte(inText)
  // ENCRYPT inData || inText
  conn.Write(inData)

  ans := make([]byte, len(inData))
  _, err := conn.Read(ans)

  if err != nil {
    fmt.Printf("Error reading in getToken\n")
    return errors.New("Error reading in getToken\n")
  }

  if inText != string(ans) {
      fmt.Printf("Error: wrong decrypted random_text\n")
      return errors.New("Error: wrong decrypted random_text\n")
  }

  bToken := []byte( getRandomText(TOKEN_SIZE) )
  conn.Write(bToken)


  tokens[args] = string(bToken)

  return nil
}

func unlogin(conn net.Conn, args string) error {
    token := strings.Split(args, ";")[0]
    login := args[len(token)+1:]

    if tokens[login] == token {
      delete(tokens, login)
    }

    return nil
}

func getRandomText(text_len int) string {
  // without ';' and ':'
  chars := "QWERTYUIOPASDFGHJKLZXCVBNMqwertyuiopasdfghjklzzxcvbnm_1234567890!@#$%^&*()_=+,.<>/?[{}]"
  res := ""
  for i := 0; i < text_len; i += 1 {
    num := rand.Intn(len(chars))
    res += chars[num:num+1]
  }
  return res
}














//
