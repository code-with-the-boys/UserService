from datetime import timestamp
from pydantic import BaseModel
from sqlalchemy.dialects.postgresql import UUID


class UserProfiles(BaseModel):
    user_prodiles_id: UUID
    user_id: UUID
    name : str
    surname: str
    patronymic: str
    date_of_birth: timestamp
    gender: str
    height_cm: float
    weight_kg: float
    fitness_goal : str
    experience_level: str
    health_limitations: str
    created_at: timestamp
    updated_at: timestamp