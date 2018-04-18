package main

import (
  "fmt"
  "net"
  "math/rand"
  "encoding/binary"
  "errors"
  "bytes"
  "strconv"
)


func sendError(conn net.Conn, sErr serverError) {
  errText := sErr.Err.Error()
  fmt.Println(ERROR_BEGIN + errText)
  if sErr.Code == SERVER_INNER_ERR {
    errText = "server inner error"
  }
  sendData(conn, errText, sErr.Code )
}

func getRandomText(text_len int) string {
  //TODO rand seed
  // without ';' and ':'
  chars := "QWERTYUIOPASDFGHJKLZXCVBNMqwertyuiopasdfghjklzzxcvbnm_1234567890!@#$%^&*()_=+,.<>/?[{}]"
  res := ""
  for i := 0; i < text_len; i += 1 {
    num := rand.Intn(len(chars))
    res += chars[num:num+1]
  }
  return res
}

func joinIntSLice(list []int, dilemitre string) string {
  res := ""
  for i, num := range list {
    res += strconv.Itoa(num)
    if i < len(list) - 1 {
      res += ","
    }
  }
  return res
}

func sendOkStatus(conn net.Conn) {
  sendData(conn, OK_ANSWER, uint8(OK_CODE))
}

func sendDataB(conn net.Conn, data []byte, code uint8) {
  dataSize := uint32(len(data))
  header := make([]byte, SERVER_HEADER_SIZE)
  binary.LittleEndian.PutUint32(header, dataSize)
  header[4] = code
  res := append(header, data...)
  conn.Write(res)
}

func sendData(conn net.Conn, textData string, code uint8) {
  sendDataB(conn, []byte(textData), code)
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
