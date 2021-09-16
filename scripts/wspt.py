import json
import sys
import requests
import re
import time
requests.packages.urllib3.disable_warnings()
ws=sys.argv[1]
def getToken(wskey):
    headers = {
        'cookie': wskey,
        'User-Agent': ua,
        'content-type': 'application/x-www-form-urlencoded; charset=UTF-8',
        'charset': 'UTF-8',
        'accept-encoding': 'br,gzip,deflate'
    }
    params = {
        'functionId': 'genToken',
        'clientVersion': '10.1.2',
        'client': 'android',
        'uuid': uuid,
        'st': st,
        'sign': sign,
        'sv': sv
    }
    url = 'https://api.m.jd.com/client.action'
    data = 'body=%7B%22action%22%3A%22to%22%2C%22to%22%3A%22https%253A%252F%252Fplogin.m.jd.com%252Fcgi-bin%252Fm%252Fthirdapp_auth_page%253Ftoken%253DAAEAIEijIw6wxF2s3bNKF0bmGsI8xfw6hkQT6Ui2QVP7z1Xg%2526client_type%253Dandroid%2526appid%253D879%2526appup_type%253D1%22%7D&'
    try:
        res = requests.post(url=url, params=params, headers=headers, data=data, verify=False, timeout=10)
        res_json = json.loads(res.text)
        # logger.info(res_json)
        tokenKey = res_json['tokenKey']
        # logger.info("Token:", tokenKey)
    except:
        try:
            res = requests.post(url=url, params=params, headers=headers, data=data, verify=False, timeout=20)
            res_json = json.loads(res.text)
            # logger.info(res_json)
            tokenKey = res_json['tokenKey']
            # logger.info("Token:", tokenKey)
            return appjmp(wskey, tokenKey)
        except:
            logger.info("WSKEY转换接口出错, 请稍后尝试, 脚本退出")
            sys.exit(1)
    else:
        return appjmp(wskey, tokenKey)


# 返回值 bool jd_ck
def appjmp(wskey, tokenKey):
    headers = {
        'User-Agent': ua,
        'accept': 'text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3',
    }
    params = {
        'tokenKey': tokenKey,
        'to': 'https://plogin.m.jd.com/cgi-bin/m/thirdapp_auth_page?token=AAEAIEijIw6wxF2s3bNKF0bmGsI8xfw6hkQT6Ui2QVP7z1Xg',
        'client_type': 'android',
        'appid': 879,
        'appup_type': 1,
    }
    url = 'https://un.m.jd.com/cgi-bin/app/appjmp'
    try:
        res = requests.get(url=url, headers=headers, params=params, verify=False, allow_redirects=False, timeout=20)
        res_set = res.cookies.get_dict()
        pt_key = 'pt_key=' + res_set['pt_key']
        pt_pin = 'pt_pin=' + res_set['pt_pin']
        jd_ck = str(pt_key) + ';' + str(pt_pin) + ';'
        wskey = wskey.split(";")[0]
        if 'fake' in pt_key:
            logger.info(str(wskey) + ";wskey状态失效\n")
            return False, jd_ck
        else:
            logger.info(str(wskey) + ";wskey状态正常\n")
            print(jd_ck)
            return True, jd_ck
    except:
        logger.info("接口转换失败, 默认wskey失效\n")
        wskey = "pt_" + str(wskey.split(";")[0])
        return False, wskey


# 返回值 svv, stt, suid, jign
def get_sign():
    url = 'https://hellodns.coding.net/p/sign/d/jsign/git/raw/master/sign'
    for i in range(3):
        try:
            res = requests.get(url=url, verify=False, timeout=20)
        except requests.exceptions.ConnectTimeout:
            logger.info("\n获取Sign超时, 正在重试!" + str(i))
            time.sleep(1)
        except requests.exceptions.ReadTimeout:
            logger.info("\n获取Sign超时, 正在重试!" + str(i))
            time.sleep(1)
        except Exception as err:
            logger.info(str(err) + "\n未知错误, 退出脚本!")
            sys.exit(1)
        else:
            try:
                sign_list = json.loads(res.text)
            except:
                logger.info("Sign Json错误")
                sys.exit(1)
            else:
                svv = sign_list['sv']
                stt = sign_list['st']
                suid = sign_list['uuid']
                jign = sign_list['sign']
                return svv, stt, suid, jign


def cloud_info():
    url = 'https://hellodns.coding.net/p/sign/d/jsign/git/raw/master/check_api'
    for i in range(3):
        try:
            res = requests.get(url=url, verify=False, timeout=20).text
        except requests.exceptions.ConnectTimeout:
            logger.info("\n获取云端参数超时, 正在重试!" + str(i))
        except requests.exceptions.ReadTimeout:
            logger.info("\n获取云端参数超时, 正在重试!" + str(i))
        except Exception as err:
            logger.info(str(err) + "\n未知错误, 退出脚本!")
            sys.exit(1)
        else:
            try:
                c_info = json.loads(res)
            except:
                logger.info("云端参数解析失败")
                sys.exit(1)
            else:
                return c_info
                print(c_info)

def checkwskey(wskey):
    flag = "wskey=" in wskey
    flag1 = "pin=" in wskey
    if flag == True & flag1 == True:
       return True
    else :
       return False

if __name__ == '__main__':
    cloud_arg = cloud_info()
    ua = cloud_arg['User-Agent']
    getToken(ws)

