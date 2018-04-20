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
  DELEMITER = "│"
  MSG_DELIMITER = "|"
  SERVER_HEADER_SIZE = 5
  CLIENT_HEADER_SIZE = 4
  DB_NAME = "msngr.db"
  RSA_KEY_LEN = 1024
  MAX_LOGIN_LEN = 32
  MIN_LOGIN_LEN = 4
  TOKEN_TTL_SECONDS = 300
  TIME_FORMAT = "2006-01-02 15:04:05 -0700 MST"
  CHAN_NEW_MESSAGE_SIZE = 4
//-------------------------------------------
  OK_ANSWER = "ok"
  ERROR_BEGIN = "ERROR: "
// ERROR CODES
  OK_CODE = 0                                 // cmd sucsessfuly finished
  SERVER_INNER_ERR = 1                        // not depends of client
  UNDEINED_CMD = 2                            // undefined server command
  WRONG_ARGS_CNT = 3                          //
  WRONG_TOKEN = 4
  WRONG_DATA = 5
  WRONG_PUBLIC_KEY_MODULE = 6                 // error tranlate public key module to bigInt
  WRONG_PUBLIC_KEY_SIZE = 7
  WRONG_KEY_EXPONENT = 8                      // error tranlate public key exponent to int
  LOGIN_IS_USED = 9
  UNDEFINED_USER = 10
  WRONG_RANDOM_TEXT = 11                      // user can`t decrypt text correctly. sign in error
  GETTER_NOT_REGISTERED = 12                  // no getter with required login

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
