
from fastapi import APIRouter, Depends
from sqlmodel import Session, select
from app.db import get_session
from app.models import SystemConfig
import json

router = APIRouter()

DEFAULT_CONFIG = {
    "cron_times": ["08:20", "13:00"],
    "rss_max_items": 20,
    "retention_days": 60
}

@router.get("/status")
async def get_system_status():
    return {"status": "ok"}

@router.get("/config")
async def get_config(session: Session = Depends(get_session)):
    """Get system configuration"""
    config = {}
    
    # Fetch from DB
    db_configs = session.exec(select(SystemConfig)).all()
    db_config_map = {item.key: item.value for item in db_configs}
    
    # Merge with defaults
    for key, default_val in DEFAULT_CONFIG.items():
        if key in db_config_map:
            try:
                # Attempt to parse JSON if it's complex, otherwise string
                # Since everything is stored as string in `value`, we need to handle types
                val_str = db_config_map[key]
                if key == "cron_times":
                    config[key] = json.loads(val_str)
                elif key in ["rss_max_items", "retention_days"]:
                    config[key] = int(val_str)
                else:
                    config[key] = val_str
            except:
                config[key] = default_val
        else:
            config[key] = default_val
            
    return config

@router.post("/config")
async def update_config(payload: dict, session: Session = Depends(get_session)):
    """Update system configuration"""
    for key, value in payload.items():
        if key not in DEFAULT_CONFIG:
            continue
            
        # Serialize complex types
        if isinstance(value, (list, dict)):
            val_str = json.dumps(value)
        else:
            val_str = str(value)
            
        # Update or Insert
        config_item = session.exec(select(SystemConfig).where(SystemConfig.key == key)).first()
        if config_item:
            config_item.value = val_str
        else:
            config_item = SystemConfig(key=key, value=val_str)
        session.add(config_item)
        
    session.commit()
    return {"status": "updated", "config": payload}
