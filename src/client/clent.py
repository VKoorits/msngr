import socket

from Crypto.Cipher import PKCS1_OAEP
from Crypto.PublicKey import RSA
from Crypto.PublicKey import RSA
import Crypto.Hash
from base64 import b64decode

RANDOM_TEXT_SIZE = 16   # in server_ex.py
TOKEN_SIZE = 8          # in server_ex.py
ANSWER_SIZE = 255
LOGIN = "barakuda"
DELEMITER = ":"
OK_ANSWER = "ok"
OK_CODE = 0
SERVER_HEADER_SIZE = 5
CLIENT_HEADER_SIZE = 4
RSA_KEY_LEN = 1024

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
    sock.connect(('localhost', 5002))
    return sock

def login(sock, login, cipher):
    req = ":".join(["GET_TKN", login])
    sendData(sock, req, datatype="cmd")

    rand_text, code = recvData(sock)
    if code != OK_CODE:
        p(rand_text, code)
        return

    # DECRYPT
    open_text = cipher.decrypt(rand_text)

    sendDataB(sock, open_text, datatype="data")


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
    req  = ":".join(["SEND_MSG", login, token, "getter", "text"])
    sendData(sock, req, datatype="cmd")


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

def sign_up(sock, login, keyPub):
    req = ":".join(["SIGN_UP", login,
            str(keyPub.key.n), str(keyPub.key.e)])
    sendData(sock, req, datatype="cmd")


    res, code = recvData(sock)
    res = res.decode("utf-8")

    if res != OK_ANSWER or code != OK_CODE:
        raise ValueError(res)

    p("SIGN_UP:\t" + res)

##################################


privateKey = RSA.generate(RSA_KEY_LEN)
pubKey = privateKey.publickey()
cipher = PKCS1_OAEP.new(privateKey, hashAlgo=Crypto.Hash.SHA256)


sock = connect()
name = "newuser4"
sign_up(sock, name, pubKey)
token = login(sock, name, cipher)
print("TOKEN:\t\t" + token)
#send_msg(sock, name, token)
#get_new_msg(sock, name, token)
#quit(sock, name, token)

sock.close()
##################################
"""
from Crypto.Cipher import PKCS1_OAEP
from Crypto.PublicKey import RSA
from Crypto.PublicKey import RSA
import Crypto.Hash
from base64 import b64decode

privateKey = RSA.generate(2048)
keyPub = privateKey.publickey()
n = str(keyPub.key.n)
e = str(keyPub.key.e)


sock = connect()


sendData(sock, str(n)+DELEMITER+str(e), datatype="data")
ans, code = recvData(sock)

cipher = PKCS1_OAEP.new(privateKey, hashAlgo=Crypto.Hash.SHA256)
openText = cipher.decrypt(ans)

sendDataB(sock, openText, datatype="data")

#print(dir(privateKey))
#cipher = PKCS1_OAEP.new(privateKey)
#message = cipher.decrypt(ans)

sock.close()
"""





















#