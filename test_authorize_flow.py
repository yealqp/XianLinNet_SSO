#!/usr/bin/env python3
"""
OAuth2 授权流程测试脚本
测试完整的授权码流程
"""

import requests
import webbrowser
from urllib.parse import urlencode, parse_qs, urlparse
from http.server import HTTPServer, BaseHTTPRequestHandler
import threading

# 配置
BASE_URL = "http://localhost:8080"  # 后端服务器地址
CLIENT_ID = "d0ad0afc-2123-48ad-bc08-dc06a95a6ed9"  # 从管理后台获取
CLIENT_SECRET = "vwyZGBZqo1e6vU0BHeikEk7-yY3TU6QyKnyGktnhhfs="  # 从管理后台获取
REDIRECT_URI = "http://localhost:8888/callback"  # 回调地址
SCOPE = "openid profile email"

# 用于接收授权码的简单 HTTP 服务器
authorization_code = None
server_running = True

class CallbackHandler(BaseHTTPRequestHandler):
    def do_GET(self):
        global authorization_code, server_running
        
        # 解析 URL
        parsed = urlparse(self.path)
        params = parse_qs(parsed.query)
        
        if 'code' in params:
            authorization_code = params['code'][0]
            
            # 返回成功页面
            self.send_response(200)
            self.send_header('Content-type', 'text/html')
            self.end_headers()
            self.wfile.write(b"""
                <html>
                <head><title>Authorization Successful</title></head>
                <body>
                    <h1>Authorization Successful!</h1>
                    <p>You can close this window now.</p>
                    <script>setTimeout(() => window.close(), 2000);</script>
                </body>
                </html>
            """)
            
            # 停止服务器
            server_running = False
        elif 'error' in params:
            error = params['error'][0]
            error_desc = params.get('error_description', ['Unknown error'])[0]
            
            self.send_response(200)
            self.send_header('Content-type', 'text/html')
            self.end_headers()
            self.wfile.write(f"""
                <html>
                <head><title>Authorization Failed</title></head>
                <body>
                    <h1>Authorization Failed</h1>
                    <p>Error: {error}</p>
                    <p>Description: {error_desc}</p>
                </body>
                </html>
            """.encode())
            
            server_running = False
    
    def log_message(self, format, *args):
        # 禁用日志输出
        pass

def start_callback_server():
    """启动回调服务器"""
    server = HTTPServer(('localhost', 8888), CallbackHandler)
    
    while server_running:
        server.handle_request()
    
    server.server_close()

