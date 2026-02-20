from datetime import timestamp
from pydantic import BaseModel
from sqlalchemy.dialects.postgresql import UUID


class UseSettings(BaseModel):
    user_settings_id : UUID
    user_id : UUID
    notifications_enabled: bool
    language: str
    timezone: str
    privacy_level: str
    updated_at: timestamp