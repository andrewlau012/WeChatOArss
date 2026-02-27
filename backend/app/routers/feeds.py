
from fastapi import APIRouter, HTTPException, Depends
from sqlmodel import Session, select
from typing import List
from app.db import get_session
from app.models import OfficialAccount, Article
from app.services.crawler import WeReadService

router = APIRouter()

@router.get("/list", response_model=List[OfficialAccount])
async def get_feed_list(session: Session = Depends(get_session)):
    """Get all monitored Official Accounts"""
    feeds = session.exec(select(OfficialAccount).order_by(OfficialAccount.latest_article_time.desc())).all()
    return feeds

@router.post("/add")
async def add_feed(payload: dict, session: Session = Depends(get_session)):
    """Add new feed by search or link"""
    crawler = WeReadService(session)
    
    if payload.get("type") == "link_parse":
        link = payload.get("value")
        if not link:
             raise HTTPException(status_code=400, detail="Link required")
        oa = await crawler.add_by_link(link)
        if not oa:
            raise HTTPException(status_code=400, detail="Invalid link or already exists")
        return oa
        
    elif payload.get("type") == "search_select":
        # Implementation for selecting from search results
        pass
    
    return {"status": "error", "message": "Not implemented"}

@router.get("/search")
async def search_feeds(keyword: str, session: Session = Depends(get_session)):
    crawler = WeReadService(session)
    results = await crawler.search_official_account(keyword)
    return results

@router.post("/refresh")
async def refresh_feeds(target_id: str = "all"):
    # Trigger async job
    return {"task_id": "job_started"}
