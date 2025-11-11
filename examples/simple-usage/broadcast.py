#!/usr/bin/python
import time

import requests

id = "foo"
print(f"Sending req with id {id}")

for _ in range(10):
    response = requests.post(
        f"http://127.0.0.1:8080/channels/{id}/broadcast", {"foo": "bar"}
    )
    response.raise_for_status()
    print(response.json)
    time.sleep(2)
