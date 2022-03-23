import asyncio
import sqlalchemy as sa

from aiomysql.sa import create_engine
from billing.license.db import License, DB_HOST, DB_USER, DB_PASS, DB

metadata = sa.MetaData()

cols = [c for c in list(License.__table__.c)]
for c in cols:
  c.table = None

License = sa.Table('licenses', metadata, *cols)

async def get_engine():
  engine = await create_engine(
    user=DB_USER,
    db=DB,
    host=DB_HOST,
    password=DB_PASS)
  return engine
