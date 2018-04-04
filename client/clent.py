import socket

RANDOM_TEXT_SIZE = 16   # in server_ex.py
TOKEN_SIZE = 8          # in server_ex.py
LOGIN = "barakuda"

def connect():
    sock = socket.socket()
    sock.connect(('localhost', 5001))
    return sock

def login(sock, name):
    req = "GET_TKN:" + name

    sock.send(bytes(req, "utf-8"))
    rand_text = sock.recv(RANDOM_TEXT_SIZE)
    # DECRYPT
    sock.send(rand_text)
    token = sock.recv(TOKEN_SIZE).decode("utf-8")
    print(token)
    if token == '':
        raise ValueError("INIT ERROR")
    return token

def quit(sock, login, token):
    req = "QUIT:" + login + ";" + token
    print(req)
    sock.send(bytes(req, "utf-8"))


##################################
sock = connect()
name = input()

token = login(sock, name)
input()
quit(sock, token, name)

sock.close()

#sock = socket.socket()
#sock.connect(('localhost', 5001))
#sock.send("GET_TKN:barakuda")
#text = sock.recv(1024).decode("utf-8")
#print(text)















#
