
import logging
from fastapi import FastAPI
from fastapi.middleware.cors import CORSMiddleware
from apscheduler.schedulers.asyncio import AsyncIOScheduler
from app.db import create_db_and_tables
from app.routers import auth, feeds, rss, system
from contextlib import asynccontextmanager

# Configure logging
logging.basicConfig(
    level=logging.INFO,
    format="%(asctime)s - %(name)s - %(levelname)s - %(message)s",
)
logger = logging.getLogger(__name__)

# Scheduler instance
scheduler = AsyncIOScheduler()

@asynccontextmanager
async def lifespan(app: FastAPI):
    # Startup
    logger.info("Starting up...")
    create_db_and_tables()
    scheduler.start()
    logger.info("Scheduler started.")
    
    # Add initial jobs here if needed
    
    yield
    
    # Shutdown
    logger.info("Shutting down...")
    scheduler.shutdown()

app = FastAPI(
    title="WeChatOA2Rss API",
    description="Backend API for WeChat Official Account RSS service",
    version="0.1.0",
    lifespan=lifespan,
    docs_url="/api/docs",
    redoc_url="/api/redoc",
    openapi_url="/api/openapi.json"
)

# CORS
origins = [
    "http://localhost",
    "http://localhost:8080",
    "http://localhost:5173",  # Vite dev server
    "*" # For development, restrict in production
]

app.add_middleware(
    CORSMiddleware,
    allow_origins=origins,
    allow_credentials=True,
    allow_methods=["*"],
    allow_headers=["*"],
)

# Include Routers
app.include_router(auth.router, prefix="/api/v1/auth", tags=["Auth"])
app.include_router(feeds.router, prefix="/api/v1/feeds", tags=["Feeds"])
app.include_router(rss.router, prefix="/api/v1/rss", tags=["RSS"])
app.include_router(system.router, prefix="/api/v1/system", tags=["System"])

@app.get("/")
async def root():
    return {"message": "WeChatOA2Rss Backend is running"}
