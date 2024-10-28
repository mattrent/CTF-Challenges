import subprocess
import yaml

with open("../challenge.yml", encoding="utf-8") as stream:
    config = yaml.safe_load(stream)

flag = config["flags"][0]

output = subprocess.check_output(
    ["../handout/rev"],
    input=bytes(flag, "utf-8"),
)
