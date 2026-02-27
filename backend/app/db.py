
from sqlmodel import SQLModel, create_engine, Session
import os

# Database configuration
DATABASE_URL = os.environ.get("DATABASE_URL", "sqlite:///./data/db.sqlite")

# Ensure data directory exists
os.makedirs(os.path.dirname(DATABASE_URL.replace("sqlite:///", "")), exist_ok=True)

engine = create_engine(DATABASE_URL, echo=False, connect_args={"check_same_thread": False})

def create_db_and_tables():
    SQLModel.metadata.create_all(engine)

def get_session():
    with Session(engine) as session:
        yield session
