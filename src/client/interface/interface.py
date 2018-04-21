from kernel.network import recvData, sendData, connect
from kernel.kernel import (login, quit, send_msg,
        get_new_msg, sign_up, find_users)
from constants import *

import os

from Crypto.Cipher import PKCS1_OAEP
from Crypto.PublicKey import RSA
from Crypto.PublicKey import RSA
import Crypto.Hash


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
    if len(msgs) == 0:
        print("No new messages")
    else:
        for msg in msgs:
            print("From " + msg[0])
            print("\t" + msg[1])
            print("\t" + msg[2].split(" ")[1])

def print_users(users, request):
    if len(users) > 0:
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
        msgs, code = recvData(sock)

        if code == OK_CODE:
            msgs = msgs.decode("utf-8")
            msgs = msgs.split(MSG_DELEMITER)
            msgs = list(map(lambda x: x.split(DELEMITER), msgs))
            if len(msgs[0]) !=3:
                msgs = []

            print_msgs(msgs)
        else:
            print("interface_listen")
            print("\tGot: ", (msgs, code))
            return code
    return OK_CODE


#############################################
def console_interface():
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
