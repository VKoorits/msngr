from constants import *
from kernel.network import *

# get from server random text, encrypted by user`s public key
# decrypt this text with private key and send to server
# if no decryption error, user will get session token,
# which he must put into all packages, that he send to server
# return token, operation_code
# if some error was, token is error descrtption
def login(sock, login, cipher):
    req = DELEMITER.join(["GET_TKN", login])
    sendData(sock, req, datatype="cmd")

    # get random text, encrypted by public key
    rand_text, code = recvData(sock)
    if code != OK_CODE:
        return rand_text, code

    # decrypt random text
    try:
        open_text = cipher.decrypt(rand_text)
    except ValueError as e:
        return "Decyptyon error", WRONG_DECRYPTION

    # send decrypted text
    sendDataB(sock, open_text, datatype="data")

    # get sesion token
    token, code = recvData(sock)
    token = token.decode("utf-8")
    if len(token) != TOKEN_SIZE  or  code != OK_CODE:
        return token, code


    res, code = recvData(sock)
    res = res.decode("utf-8")
    if res != OK_ANSWER or code != OK_CODE:
        return res, code

    return token, OK_CODE

# delete session token from server
def quit(sock, login, token):
    req = DELEMITER.join(["QUIT", login, token])
    sendData(sock, req, datatype="cmd")

    res, code = recvData(sock)
    res = res.decode("utf-8")

    if res != OK_ANSWER or code != OK_CODE:
        raise ValueError(res)

    p("QUIT:\t\t" + res)

# send new message
# message geteer must be registered
# return err_description, err_code
def send_msg(sock, login, token, getter, msg):
    req  = DELEMITER.join(["SEND_MSG", login, token, getter, msg])
    sendData(sock, req, datatype="cmd")

    res, code = recvData(sock)
    res = res.decode("utf-8")

    if res != OK_ANSWER or code != OK_CODE:
        return res, code
    return "", OK_CODE

# get list new messages
# return msgs, error_code
def get_new_msg(sock, login, token):
    req = DELEMITER.join(["GET_MSG", login, token])
    sendData(sock, req, datatype="cmd")

    msgs, code = recvData(sock)
    if code != OK_CODE:
        return msgs, code


    res, code = recvData(sock)
    res = res.decode("utf-8")

    if res != OK_ANSWER or code != OK_CODE:
        return res, code

    # translate msgs to list
    msgs = msgs.decode("utf-8")
    msgs = msgs.split(MSG_DELEMITER)
    msgs = list(map(lambda x: x.split(DELEMITER), msgs))
    if len(msgs[0]) !=3:
        msgs = []

    return msgs, code

# register new user by his rsa key
# return err_description, err_code
def sign_up(sock, login, keyPub):
    req = DELEMITER.join(["SIGN_UP", login,
            str(keyPub.key.n), str(keyPub.key.e)])
    sendData(sock, req, datatype="cmd")


    answer, code = recvData(sock)
    answer = answer.decode("utf-8")

    if answer != OK_ANSWER or code != OK_CODE:
        return answer, code

    return "", OK_CODE

# find users, that registered in messenger by login or part of login
# return
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

    # to list
    users = users.split(DELEMITER)
    if users[0] == '':
        users = []

    return users, OK_CODE
