import asyncio
from mcp.client.sse import sse_client
from mcp import ClientSession


async def main():
    async with sse_client('http://127.0.0.1:8006/sse') as streams:
        async with ClientSession(*streams) as session:
            await session.initialize()
            
            # 查看工具列表
            tools = await session.list_tools()
            print("Available tools:", tools)

            # 调用工具 get_page_content 测试GitHub链接
            pages = await session.call_tool(
                "get_page_content",
                {
                    "urls": [
                        "https://github.com/Henry-23/VideoChat"
                    ]
                }
            )
            print("get_page_content result:", pages.content[0].text)
            print("Result type:", type(pages.content[0].text))
            print("Result length:", len(pages.content[0].text))


if __name__ == "__main__":
    asyncio.run(main())
 