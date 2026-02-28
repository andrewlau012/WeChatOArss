
from typing import Optional
from sqlmodel import Field, SQLModel
from datetime import datetime
import uuid

class Account(SQLModel, table=True):
    id: Optional[int] = Field(default=None, primary_key=True)
    vid: str = Field(index=True, unique=True) # WeRead User ID
    name: str
    avatar: Optional[str] = None
    cookies: str # JSON string of cookies
    skey: Optional[str] = None
    status: str = Field(default="active") # active, blocked, expired
    last_active: datetime = Field(default_factory=datetime.utcnow)
    last_check: Optional[datetime] = None
    daily_usage: int = Field(default=0)

class OfficialAccount(SQLModel, table=True):
    id: Optional[int] = Field(default=None, primary_key=True)
    biz_id: str = Field(index=True, unique=True) # __biz
    name: str
    description: Optional[str] = None
    head_img: Optional[str] = None
    latest_article_time: Optional[datetime] = None
    unread_count: int = Field(default=0)
    status: str = Field(default="normal") # normal, hidden
    added_at: datetime = Field(default_factory=datetime.utcnow)

class Article(SQLModel, table=True):
    id: Optional[int] = Field(default=None, primary_key=True)
    article_id: str = Field(index=True, unique=True) # Unique ID from URL or docid
    biz_id: str = Field(foreign_key="officialaccount.biz_id", index=True)
    title: str
    digest: Optional[str] = None # Summary
    content_html: Optional[str] = None # Full content
    url: str
    cover_img: Optional[str] = None
    pub_date: datetime
    author: Optional[str] = None
    fetched_at: datetime = Field(default_factory=datetime.utcnow)

class SystemConfig(SQLModel, table=True):
    key: str = Field(primary_key=True)
    value: str
