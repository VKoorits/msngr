package main

import (
 "fmt"
 "net"
 "strings"
 "errors"
 "strconv"
 "io"
 "database/sql"
 "math/big"
)


// command whic server can execute
type feature struct {
  f func (net.Conn, []string, *sql.DB)serverError
  cnt_args int
  need_token bool
}

type message struct {
  sender string
  text string
  sendingTime string
}

type serverError struct {
  Err error
  Code uint8
}

func (msg message) ToStr() string {
  return msg.sender + DELEMITER +
         msg.text + DELEMITER +
         msg.sendingTime
}

var ServerFunctions map[string]feature
var NoErrors serverError
// alias only for debug
var p = fmt.Println
////////////////////////////////////////////

//echo server fo testing
func echo(conn net.Conn){
  defer conn.Close()
  bData, _ := recvDataB(conn)
  sendDataB(conn, bData, OK_CODE)
}

//init some global vars
func initServerFunctions() {
  NoErrors = serverError{nil, OK_CODE}
  // commands which server can  execute
  ServerFunctions = make( map[string]feature )
  ServerFunctions["SIGN_UP"] = feature{sign_up, 3, false}
  ServerFunctions["GET_TKN"] = feature{sign_in, 1, false}
  ServerFunctions["QUIT"] = feature{unlogin, 2, true}
  ServerFunctions["SEND_MSG"] = feature{sendMsg, 4, true}
  ServerFunctions["GET_MSG"] = feature{getNewMsg, 2, true}
  ServerFunctions["FIND_USR"] = feature{findUsernames, 3, true}
}

// function which work with connection in new gorutine
func workWithClient(conn net.Conn, db *sql.DB) {
  defer conn.Close()
  fmt.Println("connected")

  // read command from user
  for {
    data, err := recvDataB(conn)
    if err != nil {
      if err == io.EOF {
        break
      }
      // NOTE: interesting, what reading error besides EOF
      //?/sendError(conn, err.Error())
      sendError(conn, serverError{err, SERVER_INNER_ERR})
      continue
    }

    // get command and args for it
    req := strings.Split(string(data[:len(data)]), DELEMITER)
    cmd := req[0]
    args := req[1:]

    // ficha - is function which is called for this command
    ficha, ok := ServerFunctions[cmd]

    // check  params for call feature
    if !ok {
      sErr := serverError{errors.New("undefined cmd: '" + cmd + "'"), UNDEINED_CMD}
      sendError(conn, sErr)
      continue
    }

    if len(args) != ficha.cnt_args {
      sErr := serverError{errors.New("wrong count arguments for " + cmd + ". Expected " +  strconv.Itoa(ficha.cnt_args) +  ", got " + strconv.Itoa(len(args)) ),
                            WRONG_ARGS_CNT}
      sendError(conn,  sErr)
      continue
    }
    if ficha.need_token {
      okToken := checkToken(db, args[0], args[1])
      if !okToken {
        sErr := serverError{errors.New("wrong token"), WRONG_TOKEN}
        sendError(conn, sErr)
        continue
      }
    }

    // Run comand
    sErr := ficha.f(conn, args, db)

    if sErr.Err != nil {
      if sErr.Err == io.EOF {
        break
      }
      sendError(conn, sErr)
    } else {
      // command sucsessfuly done
      sendOkStatus(conn)
    }
  }
}
///////////////////////////////////////////////
//YOU FINISHED HERE <-------------------------/
///////////////////////////////////////////////

