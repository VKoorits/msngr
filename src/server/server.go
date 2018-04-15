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


type feature struct {
  f func (net.Conn, []string, *sql.DB)error
  cnt_args int
  need_token bool
}

type message struct {
  sender string
  text string
  sendingTime string
}

func (msg message) ToStr() string {
  return msg.sender + DELEMITER +
         msg.text + DELEMITER +
         msg.sendingTime
}

var ServerFunctions map[string]feature
var p = fmt.Println
////////////////////////////////////////////

func echo(conn net.Conn){
  defer conn.Close()
  bData, _ := recvDataB(conn)
  sendDataB(conn, bData, OK_CODE)
}

func initServerFunctions() {
  ServerFunctions = make( map[string]feature )
  ServerFunctions["SIGN_UP"] = feature{sign_up, 3, false}
  ServerFunctions["GET_TKN"] = feature{sign_in, 1, false}
  ServerFunctions["QUIT"] = feature{unlogin, 2, true}
  ServerFunctions["SEND_MSG"] = feature{sendMsg, 4, true}
  ServerFunctions["GET_MSG"] = feature{getNewMsg, 2, true}
  ServerFunctions["FIND_USR"] = feature{findUsernames, 3, true}
}

func workWithClient(conn net.Conn, db *sql.DB) {
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
      okToken := checkToken(db, args[0], args[1])
      if !okToken {
        sendError(conn, "wrong token")
        continue
      }
    }

    err = ficha.f(conn, args, db)


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
func sign_up(conn net.Conn, args []string, db *sql.DB) error {
  login := args[0]



  _, err := getUserID(login, db)
  if err == sql.ErrNoRows {
    // add  new user if everything ok
    // check args

    // RSA key.module must be big.Int()
    n := big.NewInt(0)
    if _, ok := n.SetString(args[1], 10); !ok {
      return errors.New("Expected public key module. Get " + args[1])
    }
    // size of n must be RSA_KEY_LEN bit
    if n.BitLen() != RSA_KEY_LEN  {
      return errors.New("Wrong public key module size. Expected " +
                  strconv.Itoa(RSA_KEY_LEN) + " bit, got " +
                  strconv.Itoa(n.BitLen()))
    }

    //check that args[2] is RSA exponent
    e, err := strconv.Atoi(args[2])
    if err != nil {
      return errors.New("error convert " + args[2] + " to key exponent")
    }

    //register new user
    _, err = db.Exec(`INSERT INTO users(login, pubKey_n, pubKey_e)
            VALUES ($1, $2, $3)`, login, args[1], e)
    return err
  } else {
    return errors.New("login " + login + " is used")
  }
  return nil
}

func sign_in(conn net.Conn, args []string, db *sql.DB) error {
  inText := getRandomText(RANDOM_TEXT_SIZE)
  login := args[0]

  userID, err := getUserID(login, db)
  if err == sql.ErrNoRows {
    return errors.New("No user with login " + login)
  } else if err != nil {
    return err
  }

  // ENCRYPT inText
  encData, err := encryptForUserById([]byte(inText), userID, db)
  if err != nil{
    return err
  }
  sendDataB(conn, encData, OK_CODE)

  ans, err := recvDataB(conn)
  if err != nil {
    return err
  }
  if inText != string(ans) {
      return errors.New("wrong decrypted random_text")
  }

  bToken := []byte( getRandomText(TOKEN_SIZE) )
  sendDataB(conn, bToken, OK_CODE)
  err = saveToken(db, login, string(bToken))
  return err
}

func unlogin(conn net.Conn, args []string, db *sql.DB) error {
    login := args[0]
    token := args[1]
    err := deleteToken(db, login, token)
    return err
}

func sendMsg(conn net.Conn, args []string, db *sql.DB) error {
  from := args[0]
  to := args[2]
  msg := args[3]
  err := saveNewMsg(db, from, to, msg)
  return err
}

func getNewMsg(conn net.Conn, args []string, db *sql.DB) error {
  login := args[0]
  msgs, err := getNewMsgFromDB(db, login)
  if err != nil {
    return nil
  }

  newMsg := ""
  for i, msg :=range msgs {
    newMsg += msg.ToStr()
    if i != len(msgs)-1 {
      newMsg += MSG_DELIMITER
    }
  }
  sendData(conn, newMsg, 0)
  return nil
}

func findUsernames(conn net.Conn, args []string, db *sql.DB) error {
  loginPart := args[2]
  usernames, err := findUsernamesDB(db, loginPart)
  if err != nil {
    return err
  }
  sendData(conn, usernames, OK_CODE)

  return nil
}












//
