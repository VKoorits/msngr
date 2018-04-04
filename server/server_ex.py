import socket
import random
import threading
import time
from queue import Queue


RANDOM_TEXT_SIZE = 16
TOKEN_SIZE = 8
COUNT_THREADS = 2
conns = Queue()

def get_token(conn, name):
    in_data = bytes( get_random_text(RANDOM_TEXT_SIZE), "utf-8" )
    # ENCRYPT
    conn.send( in_data )
    out_data = conn.recv(RANDOM_TEXT_SIZE)

    if in_data != out_data:
        raise ValueError("WRONG KEY")

    token = get_random_text(TOKEN_SIZE)
    conn.send( bytes(token, "utf-8") )
    tokens[name] = token


def unlogin(conn, args):
    login = args.split(";")[1]
    token = args.split(";")[0]

    if tokens.get(login) == token:
        del tokens[login]



tokens = {}
server_functions = {"GET_TKN":get_token, "QUIT":unlogin}


def get_random_text(text_len):

    chars = "QWERTYUIOPASDFGHJKLZXCVBNMqwertyuiopasdfghjklzzxcvbnm_1234567890!@#$%^&*()_=+,.<>/?[{}]"
    res = ""
    for i in range(text_len):
        num = random.randint(0, len(chars)-1)
        res += chars[num]
    return res

def work_with_client(conn) :
    print ('connected')
    z = 0
    while True:
        data = conn.recv(1024)
        if not data:
            break

        text = data.decode("utf-8")
        cmd = text.split(':')[0]
        args = text[len(cmd)+1:]
        try:
            server_functions[cmd](conn, args)
            print(tokens)
        except ValueError as e:
            print("ERROR")
            conn.close()
            return


def work_with_tasks(e1, e2):
    global conns
    while True:
        time.sleep(1)
        conn = conns.get()
        work_with_client(conn)

###################################
sock = socket.socket()
sock.bind(('', 5001))
sock.listen(10)

e1 = threading.Event()
e2 = threading.Event()

t = [threading.Thread(target=work_with_tasks, args=(e1, e2)) for i in range(COUNT_THREADS) ]
for q in t:
    q.start()

# обработка соединений
while True:
    conn, addr = sock.accept()
    conns.put(conn)
















#
