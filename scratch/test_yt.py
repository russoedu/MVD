import urllib.request
import re
import json

url = 'https://www.youtube.com/watch?v=rnlp_avexYQ'
req = urllib.request.Request(url, headers={
    'User-Agent': 'Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36',
    'Accept-Language': 'en-US,en;q=0.9'
})
html = urllib.request.urlopen(req).read().decode('utf-8', errors='ignore')

m = re.search(r'var ytInitialData = ({.*?});</script>', html)
if m:
    data = json.loads(m.group(1))

    # Recursive search for '-bsONE-kZwI'
    def find_key(obj, target, path=""):
        if isinstance(obj, dict):
            for k, v in obj.items():
                if target in str(v):
                    find_key(v, target, path + "." + k)
        elif isinstance(obj, list):
            for i, item in enumerate(obj):
                if target in str(item):
                    find_key(item, target, f"{path}[{i}]")
        else:
            if target in str(obj):
                print(f"FOUND at {path}: {obj}")

    find_key(data, "-bsONE-kZwI")
