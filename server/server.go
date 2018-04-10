package main

import (
 "fmt"
 "net"
 "strings"
 "errors"
 "strconv"
 "io"
)


type feature struct {
  f func (net.Conn, []string)error
  cnt_args int
  need_token bool
}

var ServerFunctions map[string]feature
var tokens map[string]string
var p = fmt.Println
////////////////////////////////////////////

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
  //TODO mutex or atomic or DB
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
    //token := args[0]
    login := args[1]

    //TODO mutex or atomic or DB
    delete(tokens, login)

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
