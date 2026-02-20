from datetime import timestamp
from pydantic import BaseModel
from sqlalchemy.dialects.postgresql import UUID


class User(BaseModel):
    user_id: UUID
    email: str
    phone: str
    password: str
    subscription_status: str
    subscription_expires: timestamp
    is_active: bool
    created_at: timestamp
    updated_at: timestamp