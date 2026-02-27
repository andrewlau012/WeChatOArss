
import logging
import httpx
from typing import Optional, Dict
from sqlmodel import Session, select
from app.models import Account, OfficialAccount, Article
from datetime import datetime
import re
import asyncio

logger = logging.getLogger(__name__)

class WeReadService:
    def __init__(self, session: Session):
        self.session = session
        self.base_url = "https://weread.qq.com"
    
    def get_active_account(self) -> Optional[Account]:
        """Get a rotating active account"""
        # Simple round-robin or least used logic
        # For now, just pick the first active one
        statement = select(Account).where(Account.status == "active").order_by(Account.last_active)
        account = self.session.exec(statement).first()
        if account:
            account.last_active = datetime.utcnow()
            account.daily_usage += 1
            self.session.add(account)
            self.session.commit()
            return account
        return None

    async def search_official_account(self, keyword: str) -> list[dict]:
        """Search for Official Accounts by keyword"""
        account = self.get_active_account()
        if not account:
            logger.warning("No active account available for search")
            return []
            
        # Use WeRead search API (This is a mock URL based on typical behavior, needs real API analysis)
        # Assuming: /web/search/global?keyword=...
        
        # Real implementation requires correct headers & cookies from the Account object
        # cookies = json.loads(account.cookies)
        
        # Mock Response for now to allow frontend dev
        return [
            {
                "biz_id": "MzI3Mzc2NDY4MA==", 
                "name": f"{keyword}_Demo",
                "head_img": "https://res.wx.qq.com/a/wx_fed/assets/res/NTI4MWUi5.ico"
            }
        ]

    async def add_by_link(self, link: str) -> Optional[OfficialAccount]:
        """Parse article link and add OA"""
        # Regex to extract __biz
        biz_match = re.search(r'__biz=([^&]+)', link)
        if not biz_match:
            return None
            
        biz_id = biz_match.group(1)
        
        # Check if exists
        existing = self.session.exec(select(OfficialAccount).where(OfficialAccount.biz_id == biz_id)).first()
        if existing:
            return existing
            
        # If new, we need to fetch metadata (name, etc) from the link page
        # async with httpx.AsyncClient() as client:
        #    resp = await client.get(link)
        #    # Parse HTML for nickname...
        
        new_oa = OfficialAccount(
            biz_id=biz_id,
            name="Parsed Account", # Placeholder
            status="normal"
        )
        self.session.add(new_oa)
        self.session.commit()
        self.session.refresh(new_oa)
        return new_oa

    async def fetch_articles(self, oa: OfficialAccount):
        """Fetch latest articles for a specific OA"""
        # Logic to fetch list from WeRead or MP
        pass
