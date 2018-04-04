from Crypto.PublicKey import RSA
import random

def new_keys():
    # каждый раз одинаковые
    privatekey = RSA.generate(1024)
    publickey = privatekey.publickey()
    return (privatekey, publickey)

def write_keys(privatekey, publickey):
    with open('private.key','wb') as f:
        f.write(bytes(privatekey.exportKey()))
    with open('public.key','wb') as  f:
        f.write(bytes(publickey.exportKey()))

def read_keys():
    privatekey = RSA.importKey(open('private.key','rb').read())
    publickey = RSA.importKey(open('public.key','rb').read())
    return (privatekey, publickey)

(private, public) = new_keys()
enc_data = public.encrypt('abcdefgh', "offset")
