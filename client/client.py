import time
import progressbar
import requests
import json
import hmac
import hashlib
import uuid
import os
from urllib3.exceptions import ReadTimeoutError

u = str(uuid.uuid4())

target = os.getenv("DAST_API_TARGET", "https://ginandjuice.shop/")
api_url = os.getenv("DAST_API_URL", "https://dast.prodsec-dev.glovoint.com")
secret = os.getenv("DAST_HMAC_SECRET", "")
application = os.getenv("DAST_TARGET_APP", "dast-api")
build_id = os.getenv("DAST_BUILD_ID", u)

api_status = requests.get(api_url+"/ping", timeout=2)
print("Api status: \n\n")
print(json.dumps(api_status.json(), indent=4) + "\n")

status_dict = api_status.json()
if status_dict["zap"] != "ok":
    print("Scanner down\n")
    exit()

scan_body = {
    "build_id": build_id,
    "source": "github",
    "target": target,
    "application": application
}

body = json.dumps(scan_body).encode()


h = hmac.new(bytearray.fromhex(secret), body, hashlib.sha256)
s = h.hexdigest()

headers = {"Signature": s}
try:
    create_scan = requests.post(api_url+"/scan", json=scan_body, timeout=60, headers=headers)
except ReadTimeoutError:
    print("Scan creation failed")
    exit()

scan_created_data = create_scan.json()
s = scan_created_data["status"]

if s != "started":
    print("Scan creation failed, status: " + s)
    exit()

print("Scan started with ID " + scan_created_data["scanID"] + "\n")
finished = False

with progressbar.ProgressBar(max_value=100) as bar:
    while not finished:
        b = {"ScanID": scan_body["build_id"]}
        h = hmac.new(bytearray.fromhex(secret), json.dumps(b).encode(), hashlib.sha256)
        s = h.hexdigest()
        headers = {"Signature": s}
        scan_status = requests.post(api_url+"/status", json=b, headers=headers)
        scan_status_dict = scan_status.json()
        finished = scan_status_dict["status"] in ["passed", "failed", "error"]
        time.sleep(2)

        if scan_status_dict["status"] == "running":
            try:
                progress = int(scan_status_dict["progress"])
            except ValueError:
                progress = 100
            bar.update(progress)

print(scan_status.json())
