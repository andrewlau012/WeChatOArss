
import asyncio
from playwright.async_api import async_playwright, Browser, Page
from typing import Optional, Dict
import logging
from datetime import datetime

logger = logging.getLogger("wechat_service")

class WeChatService:
    _instance = None
    _browser: Optional[Browser] = None
    _context = None
    _page: Optional[Page] = None
    _cookies: Dict = {}
    
    def __new__(cls):
        if cls._instance is None:
            cls._instance = super(WeChatService, cls).__new__(cls)
        return cls._instance

    async def init_browser(self):
        if not self._browser:
            p = await async_playwright().start()
            self._browser = await p.chromium.launch(headless=True, args=['--no-sandbox', '--disable-setuid-sandbox'])
            self._context = await self._browser.new_context(
                user_agent="Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36"
            )
            self._page = await self._context.new_page()

    async def get_login_qrcode(self) -> str:
        """
        Navigate to WeRead login page and return QR Code image URL (base64 or http)
        For simplicity in this demo, we might return the page screenshot or element screenshot.
        """
        await self.init_browser()
        if not self._page:
            raise Exception("Browser init failed")

        try:
            await self._page.goto("https://weread.qq.com/web/login")
            # Wait for QR code element
            # WeRead login usually has a QR code container
            qr_selector = ".wr_login_img" 
            await self._page.wait_for_selector(qr_selector, timeout=10000)
            
            # Take a screenshot of the QR code
            # In a real app, you might want to extract the QR content or return the image directly
            # Here we return a data URL of the screenshot for frontend to display
            element = self._page.locator(qr_selector)
            screenshot_bytes = await element.screenshot()
            import base64
            base64_str = base64.b64encode(screenshot_bytes).decode('utf-8')
            return f"data:image/png;base64,{base64_str}"
        except Exception as e:
            logger.error(f"Failed to get QR code: {e}")
            raise

    async def check_login_status(self) -> Optional[Dict]:
        """
        Check if current page is logged in (url change or specific element)
        Returns cookies if logged in, None otherwise
        """
        if not self._page:
            return None
        
        try:
            # Check if URL changed to home or specific element exists
            # WeRead redirects to /web/shelf or similar after login
            if "login" not in self._page.url:
                cookies = await self._context.cookies()
                cookie_dict = {c['name']: c['value'] for c in cookies}
                self._cookies = cookie_dict
                return cookie_dict
            
            # Also check for user avatar or similar
            if await self._page.query_selector(".wr_avatar"):
                cookies = await self._context.cookies()
                cookie_dict = {c['name']: c['value'] for c in cookies}
                self._cookies = cookie_dict
                return cookie_dict
                
        except Exception as e:
            logger.error(f"Check login error: {e}")
            
        return None

    async def parse_article_link(self, url: str) -> Dict:
        """
        Parse a WeChat Article URL to extract metadata
        """
        # Ensure browser is ready (use a new page for parsing to avoid messing up login state)
        if not self._browser:
            await self.init_browser()
            
        page = await self._browser.new_page()
        try:
            await page.goto(url)
            await page.wait_for_load_state("networkidle")
            
            # Extract Data using JS evaluation
            # WeChat articles have valid JS vars: nickname, appmsg_type, item_show_type, msg_title, msg_desc, msg_cdn_url, etc.
            data = await page.evaluate("""() => {
                const biz = window.biz || ''
                const nickname = (document.querySelector('.profile_nickname') || {}).innerText || window.nickname || ''
                const title = (document.querySelector('.rich_media_title') || {}).innerText || window.msg_title || ''
                const digest = (document.querySelector('.rich_media_meta_text') || {}).innerText || window.msg_desc || ''
                const cover = window.msg_cdn_url || ''
                const date = (document.querySelector('#publish_time') || {}).innerText || ''
                
                // Try to extract __biz from URL if not in window
                const urlParams = new URLSearchParams(window.location.search);
                const bizFromUrl = urlParams.get('__biz');

                return {
                    biz: biz || bizFromUrl,
                    name: nickname,
                    title: title,
                    digest: digest,
                    cover: cover,
                    date: date
                }
            }""")
            
            if not data['biz']:
                # Fallback: try regex on HTML content
                pass
                
            return data
        except Exception as e:
            logger.error(f"Parse article error: {e}")
            raise
        finally:
            await page.close()

wechat_service = WeChatService()
