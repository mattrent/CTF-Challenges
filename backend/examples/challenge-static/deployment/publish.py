import os
import shutil
import requests
import urllib3
import yaml

host = os.environ["DEPLOYER_URL"]
username = os.environ["DEPLOYER_USERNAME"]
password = os.environ["DEPLOYER_PASSWORD"]

shutil.make_archive("handout", "zip", "../handout/")

s = requests.session()
s.verify=False
urllib3.disable_warnings(urllib3.exceptions.InsecureRequestWarning)

# load config
with open("../challenge.yml", encoding="utf-8") as stream:
    config = yaml.safe_load(stream)

# login
r = s.post(host + "/users/login", json={ "username": username, "password": password }, timeout=20)
print("login:", r.status_code, r.content)
r.raise_for_status()
s.headers = {"Authorization": "Bearer " + r.json().get("token")}

# add challenge
r = s.post(host + "/challenges", files=[
    ("upload[]", open("../challenge.yml", "rb")),
    ("upload[]", open("handout.zip", "rb"))], timeout=20)
print("add challenge:", r.status_code, r.content)
r.raise_for_status()
challenge_id = r.json().get("challengeid")

# publish challenge to CTFd
r = s.post(host + "/challenges/" + challenge_id + "/publish", timeout=20)
print("publish:", r.status_code, r.content)
