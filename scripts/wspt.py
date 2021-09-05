import requests
import json
import os
import time
import json
import re
from urllib.parse import urlencode

requests.packages.urllib3.disable_warnings()
os.environ['no_proxy'] = '*'


def getsign():
    url = "https://hellodns.coding.net/p/sign/d/jsign/git/raw/master/sign"
    s = requests.session()
    r = s.get(url, verify=False)
    res_json = "&uuid=" + json.loads(r.text)["uuid"] + "&st=" + json.loads(r.text)["st"] + "&sign=" + json.loads(r.text)["sign"] + "&sv=" + json.loads(r.text)["sv"]
    return res_json

def getToken():
    headers = {
        'cookie': os.getenv('wskey'),
        'user-agent': 'okhttp/3.12.1;jdmall;android;version/;build/0;screen/1080x1920;os/5.1.1;network/wifi;',
        'content-type': 'application/x-www-form-urlencoded; charset=UTF-8',
    }
    
    url = 'https://api.m.jd.com/client.action?functionId=genToken&clientVersion=10.1.2&build=89743&client=android&d_brand=OPPO&d_model=PCRT00&osVersion=5.1.1&screen=1920*1080&partner=lc023&oaid=&eid=eidAe81b812187s36z8QOkxpRJWzMceSvZJ6Ges/EbXnbK3TBxc/JEcutXxuELIRMJDVeTNJFcAF/+tx1qw9GllLTdSnFeV3ic6909a697SbDL9zxEc4&sdkVersion=22&lang=zh_CN&aid=21e9fa9db1e4e15d&area=19_1601_3633_63257&networkType=wifi&wifiBssid=unknown&uts=0f31TVRjBSsqndu4%2FjgUPz6uymy50MQJw%2B3mGtYmx2hY8nVZkXFqGJ2D3wO8rvc%2BnAbe881zrDZjz3yU3z8vQgL8NZ7e39M3H2YpLER13q%2B3VUzHQXXLg4BMmeH%2B1W0%2BxQY%2FL%2FR4Y58JMW9A9F9yD2BtQPynkeKYtBsYDCkOn35Tv9ci57mPbqxYWU0TDVJ8t7JBXRhLckTorzxtEAVucA%3D%3D&uemps=0-0&harmonyOs=0' + getsign()
    
    body = 'body=%7B%22action%22%3A%22to%22%2C%22to%22%3A%22https%253A%252F%252Fplogin.m.jd.com%252Fcgi-bin%252Fm%252Fthirdapp_auth_page%253Ftoken%253DAAEAIEijIw6wxF2s3bNKF0bmGsI8xfw6hkQT6Ui2QVP7z1Xg%2526client_type%253Dandroid%2526appid%253D879%2526appup_type%253D1%22%7D&'
    
    res = requests.post(url, data=body, headers=headers, verify=False)
    res_json = json.loads(res.text)
    totokenKey = res_json['tokenKey']
    url = res_json.get('url')
    params = {
        'tokenKey': totokenKey,
        'to': 'https://plogin.m.jd.com/jd-mlogin/static/html/appjmp_blank.html'
    }
    res = requests.get(url=url, headers=headers, params=params, verify=False, allow_redirects=False)

    res_set = res.cookies.get_dict()
    pt_key = 'pt_key=' + res_set['pt_key']
    pt_pin = 'pt_pin=' + res_set['pt_pin']
    ck = str(pt_key) + ';' + str(pt_pin) + ';'
    print(ck)


if __name__ == '__main__':
    getToken()
