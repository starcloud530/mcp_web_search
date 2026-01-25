# 读取config.yaml 

import yaml 
import os 

# 全局配置变量
_config = None

def load_config(config_file=None):
    """
    加载配置文件并设置全局配置
    如果不提供config_file，默认使用项目根目录的统一配置文件
    """
    global _config
    
    # 如果没有提供配置文件路径，使用项目根目录的统一配置文件
    if config_file is None:
        # 获取当前文件路径
        current_file_path = os.path.abspath(__file__)
        # 获取项目根目录（上上级目录）
        project_root = os.path.dirname(os.path.dirname(os.path.dirname(current_file_path)))
        config_file = os.path.join(project_root, "config.yaml")
    
    with open(config_file, 'r') as f:
        _config = yaml.load(f, Loader=yaml.FullLoader)
    return _config

def get_config():
    """
    获取全局配置
    """
    if _config is None:
        # 如果配置未加载，自动加载统一配置文件
        load_config()
    return _config

def get_source_config(source_name):
    """
    获取特定源的配置
    """
    config = get_config()
    return config.get("links_search", {}).get("sources", {}).get(source_name, {})

def is_source_enabled(source_name):
    """
    检查特定源是否启用
    """
    source_config = get_source_config(source_name)
    return source_config.get("enabled", False)

def get_server_config():
    """
    获取服务器配置
    """
    config = get_config()
    return config.get("links_search", {})



