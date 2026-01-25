
from api.links_api import app
from utils.config import load_config, get_server_config
if __name__ == "__main__":
    # 加载统一配置文件
    load_config()
    # 获取服务器配置
    server_config = get_server_config()
    import uvicorn
    uvicorn.run(app, host=server_config['host'], port=server_config['port'])
