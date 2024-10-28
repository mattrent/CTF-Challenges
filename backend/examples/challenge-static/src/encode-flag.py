import yaml

with open("../challenge.yml", encoding="utf-8") as stream:
    config = yaml.safe_load(stream)

flag = config["flags"][0]
result = []

for i, v in enumerate(flag):
    result.append(ord(v) - i)

print(result)
