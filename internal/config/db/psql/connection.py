from sqlalchemy import create_engine
import os


def make_connection_string_from_env() -> str:
    pas = os.getenv("POSTGRES_PASSWORD")
    user = os.getenv("POSTGRES_USER")
    host = os.getenv("POSTGRES_HOST")
    port = os.getenv("POSTGRES_PORT")
    db = os.getenv("POSTGRES_DB")

    s = f"postgresql://{user}:{pas}@{host}:{port}/{db}"
    return s
    

dsn = make_connection_string_from_env()

engine = create_engine(dsn)

class DB:
    def __init__(self):
        self.engine = engine
    
    def get_engine(self):
        return self.engine
    def close(self):
        self.engine.dispose()