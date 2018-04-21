package main

import (
  _ "github.com/mattn/go-sqlite3"
  "database/sql"
  "strconv"
  "time"
  "errors"
)

func getUserID(login string, db *sql.DB) (int, error) {
  row := db.QueryRow("SELECT (user_id) FROM users WHERE login = $1", login)

  var id int
  err := row.Scan(&id)
  return id, err
}

func saveToken(db *sql.DB, login string, token string) error {
  row := db.QueryRow("SELECT (token) FROM tokens WHERE login = $1", login)
  var buf string

  now := time.Now()
  end_period := now.Add(time.Second*TOKEN_TTL_SECONDS)
  req := ""

  err := row.Scan(&buf)
  if err == sql.ErrNoRows {
    req = `INSERT INTO tokens(login, token, period_work)
                      VALUES($1, $2, $3)`
    _, err = db.Exec(req, login, token, end_period.Format(TIME_FORMAT))
  } else {
    req = `UPDATE tokens
              SET token=$1, period_work=$2
              WHERE login = $3`
    _, err = db.Exec(req, token,  end_period.Format(TIME_FORMAT), login)
  }

  return err
}

func deleteToken(db *sql.DB, login string, token string) error {
  req := `DELETE FROM tokens
              WHERE login=$1 AND token=$2`
  _, err := db.Exec(req, login, token)
  return err
}

func checkToken(db *sql.DB, login string, token string) bool {
  row := db.QueryRow("SELECT token, period_work FROM tokens WHERE login = $1", login)

  var tokenFromDB string
  var timeStr string
  err := row.Scan(&tokenFromDB, &timeStr)
  if err == sql.ErrNoRows {
    return false
  } else if err != nil {
    p("ERROR CHECK LOGIN", err)
    return false
  }
  endPeriod, err := time.Parse(TIME_FORMAT, timeStr)
  if err != nil {
    p("ERROR CHECK LOGIN", err)
    return false
  }

  if endPeriod.Before(time.Now()) {
    p("old TOKEN")
    deleteToken(db, login, token)
    return false
  }

  return tokenFromDB == token
}

func saveNewMsg(db *sql.DB, from string, to string, msg string) serverError {
  _, err := getUserID(to, db)
  if err == sql.ErrNoRows {
    return serverError{ errors.New("User " + to + " is not registered"),
                        GETTER_NOT_REGISTERED }
  } else if err != nil {
    return serverError{err, SERVER_INNER_ERR}
  }

  req := `INSERT INTO messages(sender, getter, msg, date, isNew)
                    VALUES($1, $2, $3, $4, 1)`
  _, err = db.Exec(req, from, to, msg, time.Now().Format(TIME_FORMAT), 1)
  return serverError{nil, OK_CODE}
}

func getNewMsgFromDB(db *sql.DB, login string) ([]message, error) {
  rows, err := db.Query(`SELECT msg_id, sender, msg, date
    FROM messages WHERE getter=$1 AND isNew=1`, login)
  if err != nil {
    return nil, err
  }
  defer rows.Close()

  var msg message
  msgs := make([]message, 0)
  msgsID := make([]int, 0)
  var thisID int
  for rows.Next() {
    err = rows.Scan(&thisID, &msg.sender, &msg.text, &msg.sendingTime);
    if err != nil {
      return nil, err
    }
    msgs = append(msgs, msg)
    msgsID = append(msgsID, thisID)
  }

  listID := joinIntSLice(msgsID, ",")

  req := `UPDATE messages SET isNew=0 WHERE msg_id IN(` + listID + `)`
  _, err = db.Exec(req)
  if err != nil {
    return nil, err
  }

  return msgs, nil
}

func findUsernamesDB(db *sql.DB, loginPart string) (string, error) {
  req := `SELECT login FROM users
          WHERE INSTR(login, $1)
          ORDER BY INSTR(login, $1)`
  rows, err := db.Query(req, loginPart)
  if err != nil {
    return "", err
  }
  defer rows.Close()

  usernames := ""
  var oneLogin string
  for rows.Next() {
    err = rows.Scan(&oneLogin);
    if err != nil {
      return "", err
    }
    usernames += oneLogin + DELEMITER
  }

  slice := []rune(usernames)
  if len(slice) != 0 {
    slice = slice[:len(slice) - 1]
  }
  return string(slice), nil
}

func InitDB(filepath string) *sql.DB {
	db, err := sql.Open("sqlite3", filepath)
	if err != nil { panic(err) }
	if db == nil { panic("db nil") }

  req := `CREATE TABLE IF NOT EXISTS users(
             user_id INTEGER PRIMARY KEY  AUTOINCREMENT,
      			 login VARCHAR(32) NOT NULL,
      			 pubKey_n VARCHAR(1024) NOT NULL,
      			 pubKey_e INTEGER NOT NULL)`
  _, err = db.Exec(req)
  if err != nil { panic(err) }


  req = `CREATE TABLE IF NOT EXISTS tokens(
            login VARCHAR(32) PRIMARY KEY NOT NULL,
            token CHAR(` + strconv.Itoa(TOKEN_SIZE) + `) NOT NULL,
            period_work DATATIME NOT NULL)`
  _, err = db.Exec(req)
  if err != nil { panic(err) }


  req = `CREATE TABLE IF NOT EXISTS messages(
            msg_id INTEGER PRIMARY KEY  AUTOINCREMENT,
            sender VARCHAR(32) NOT NULL,
            getter VARCHAR(32) NOT NULL,
            msg TEXT NOT NULL,
            date DATATIME NOT NULL,
            isNew TINYINT(1) NOT NULL)`
  _, err = db.Exec(req)
  if err != nil { panic(err) }



	return db
}