// register new user by rsa key
func sign_up(conn net.Conn, args []string, db *sql.DB) serverError {
  login := args[0]
  // check this login
  // login must be free
  _, err := getUserID(login, db)
  if err == sql.ErrNoRows {
    // add  new user if everything ok
    // check args

    // RSA key.module must be big.Int()
    n := big.NewInt(0)
    if _, ok := n.SetString(args[1], 10); !ok {
      return serverError{ errors.New("Expected public key module. Got " + args[1]),
                          WRONG_PUBLIC_KEY_MODULE}
    }
    // size of n must be RSA_KEY_LEN bit
    if n.BitLen() != RSA_KEY_LEN  {
      return serverError{ errors.New("Wrong public key module size. Expected " +
                  strconv.Itoa(RSA_KEY_LEN) + " bit, got " +
                  strconv.Itoa(n.BitLen())), WRONG_PUBLIC_KEY_SIZE}
    }

    //check that args[2] is RSA exponent
    e, err := strconv.Atoi(args[2])
    if err != nil {
      return serverError{ errors.New("error convert " + args[2] + " to key exponent"),
                          WRONG_KEY_EXPONENT }
    }

    //register new user
    _, err = db.Exec(`INSERT INTO users(login, pubKey_n, pubKey_e)
            VALUES ($1, $2, $3)`, login, args[1], e)
    return serverError{err, SERVER_INNER_ERR}
  } else {
    return serverError{ errors.New("login " + login + " is used"),
                        LOGIN_IS_USED }
  }

  return NoErrors
}

// sign in by rsa key
func sign_in(conn net.Conn, args []string, db *sql.DB) serverError {
  inText := getRandomText(RANDOM_TEXT_SIZE)
  login := args[0]

  // user with login = $login must be registered
  userID, err := getUserID(login, db)
  if err == sql.ErrNoRows {
    return serverError{ errors.New("No user with login " + login),
                        UNDEFINED_USER }
  } else if err != nil {
    return serverError{err, SERVER_INNER_ERR}
  }

  // ENCRYPT inText
  encData, sErr := encryptForUserById([]byte(inText), userID, db)
  if sErr != NoErrors {
    return sErr
  }
  sendDataB(conn, encData, OK_CODE)

  // check decrypted random texxt from user
  ans, err := recvDataB(conn)
  if err != nil {
    return serverError{err, SERVER_INNER_ERR}
  }
  if inText != string(ans) {
      return serverError{ errors.New("wrong decrypted random_text"),
                          WRONG_RANDOM_TEXT }
  }

  // generate token for user
  bToken := []byte( getRandomText(TOKEN_SIZE) )

  sendDataB(conn, bToken, OK_CODE)
  err = saveToken(db, login, string(bToken))

  if err != nil {
    return serverError{err, SERVER_INNER_ERR}
  }

  return NoErrors
}

func unlogin(conn net.Conn, args []string, db *sql.DB) serverError {
    login := args[0]
    token := args[1]
    err := deleteToken(db, login, token)

    if err != nil {
      return serverError{err, SERVER_INNER_ERR}
    }
    return NoErrors
}

// TODO write message to getter connetion if it open
func sendMsg(conn net.Conn, args []string, db *sql.DB) serverError {
  from := args[0]
  to := args[2]
  msg := args[3]
  sErr := saveNewMsg(db, from, to, msg)

  return sErr
}

// get unread messages
func getNewMsg(conn net.Conn, args []string, db *sql.DB) serverError {
  login := args[0]
  msgs, err := getNewMsgFromDB(db, login)
  if err != nil {
    return serverError{err, SERVER_INNER_ERR}
  }

  // if messages sucsessfuly got
  newMsg := ""
  for i, msg :=range msgs {
    newMsg += msg.ToStr()
    if i != len(msgs)-1 {
      newMsg += MSG_DELIMITER
    }
  }
  sendData(conn, newMsg, 0)
  return NoErrors
}

// find usernames by part of login
// Sorted by position this part in login
func findUsernames(conn net.Conn, args []string, db *sql.DB) serverError {
  loginPart := args[2]
  usernames, err := findUsernamesDB(db, loginPart)
  if err != nil {
    return serverError{err, SERVER_INNER_ERR}
  }
  sendData(conn, usernames, OK_CODE)

  return NoErrors
}












//
