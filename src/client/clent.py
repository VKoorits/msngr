import socket
import random
import os
import sys

from Crypto.Cipher import PKCS1_OAEP
from Crypto.PublicKey import RSA
from Crypto.PublicKey import RSA
import Crypto.Hash
from base64 import b64decode


SERVER_ADDR = 'localhost'
SERVER_PORT = 5002
RANDOM_TEXT_SIZE = 16   # in server_ex.py
TOKEN_SIZE = 8          # in server_ex.py
ANSWER_SIZE = 255
DELEMITER = "│"
MSG_DELEMITER = "|"
OK_ANSWER = "ok"
SERVER_HEADER_SIZE = 5
CLIENT_HEADER_SIZE = 4
RSA_KEY_LEN = 1024
########################################
OK_CODE = 0
WRONG_DECRYPTION = 255


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
    sock = None
    try:
        sock = socket.socket()
        sock.connect((SERVER_ADDR, SERVER_PORT))
    except:
        print("Server " + SERVER_ADDR + ":" +  str(SERVER_PORT) + "is not available")
        print("Please try again later")
        sock = None
    return sock

##################################

def login(sock, login, cipher):
    req = DELEMITER.join(["GET_TKN", login])
    sendData(sock, req, datatype="cmd")

    rand_text, code = recvData(sock)
    if code != OK_CODE:
        return rand_text, code

    # DECRYPT
    try:
        open_text = cipher.decrypt(rand_text)
    except ValueError as e:
        return "Decyptyon error", WRONG_DECRYPTION

    sendDataB(sock, open_text, datatype="data")


    token, code = recvData(sock)
    token = token.decode("utf-8")
    if len(token) != TOKEN_SIZE  or  code != OK_CODE:
        return token, code


    res, code = recvData(sock)
    res = res.decode("utf-8")
    if res != OK_ANSWER or code != OK_CODE:
        return res, code

    return token, OK_CODE

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
        return res, code
    return "", OK_CODE


def get_new_msg(sock, login, token):
    req = DELEMITER.join(["GET_MSG", login, token])
    sendData(sock, req, datatype="cmd")

    msgs, code = recvData(sock)
    if code != OK_CODE:
        return msgs, code
    msgs = msgs.decode("utf-8")


    res, code = recvData(sock)
    res = res.decode("utf-8")

    if res != OK_ANSWER or code != OK_CODE:
        return res, code
    return msgs, code


def sign_up(sock, login, keyPub):
    req = DELEMITER.join(["SIGN_UP", login,
            str(keyPub.key.n), str(keyPub.key.e)])
    sendData(sock, req, datatype="cmd")


    answer, code = recvData(sock)
    answer = answer.decode("utf-8")

    if answer != OK_ANSWER or code != OK_CODE:
        return answer, code

    return "", OK_CODE

def find_users(sock, login, token, login_part):
    req = DELEMITER.join(["FIND_USR", login, token, login_part])
    sendData(sock, req, datatype="cmd")


    users, code = recvData(sock)
    if code != OK_CODE:
        return users, code
    users = users.decode("utf-8")


    res, code = recvData(sock)
    res = res.decode("utf-8")

    if res != OK_ANSWER or code != OK_CODE:
        return res, code
    return users, OK_CODE

##################################
def error_worker(errText, code):
    if code != OK_CODE:
        print("ERROR(" + str(code) + "): ", end="")
        print(errText)
    return code


def get_key(name, sock):
    # read or generate key
    fileName = name + ".pem"
    privateKey, pubKey = 0, 0
    code = OK_CODE
    try:
        #read
        keyText = open(fileName, 'r').read()
        privateKey = RSA.importKey(keyText)
        pubKey = privateKey.publickey()
    except FileNotFoundError as e:
        #generate
        privateKey = RSA.generate(RSA_KEY_LEN)
        pubKey = privateKey.publickey()
        # save key
        keyStr = privateKey.exportKey("PEM").decode("ascii")
        with open(fileName, "w") as f:
            f.write(keyStr)

        if input("Do yo want to register[Y/n]? ") == "Y":
            #registered
            answer, code = sign_up(sock, name, pubKey)
            code = error_worker(answer, code)
            if code == OK_CODE:
                print(name + " was sucsessfuly registered")
        else:
            #delete key
            os.remove(fileName)
            return None, None

    cipher = PKCS1_OAEP.new(privateKey, hashAlgo=Crypto.Hash.SHA256)
    return pubKey, cipher


def print_msgs(msgs):
    msgs = msgs.split(MSG_DELEMITER)
    msgs = list(map(lambda x: x.split(DELEMITER), msgs))
    if len(msgs[0]) !=3:
        print("No new messages")
    else:
        for msg in msgs:
            print("From " + msg[0])
            print("\t" + msg[1])
            print("\t" + msg[2].split(" ")[1])

def print_users(users, request):
    users = users.split(DELEMITER)
    if users[0] != '':
        i = 1
        for user in users:
            print(str(i) + "\t:" + user)
            i += 1
    else:
        print("not found users request '" + request + "'")

def undefined_behavior():
    print("Something went wrong.")
    print("You can send the report to the address koorits.viktor@yandex.ru")

def get_choose():
    print("hello, now you can:")
    print("\t1)get new messages")
    print("\t2)send message")
    print("\t3)find users")
    print("\t4)quit")
    print("\t5)listen")

    choose = 0
    while True:
        try:
            choose = int(input())
        except:
            pass
        if choose not in range(1,6):
            print("enter number between 1 and 5")
        else:
            return choose
#############################################
def interface_get_msg(sock, name, token):
    msgs, code = get_new_msg(sock, name, token)

    code = error_worker(msgs, code)
    if code != OK_CODE:
        undefined_behavior()
        return code

    print_msgs(msgs)
    return OK_CODE

def interface_send_msg(sock, name, token):
    text = input("text: ")
    getter = input("getter: ")
    answer, code = send_msg(sock, name, token, getter, text)

    code = error_worker(answer, code)
    if code != OK_CODE:
        undefined_behavior()
        return code
    return OK_CODE

def interface_find_users(sock, name, token):
    loginPart = input("login or part of login: ")
    users, code = find_users(sock, name, token, loginPart)

    code = error_worker(users, code)
    if code != OK_CODE:
        undefined_behavior()
        return code

    print_users(users, loginPart)
    return OK_CODE

def interface_quit(sock, name, token):
    return -1

def interface_listen(sock, name, token):
    while True:
        data, code = recvData(sock)
        data = data.decode("utf-8")
        print_msgs(data)
    return OK_CODE



#############################################
def interface():
    sock = connect()
    if sock is None:
        return

    name = input("Enter your login: ")
    pubKey, cipher = get_key(name, sock)
    if pubKey is None or cipher is None:
        return

    token, code = login(sock, name, cipher)
    code = error_worker(token, code)
    if code != OK_CODE:
        undefined_behavior()
        return

    clientFunctions = [None,
                        interface_get_msg,
                        interface_send_msg,
                        interface_find_users,
                        interface_quit,
                        interface_listen]
    while True:
        choose = get_choose()
        code = clientFunctions[choose](sock, name, token)
        if code != OK_CODE:
            break





interface()


"""
sock = connect()
name = sys.argv[1]
pubKey, cipher = get_key(name, sock)

input()
token = login(sock, name, cipher)
get_new_msg(sock, name, token)

#text = input("text: ")
#getter = input("getter: ")
#send_msg(sock, name, token, getter, text)
find_users(sock, name, token, "ya")
quit(sock, name, token)




sock.close()
"""




















#
