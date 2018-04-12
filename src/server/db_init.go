package main

import (
  _ "github.com/mattn/go-sqlite3"
  "database/sql"
)

func getUserID(login string, db *sql.DB) (int, error) {
  row := db.QueryRow("SELECT (user_id) FROM users WHERE login = $1", login)

  var id int
  err := row.Scan(&id)
  return id, err
}

func InitDB(filepath string) *sql.DB {
	db, err := sql.Open("sqlite3", filepath)
	if err != nil { panic(err) }
	if db == nil { panic("db nil") }

  req := `CREATE TABLE IF NOT EXISTS users(
             user_id INTEGER PRIMARY KEY     AUTOINCREMENT,
      			 login VARCHAR(32) NOT NULL,
      			 pubKey_n VARCHAR(1024) NOT NULL,
      			 pubKey_e INTEGER NOT NULL)`
  _, err = db.Exec(req)
  if err != nil { panic(err) }

  /*
  req = `INSERT INTO users(login, pubKey_n, pubKey_e)
          VALUES ('viktor', '24152877953784267419304520481391018473737012153547915063558720733535328430582535788180199160617420046215519509192721232166713427917547312803942953634194619494381283260953592961695570130493368960604825458221516073405416236037785080969005761124467961465672351048410009604379037134243563347689259741664015351938085386191808082024277796668765357135805716602181192706292630181459057204996857339055423004846734999447916606926094426203567472397977601822454190644905448519260866841695013694767666250024824578538107054872004109290393007307045954111450350709384025777312917009139303662667790540397589411720813748979377665742207', 65537)`
  _, err = db.Exec(req)
  //p(req)
  if err != nil { panic(err) }
  */



	return db
}