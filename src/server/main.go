package main

import (
 "fmt"
 "net"
)

const (
 CONN_HOST = "localhost"
 CONN_PORT = "5002"
 CONN_TYPE = "tcp"
 //---------------------
 RANDOM_TEXT_SIZE = 16
 TOKEN_SIZE = 8
 DELEMITER = ":"
 OK_ANSWER = "ok"
 OK_CODE = 0
 ERROR_CODE = 42
 ERROR_BEGIN = "?ERROR: "
 SERVER_HEADER_SIZE = 5
 CLIENT_HEADER_SIZE = 4
 DB_NAME = "msngr.db"
 RSA_KEY_LEN = 1024
 MAX_LOGIN_LEN = 32
 MIN_LOGIN_LEN = 4
)



func main() {
  initServerFunctions()
  sock, err := net.Listen(CONN_TYPE, CONN_HOST + ":" + CONN_PORT)
  if err != nil {
    fmt.Println("Error listening:", err.Error())
    return
  }
  db := InitDB(DB_NAME)

  defer sock.Close()
  defer db.Close()

  fmt.Println("Listening on " + CONN_HOST + ":" + CONN_PORT)

  for {
    conn, err := sock.Accept()
    if err != nil {
      fmt.Println("Error accepting: ", err.Error())
      continue
    }
    go workWithClient(conn, db)
    //go echo(conn)
  }
}
