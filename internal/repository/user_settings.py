from sqlalchemy import Column, String, Boolean, TIMESTAMP, ForeignKey
from sqlalchemy.dialects.postgresql import UUID
from sqlalchemy.ext.declarative import declarative_base
from sqlalchemy.sql import func
import uuid

Base = declarative_base()

class UserSettings(Base):
    __tablename__ = 'user_settings'
    
    user_settings_id = Column(UUID(as_uuid=True), primary_key=True, default=uuid.uuid4)
    user_id = Column(UUID(as_uuid=True), ForeignKey('users.id'), nullable=False, unique=True)
    
    notifications_enabled = Column(Boolean, nullable=False, default=True)
    language = Column(String(5), nullable=False, default='ru')
    timezone = Column(String(50), nullable=False, default='Europe/Moscow')
    privacy_level = Column(String(20), nullable=False, default='public')
    
    updated_at = Column(TIMESTAMP, server_default=func.now(), onupdate=func.now())