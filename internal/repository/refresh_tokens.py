from sqlalchemy import Column, String, Boolean, TIMESTAMP, ForeignKey, Index
from sqlalchemy.dialects.postgresql import UUID
from sqlalchemy.ext.declarative import declarative_base
from sqlalchemy.sql import func
import uuid

Base = declarative_base()

class RefreshToken(Base):
    __tablename__ = 'refresh_tokens'
    
    refresh_tokens_id = Column(UUID(as_uuid=True), primary_key=True, default=uuid.uuid4)
    user_id = Column(UUID(as_uuid=True), ForeignKey('users.id', ondelete='CASCADE'), nullable=False)
    
    token_hash = Column(String(255), nullable=False, unique=True)
    expires_at = Column(TIMESTAMP, nullable=False)
    created_at = Column(TIMESTAMP, server_default=func.now(), nullable=False)
    is_revoked = Column(Boolean, nullable=False, default=False)
    
    # Индексы для оптимизации запросов
    __table_args__ = (
        Index('idx_refresh_tokens_user_id', user_id),
        Index('idx_refresh_tokens_expires_at', expires_at),
        Index('idx_refresh_tokens_is_revoked', is_revoked),
    )
    
    def __repr__(self):
        return f"RefreshToken(id={self.id}, user_id={self.user_id}, token_hash={self.token_hash}, expires_at={self.expires_at}, created_at={self.created_at}, is_revoked={self.is_revoked})"