package main

import (
 "fmt"
 "net"
 "strings"
 "errors"
 "strconv"
 "io"
 "math/big"
 "crypto/rand"
 "crypto/rsa"
 "crypto/sha256"
 "database/sql"
)


type feature struct {
  f func (net.Conn, []string, *sql.DB)error
  cnt_args int
  need_token bool
}

var ServerFunctions map[string]feature
var tokens map[string]string
var p = fmt.Println
////////////////////////////////////////////

func echo(conn net.Conn){
  defer conn.Close()
  bData, _ := recvDataB(conn)
  key := strings.Split(string(bData), DELEMITER)


  n := big.NewInt(0)
  _, ok := n.SetString(key[0], 10)
  if !ok {
    sendError(conn, "Expeted public key module. Get " + key[0])
    return
  }

  e, err := strconv.Atoi(key[1])
  if err != nil {
    sendError(conn, err.Error())
    return
  }

  pub := &rsa.PublicKey{n, e}
  p(pub)
  randText := getRandomText(RANDOM_TEXT_SIZE)

  message := []byte(randText)
  cipherText, err := rsa.EncryptOAEP(sha256.New(),
                          rand.Reader, pub, message, []byte(""))
  sendDataB(conn, cipherText, OK_CODE)
  openData, err := recvDataB(conn)
  p(openData)
  p(string(openData) == randText)
}

func initServerFunctions() {
  ServerFunctions = make( map[string]feature )
  tokens  = make(map[string]string)
  ServerFunctions["SIGN_UP"] = feature{sign_up, 3, false}
  ServerFunctions["GET_TKN"] = feature{sign_in, 1, false}
  ServerFunctions["QUIT"] = feature{unlogin, 2, true}
  ServerFunctions["SEND_MSG"] = feature{sendMsg, 4, true}
  ServerFunctions["GET_MSG"] = feature{getNewMsg, 2, true}
  ServerFunctions["TEST"] = feature{test, 2, true}
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
      okToken := checkToken(args[0], args[1])
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

  //TODO other checks for correct login (SIGN_UP)
  if len(login) < MIN_LOGIN_LEN || len(login) > MAX_LOGIN_LEN {
    return errors.New("Wrong login len")
  }

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



func checkToken(login string, token string) bool {
  //TODO mutex or atomic or DB
  return tokens[login] == token
}

func encryptForUserById(inData []byte, UserID int, db *sql.DB) ([]byte, error) {
  row := db.QueryRow("SELECT pubKey_n, pubKey_e FROM users WHERE user_id = $1", UserID)

  var n_str string
  var e int
  err := row.Scan(&n_str, &e)
  if err != nil {
    return nil, err
  }

  //create Key
  // n is big.Int; checked in sign_up
  n := big.NewInt(0)
  n.SetString(n_str, 10)

  pub := &rsa.PublicKey{n, e}
  encData, err := rsa.EncryptOAEP(sha256.New(),
                          rand.Reader, pub, inData, []byte(""))
  if err != nil {
    return nil, err
  }

  return encData, err
}

func sign_in(conn net.Conn, args []string, db *sql.DB) error {
  inText := getRandomText(RANDOM_TEXT_SIZE)
  //inData := []byte(inText)

  login := args[0]
  userID, err := getUserID(login, db)
  if err == sql.ErrNoRows {
    return errors.New("No user with login " + login)
  } else if err != nil {
    return err
  }
  p(userID)

  // ENCRYPT inData || inText
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


  tokens[args[0]] = string(bToken)

  return nil
}

func unlogin(conn net.Conn, args []string, db *sql.DB) error {
    //token := args[0]
    login := args[1]

    //TODO mutex or atomic or DB
    delete(tokens, login)

    return nil
}

func test(conn net.Conn, args []string, db *sql.DB) error {
  return nil
}

func sendMsg(conn net.Conn, args []string, db *sql.DB) error {
  return nil
}

func getNewMsg(conn net.Conn, args []string, db *sql.DB) error {
  return nil
}
