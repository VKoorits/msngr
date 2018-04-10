package main

import (
 "fmt"
 "net"
 "os"
 "strings"
 "math/rand"
 "errors"
 "strconv"
 "encoding/binary"
 "bytes"
 "io"
)

const (
 CONN_HOST = "localhost"
 CONN_PORT = "5001"
 CONN_TYPE = "tcp"
 //---------------------
 RANDOM_TEXT_SIZE = 16
 TOKEN_SIZE = 8
 DELEMITER = ":"
 OK_ANSWER = "ok"
 OK_CODE = 0
 ERROR_CODE = 42
 SERVER_HEADER_SIZE = 5
 CLIENT_HEADER_SIZE = 4
 ERROR_BEGIN = "?ERROR: "
)

type feature struct {
  f func (net.Conn, []string)error
  cnt_args int
  need_token bool
}

var ServerFunctions map[string]feature
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
    //go echo(conn)
  }
}

func echo(conn net.Conn){
  header := []byte("hello")
  data, _ := recvDataB(conn)
  fmt.Println(data)
  conn.Write(header)
}

func initServerFunctions() {
  ServerFunctions = make( map[string]feature )
  tokens  = make(map[string]string)
  ServerFunctions["GET_TKN"] = feature{getToken, 1, false}
  ServerFunctions["QUIT"] = feature{unlogin, 2, true}
  ServerFunctions["SEND_MSG"] = feature{sendMsg, 4, true}
  ServerFunctions["GET_MSG"] = feature{getNewMsg, 2, true}
  ServerFunctions["TEST"] = feature{test, 2, true}
}

func sendError(conn net.Conn, errText string) {
  fmt.Println(ERROR_BEGIN + errText)
  sendData(conn, ERROR_BEGIN + errText, ERROR_CODE )
}

func workWithClient(conn net.Conn) {
  defer conn.Close()
  fmt.Println("connected")
  for {
    data, err := recvDataB(conn)
    if err != nil {
      if err == io.EOF {
        break
      }
      sendError(conn, err.Error())
      continue
    }

    req := strings.Split(string(data[:len(data)]), DELEMITER)
    cmd := req[0]
    args := req[1:]

    ficha, ok := ServerFunctions[cmd]
    if !ok {
      sendError(conn, "wrong request cmd: '" + cmd + "'")
      continue
    }
    if len(args) != ficha.cnt_args {
      sendError(conn, "wrong count arguments in " + cmd + ". Expected " +
                  strconv.Itoa(ficha.cnt_args) +  ", got " + strconv.Itoa(len(args)) )
      continue
    }
    // Run comand
    if ficha.need_token {
      okToken := checkToken(args[0], args[1])
      if !okToken {
        sendError(conn, "wrong token")
        continue
      }
    }
    err = ficha.f(conn, args)

    if err != nil {
      if err == io.EOF {
        break
      }
      sendError(conn, err.Error())
    } else {
      sendOkStatus(conn)
    }
  }
}

/////////////////////////////////////////

func checkToken(login string, token string) bool {
  return tokens[login] == token
}

func getToken(conn net.Conn, args []string) error {
  inText := getRandomText(RANDOM_TEXT_SIZE)
  inData := []byte(inText)
  // ENCRYPT inData || inText
  sendDataB(conn, inData, uint32(len(inData)), OK_CODE)

  ans, err := recvDataB(conn)

  if err != nil {
    return err
    //return errors.New("reading in getToken")
  }

  if inText != string(ans) {
      return errors.New("wrong decrypted random_text")
  }

  bToken := []byte( getRandomText(TOKEN_SIZE) )
  sendDataB(conn, bToken, uint32(len(bToken)), OK_CODE)


  tokens[args[0]] = string(bToken)

  return nil
}

func unlogin(conn net.Conn, args []string) error {
    token := args[0]
    login := args[1]

    if tokens[login] == token {
      delete(tokens, login)
    }
    return nil
}

func test(conn net.Conn, args []string) error {
  return nil
}

func sendMsg(conn net.Conn, args []string) error {
  return nil
}

func getNewMsg(conn net.Conn, args []string) error {
  return nil
}

///////////////////////////////////////

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

func sendOkStatus(conn net.Conn) {
  sendData(conn, OK_ANSWER, OK_CODE)
}

func sendDataB(conn net.Conn, data []byte, dataSize uint32, code uint8) {
  header := make([]byte, SERVER_HEADER_SIZE)
  binary.LittleEndian.PutUint32(header, dataSize)
  header[4] = code
  res := append(header, data...)
  conn.Write(res)
}

func sendData(conn net.Conn, textData string, code uint8) {
  dataSize := uint32(len(textData))
  sendDataB(conn, []byte(textData), dataSize, code)
}

func recvDataB(conn net.Conn) ([]byte, error) {
  header := make( []byte, CLIENT_HEADER_SIZE)
  data_len, err := conn.Read(header)

  if err != nil {
    return nil, err
  }

  var dataSize uint32
  binary.Read(bytes.NewReader(header[0:4]), binary.LittleEndian, &dataSize)

  data := make([]byte, dataSize)
  data_len, err = conn.Read(data)

  if err != nil {
    return nil, err
  }
  if data_len != int(dataSize) {
    return nil, errors.New("real and expected data size not equal")
  }

  return data, nil
}













//
