
from fastapi import APIRouter, HTTPException, Depends
from sqlmodel import Session
from app.db import get_session
from app.models import Account
from app.services.wechat_service import wechat_service
import uuid
import json
import logging

logger = logging.getLogger(__name__)
router = APIRouter()

@router.get("/qrcode")
async def get_login_qrcode():
    """Get WeRead login QR Code using Playwright"""
    try:
        # Get base64 QR code image
        qr_image = await wechat_service.get_login_qrcode()
        session_uuid = str(uuid.uuid4())
        
        return {
            "uuid": session_uuid,
            "qrcode_url": qr_image, # This is a data:image/png;base64,... string
            "expire_seconds": 300,
            "message": "Please scan the QR code with WeChat"
        }
    except Exception as e:
        logger.error(f"Failed to generate QR code: {e}")
        raise HTTPException(status_code=500, detail="Failed to generate QR code")

@router.get("/status")
async def check_login_status(uuid: str, session: Session = Depends(get_session)):
    """Check if the user has scanned and logged in"""
    # Check if cookies are available in the service
    cookies = await wechat_service.check_login_status()
    
    if cookies:
        # Save account to DB
        # We need to extract VID or Name from cookies or page
        # For now, use a placeholder or extract from specific cookie if known
        vid = cookies.get('wr_vid', 'unknown_vid') 
        name = cookies.get('wr_name', 'WeRead User')
        
        # Check if exists
        account = session.query(Account).filter(Account.vid == vid).first()
        if not account:
            account = Account(
                vid=vid,
                name=name,
                cookies=json.dumps(cookies),
                status="active"
            )
        else:
            account.cookies = json.dumps(cookies)
            account.status = "active"
            account.last_active = datetime.utcnow()
            
        session.add(account)
        session.commit()
        
        return {"status": "confirmed", "vid": vid}
    
    return {"status": "waiting"}

@router.get("/accounts")
async def list_accounts(session: Session = Depends(get_session)):
    """List all registered WeRead accounts"""
    accounts = session.query(Account).all()
    return accounts
