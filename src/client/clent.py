import socket
import random
import os
import sys

from Crypto.Cipher import PKCS1_OAEP
from Crypto.PublicKey import RSA
from Crypto.PublicKey import RSA
import Crypto.Hash
from base64 import b64decode

RANDOM_TEXT_SIZE = 16   # in server_ex.py
TOKEN_SIZE = 8          # in server_ex.py
ANSWER_SIZE = 255
LOGIN = "barakuda"
DELEMITER = "│"
MSG_DELEMITER = "|"
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

##################################

def login(sock, login, cipher):
    req = DELEMITER.join(["GET_TKN", login])
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
    req = DELEMITER.join(["QUIT", login, token])
    sendData(sock, req, datatype="cmd")

    res, code = recvData(sock)
    res = res.decode("utf-8")

    if res != OK_ANSWER or code != OK_CODE:
        raise ValueError(res)

    p("QUIT:\t\t" + res)

def send_msg(sock, login, token, getter, msg):
    req  = DELEMITER.join(["SEND_MSG", login, token, getter, msg])
    sendData(sock, req, datatype="cmd")


    res, code = recvData(sock)
    res = res.decode("utf-8")

    if res != OK_ANSWER or code != OK_CODE:
        raise ValueError(res)

    p("SEND:\t\t" + res)

def get_new_msg(sock, login, token):
    req = DELEMITER.join(["GET_MSG", login, token])
    sendData(sock, req, datatype="cmd")

    msgs, code = recvData(sock)
    if code != OK_CODE:
        raise ValueError(res)
    msgs = msgs.decode("utf-8")


    res, code = recvData(sock)
    res = res.decode("utf-8")

    if res != OK_ANSWER or code != OK_CODE:
        raise ValueError(res)
    print_msgs(msgs)
    p("GET_MSG:\t" + res)

def sign_up(sock, login, keyPub):
    req = DELEMITER.join(["SIGN_UP", login,
            str(keyPub.key.n), str(keyPub.key.e)])
    sendData(sock, req, datatype="cmd")


    res, code = recvData(sock)
    res = res.decode("utf-8")

    if res != OK_ANSWER or code != OK_CODE:
        raise ValueError(res)

    p("SIGN_UP:\t" + res)

##################################

def get_key(name, sock):
    fileName = name + ".pem"
    privateKey, pubKey = 0, 0
    newKey = False #?!
    try:
        keyText = open(fileName, 'r').read()
        privateKey = RSA.importKey(keyText)
        pubKey = privateKey.publickey()
    except:
        privateKey = RSA.generate(RSA_KEY_LEN)
        pubKey = privateKey.publickey()
        keyStr = privateKey.exportKey("PEM").decode("ascii")
        with open(fileName, "w") as f:
            f.write(keyStr)
        newKey = True

    if newKey:
        try:
            sign_up(sock, name, pubKey)
        except ValueError as e:
            if str(e) != "?ERROR: login " + name + " is used":
                print("Registration error:")
                print("\t"+ str(e))
            else:
                print("Where is your key?!")
                path = os.path.join(os.path.abspath(os.path.dirname(__file__)), filename)
                os.remove(path)
            os._exit(0)

    cipher = PKCS1_OAEP.new(privateKey, hashAlgo=Crypto.Hash.SHA256)
    return pubKey, cipher


def print_msgs(msgs):
    msgs = msgs.split(MSG_DELEMITER)
    msgs = list(map(lambda x: x.split(DELEMITER), msgs))
    if len(msgs[0]) !=3:
        print("No messages")
    else:
        for msg in msgs:
            print("From " + msg[0])
            print("\t" + msg[1])
            print("\t" + msg[2].split(" ")[1])

#############################################


name = sys.argv[1]
sock = connect()
pubKey, cipher = get_key(name, sock)

token = login(sock, name, cipher)
get_new_msg(sock, name, token)

text = input("text: ")
getter = input("getter: ")
send_msg(sock, name, token, getter, text)
quit(sock, name, token)


sock.close()






















#
