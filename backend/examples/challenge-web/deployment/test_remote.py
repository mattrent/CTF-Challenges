import time
import shutil
import requests
import urllib3
import yaml
import os

host = os.environ["DEPLOYER_URL"]
username = os.environ["DEPLOYER_USERNAME"]
password = os.environ["DEPLOYER_PASSWORD"]

shutil.make_archive("handout", "zip", "../handout/")
shutil.make_archive("challenge", "zip", "../src/")

s = requests.session()
s.verify=False
urllib3.disable_warnings(urllib3.exceptions.InsecureRequestWarning)

# load config
with open("../challenge.yml", encoding="utf-8") as stream:
    config = yaml.safe_load(stream)

# login
login_with_keycloak = False
if login_with_keycloak:
    r = s.post("https://ctf.sdu.dk/keycloak/realms/ctf/protocol/openid-connect/token/", data={"client_id":"deployer", "username": username, "password": password, "grant_type": "password", "scope": "openid"}, timeout=20)
    print("login:", r.status_code, r.content)
    r.raise_for_status()
    s.headers = {"Authorization": "Bearer " + r.json().get("access_token")}
else:
    r = s.post(host + "/users/login", json={ "username": username, "password": password }, timeout=20)
    print("login:", r.status_code, r.content)
    r.raise_for_status()
    s.headers = {"Authorization": "Bearer " + r.json().get("token")}

# add challenge
r = s.post(host + "/challenges", files=[
    ("upload[]", open("challenge.zip", "rb")),
    ("upload[]", open("../challenge.yml", "rb")),
    ("upload[]", open("handout.zip", "rb"))], timeout=20)
print("add challenge:", r.status_code, r.content)
r.raise_for_status()
challenge_id = r.json().get("challengeid")

# start challenge
r = s.post(host + "/challenges/" + challenge_id + "/start", timeout=20)
print("start challenge:", r.status_code, r.content)
r.raise_for_status()
url = r.json().get("url")

# interact with challenge
for i in range(30):
    try:
        r = requests.get("http://" + url, verify=False, timeout=20)
        print("test challenge:", r.status_code, r.content)
        r.raise_for_status()
        break
    except Exception as e:
        print("test challenge failed:", r.status_code, r.content, e)
        time.sleep(5)

# check flag
print(r.text)
if config["flags"][0] in r.text:
    print("Flag found:", config["flags"][0])
else:
    raise Exception("Flag not found")

# stop challenge
r = s.post(host + "/challenges/" + challenge_id + "/stop", timeout=20)
print("stop challenge:", r.status_code, r.content)
r.raise_for_status()

# publish challenge to CTFd
# r = s.post(host + "/challenges/" + challenge_id + "/publish", timeout=20)
# print("publish:", r.status_code, r.content)
