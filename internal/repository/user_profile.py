from sqlalchemy import Column, String, Integer, DECIMAL, Date, Text, TIMESTAMP, ForeignKey
from sqlalchemy.dialects.postgresql import UUID
from sqlalchemy.ext.declarative import declarative_base
from sqlalchemy.sql import func
import uuid

Base = declarative_base()

class UserProfile(Base):
    __tablename__ = 'user_profiles'
    
    user_prodiles_id = Column(UUID(as_uuid=True), primary_key=True, default=uuid.uuid4)
    user_id = Column(UUID(as_uuid=True), ForeignKey('users.id'), nullable=False)
    
    name = Column(String(50), nullable=False)
    surname = Column(String(50), nullable=False)
    patronymic = Column(String(50), nullable=True)
    
    date_of_birth = Column(Date, nullable=True)
    gender = Column(String(10), nullable=True)
    height_cm = Column(DECIMAL(5,2), nullable=True)
    weight_kg = Column(DECIMAL(5,2), nullable=True)
    fitness_goal = Column(String(50), nullable=True)
    experience_level = Column(String(20), nullable=True)
    health_limitations = Column(Text, nullable=True)
    
    created_at = Column(TIMESTAMP, server_default=func.now())
    updated_at = Column(TIMESTAMP, onupdate=func.now())