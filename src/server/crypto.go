package main

import (
  "math/big"
  "crypto/rand"
  "crypto/rsa"
  "crypto/sha256"
  "database/sql"
)


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
