import socket
from constants import *

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