def test_authorization_flow():
    """测试完整的 OAuth2 授权码流程"""
    
    print("=" * 60)
    print("OAuth2 授权流程测试")
    print("=" * 60)
    
    # 步骤 1: 启动回调服务器
    print("\n[1] 启动回调服务器...")
    server_thread = threading.Thread(target=start_callback_server, daemon=True)
    server_thread.start()
    
    # 等待服务器启动
    import time
    time.sleep(0.5)
    
    print("✓ 回调服务器已启动在 http://localhost:8888")
    
    # 步骤 2: 构建授权 URL
    print("\n[2] 构建授权 URL...")
    auth_params = {
        'client_id': CLIENT_ID,
        'response_type': 'code',
        'redirect_uri': REDIRECT_URI,
        'scope': SCOPE,
        'state': 'random_state_string'
    }
    
    # 使用前端地址，前端会处理授权页面
    auth_url = f"http://localhost:5173/authorize?{urlencode(auth_params)}"
    print(f"授权 URL: {auth_url}")
    
    # 步骤 3: 打开浏览器进行授权
    print("\n[3] 打开浏览器进行授权...")
    print("请在浏览器中登录并授权应用")
    print("(如果您还没有账号，请先注册)")
    
    webbrowser.open(auth_url)
    
    # 等待授权码
    print("\n等待授权...")
    while authorization_code is None and server_running:
        pass
    
    if authorization_code is None:
        print("✗ 授权失败或被取消")
        return
    
    print(f"✓ 收到授权码: {authorization_code[:20]}...")
    
    # 步骤 4: 使用授权码交换 access token
    print("\n[4] 使用授权码交换 access token...")
    token_url = f"{BASE_URL}/api/oauth/token"
    token_data = {
        'grant_type': 'authorization_code',
        'code': authorization_code,
        'redirect_uri': REDIRECT_URI,
        'client_id': CLIENT_ID,
        'client_secret': CLIENT_SECRET
    }
    
    try:
        response = requests.post(token_url, data=token_data)
        response.raise_for_status()
        
        token_response = response.json()
        
        if 'access_token' in token_response:
            access_token = token_response['access_token']
            refresh_token = token_response.get('refresh_token')
            expires_in = token_response.get('expires_in')
            
            print(f"✓ 获取 access token 成功")
            print(f"  Access Token: {access_token}")
            if refresh_token:
                print(f"  Refresh Token: {refresh_token}")
            if expires_in:
                print(f"  过期时间: {expires_in} 秒")
            
            # 调试：打印完整 token（仅用于测试）
            print(f"\n  [调试] 完整 Access Token: {access_token}")
            
            # 步骤 5: 使用 access token 获取用户信息
            print("\n[5] 使用 access token 获取用户信息...")
            userinfo_url = f"{BASE_URL}/api/userinfo"
            headers = {'Authorization': f'Bearer {access_token}'}
            
            print(f"  [调试] 请求 URL: {userinfo_url}")
            print(f"  [调试] Authorization Header: Bearer {access_token}")
            
            userinfo_response = requests.get(userinfo_url, headers=headers)
            
            print(f"  [调试] 响应状态码: {userinfo_response.status_code}")
            print(f"  [调试] 响应内容: {userinfo_response.text}")
            
            userinfo_response.raise_for_status()
            
            userinfo = userinfo_response.json()
            print("✓ 获取用户信息成功:")
            print(f"  {json.dumps(userinfo, indent=2, ensure_ascii=False)}")
            
            # 步骤 6: 测试 token 刷新（如果有 refresh token）
            if refresh_token:
                print("\n[6] 测试 token 刷新...")
                refresh_data = {
                    'grant_type': 'refresh_token',
                    'refresh_token': refresh_token,
                    'client_id': CLIENT_ID,
                    'client_secret': CLIENT_SECRET
                }
                
                refresh_response = requests.post(token_url, data=refresh_data)
                refresh_response.raise_for_status()
                
                new_token_response = refresh_response.json()
                if 'access_token' in new_token_response:
                    print(f"✓ Token 刷新成功")
                    print(f"  新 Access Token: {new_token_response['access_token']}")
                else:
                    print("✗ Token 刷新失败")
            
            print("\n" + "=" * 60)
            print("✓ OAuth2 授权流程测试完成！")
            print("=" * 60)
            
        else:
            print(f"✗ 获取 token 失败: {token_response}")
    
    except requests.exceptions.RequestException as e:
        print(f"✗ 请求失败: {e}")
        if hasattr(e.response, 'text'):
            print(f"  响应: {e.response.text}")

if __name__ == '__main__':
    import json
    
    print("\n请确保:")
    print("1. OAuth 服务器正在运行 (http://localhost:8080)")
    print("2. 您已经在管理后台创建了应用并获取了 client_id 和 client_secret")
    print("3. 应用的重定向 URI 包含: http://localhost:8888/callback")
    print()
    
    if not CLIENT_ID or CLIENT_ID == "your_client_id":
        print("\n✗ 请先配置 CLIENT_ID 和 CLIENT_SECRET")
        print("  可以使用内置应用: admin/app-built-in")
        print("  Client ID 和 Secret 可以在管理后台的应用管理中查看")
        exit(1)
    
    test_authorization_flow()
