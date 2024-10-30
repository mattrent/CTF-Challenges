import subprocess
import yaml

with open("challenge.yml", encoding="utf-8") as stream:
    config = yaml.safe_load(stream)

flag = config["flags"][0]
print("Testing with: " + flag)

try:
    output = subprocess.check_output("./challenge", input=bytes(flag, "utf-8"))
    print(output.decode())
except subprocess.CalledProcessError as e:
    print("Output:\n", e.output)
