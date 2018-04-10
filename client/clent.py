import socket


RANDOM_TEXT_SIZE = 16   # in server_ex.py
TOKEN_SIZE = 8          # in server_ex.py
ANSWER_SIZE = 255
LOGIN = "barakuda"
DELEMITER = ":"
OK_ANSWER = "ok"
OK_CODE = 0
SERVER_HEADER_SIZE = 5
CLIENT_HEADER_SIZE = 4

p = print
# p     - debug
# print - interface

def recvData(sock):
    header = sock.recv(SERVER_HEADER_SIZE)
    data_size = int.from_bytes(header[0:4], byteorder='little')
    code = int.from_bytes(header[4:5], byteorder='little')

    data = sock.recv(data_size)
    return data, code

def sendDataB(sock, data, datatype):
    req = (len(data)).to_bytes(4, byteorder='little') + data
    sock.send(req)

def sendData(sock, data, datatype):
    sendDataB(sock, bytes(data, "utf-8"), datatype)


def connect():
    sock = socket.socket()
    sock.connect(('localhost', 5001))
    return sock

def login(sock, login):
    req = ":".join(["GET_TKN", login])
    sendData(sock, req, datatype="cmd")

    rand_text, code = recvData(sock)
    if code != OK_CODE:
        p(rand_text, code)
        return
    # DECRYPT
    #rand_text = b"popa"
    sendDataB(sock, rand_text, datatype="data")



    token, code = recvData(sock)
    token = token.decode("utf-8")
    if code != OK_CODE:
        p(token, code)
        return

    if token == '':
        raise ValueError("INIT ERROR, empty token")

    res, code = recvData(sock)
    res = res.decode("utf-8")
    if res != OK_ANSWER or code != OK_CODE:
        raise ValueError(res)
    p("LOGIN:\t\t" + res)
    return token

def quit(sock, login, token):
    req = ":".join(["QUIT", login, token])
    sendData(sock, req, datatype="cmd")

    res, code = recvData(sock)
    res = res.decode("utf-8")

    if res != OK_ANSWER or code != OK_CODE:
        raise ValueError(res)

    p("QUIT:\t\t" + res)

def send_msg(sock, login, token):
    msg = input().rstrip()
    getter = input().rstrip()
    req  = ":".join(["SEND_MSG", login, token, getter, msg])
    sock.send( bytes(req, "utf-8") )
    res, code = recvData(sock)
    res = res.decode("utf-8")

    if res != OK_ANSWER or code != OK_CODE:
        raise ValueError(res)

    p("SEND:\t\t" + res)

def get_new_msg(sock, login, token):
    req = ":".join(["GET_MSG", login, token])
    sendData(sock, req, datatype="cmd")


    res, code = recvData(sock)
    res = res.decode("utf-8")

    if res != OK_ANSWER or code != OK_CODE:
        raise ValueError(res)

    p("GET_MSG:\t" + res)

def test(sock, login, token):
    req = ":".join(["TEST", login, token])
    sendData(sock, req, datatype="cmd")


    res, code = recvData(sock)
    res = res.decode("utf-8")

    if res != OK_ANSWER or code != OK_CODE:
        raise ValueError(res)

    p("TEST:\t\t" + res)


##################################
sock = connect()
name = "viktor"

token = login(sock, name)
test(sock, name, token)
#send_msg(sock, name, token)
get_new_msg(sock, name, token)
quit(sock, token, name)

sock.close()

















#
