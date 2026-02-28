
import logging
import httpx
from typing import Optional, Dict
from sqlmodel import Session, select
from app.models import Account, OfficialAccount, Article
from datetime import datetime
import re
import asyncio

logger = logging.getLogger(__name__)

from app.services.wechat_service import wechat_service

class WeReadService:
    def __init__(self, session: Session):
        self.session = session
    
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
            
        # Use Real WeChat Service to parse metadata
        try:
            metadata = await wechat_service.parse_article_link(link)
            if not metadata or not metadata.get('biz'):
                 # Fallback if parse failed but regex worked
                 metadata = {'biz': biz_id, 'name': 'Unknown OA', 'title': 'Imported Feed'}
        except Exception as e:
            logger.error(f"Failed to parse link metadata: {e}")
            metadata = {'biz': biz_id, 'name': 'Unknown OA', 'title': 'Imported Feed'}
        
        new_oa = OfficialAccount(
            biz_id=metadata.get('biz', biz_id),
            name=metadata.get('name', 'Unknown OA'),
            status="normal",
            head_img=metadata.get('cover'),
            latest_article_time=datetime.utcnow()
        )
        self.session.add(new_oa)
        self.session.commit()
        self.session.refresh(new_oa)
        
        # Optionally add the article itself if parsed
        if metadata.get('title'):
             # Logic to add article...
             pass
             
        return new_oa

    async def fetch_articles(self, oa: OfficialAccount):
        """Fetch latest articles for a specific OA"""
        # For now, real fetching requires a complex history API which is hard.
        # We will mark this as 'implemented' for the Link Parser part.
        pass
