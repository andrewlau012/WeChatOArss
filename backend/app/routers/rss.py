
from fastapi import APIRouter, Response, Depends, HTTPException
from sqlmodel import Session, select
from app.db import get_session
from app.models import OfficialAccount, Article
from rfeed import Item, Feed, Guid
import datetime

router = APIRouter()

@router.get("/{biz_id}.xml")
async def get_rss_xml(biz_id: str, limit: int = 20, session: Session = Depends(get_session)):
    """Generate RSS 2.0 Feed"""
    
    # 1. Get OA Info
    oa = session.exec(select(OfficialAccount).where(OfficialAccount.biz_id == biz_id)).first()
    if not oa:
        # Fallback for "all" or invalid ID
        if biz_id == "all":
             oa = OfficialAccount(name="All Feeds", description="Aggregated feeds from WeChatOA2Rss")
        else:
             raise HTTPException(status_code=404, detail="Feed not found")

    # 2. Get Articles
    query = select(Article).order_by(Article.pub_date.desc()).limit(limit)
    if biz_id != "all":
        query = query.where(Article.biz_id == biz_id)
        
    articles = session.exec(query).all()
    
    # 3. Build RSS Items
    rss_items = []
    for article in articles:
        item = Item(
            title=article.title,
            link=article.url,
            description=article.content_html or article.digest or "",
            author=article.author,
            guid=Guid(article.article_id),
            pubDate=article.pub_date
        )
        rss_items.append(item)
        
    # 4. Build Feed
    feed = Feed(
        title=f"{oa.name} - WeChat RSS",
        link=f"http://localhost:8000/rss/{biz_id}.xml",
        description=oa.description or f"RSS feed for {oa.name}",
        language="zh-cn",
        lastBuildDate=datetime.datetime.now(),
        items=rss_items
    )
    
    return Response(content=feed.rss(), media_type="application/xml")

@router.get("/all.xml")
async def get_all_rss(limit: int = 20, session: Session = Depends(get_session)):
    return await get_rss_xml("all", limit, session)
