from datetime import timestamp
from pydantic import BaseModel
from sqlalchemy.dialects.postgresql import UUID

class Refresh_tokens(BaseModel):
    refresh_tokens_id: UUID
    user_id: UUID
    token: str
    expires_at: timestamp
    created_at: timestamp
    is_revoked: bool 