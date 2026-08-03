import os
from datetime import datetime
from faker import Faker
from psycopg2 import OperationalError
from sqlalchemy import create_engine, text

fake = Faker()

DATABASE_URL = os.getenv("DATABASE_URL", "postgresql://postgres:postgres@localhost:5432/my_go_db")
TOTAL_USERS = 20
BATCH_SIZE = 4

def generate_user_data():
    return {
            "email": fake.free_email(),
            "username": fake.user_name(),
            "password": fake.password(),
            "first_name": fake.first_name(),
            "last_name": fake.last_name(),
        }

def seed_users():
    print(f"🚀 Starting database connection...")
    engine = create_engine(DATABASE_URL, echo=False)

    # Query SQL Insert
    insert_query = text("""
        INSERT INTO users (email, username, password, first_name, last_name)
        VALUES (:email, :username, :password, :first_name, :last_name)
    """)

    start_time = datetime.now()
    total_inserted = 0

    try:
        with engine.begin() as connection:
            print(f"📦 Generate and insert {TOTAL_USERS} data user...")
            
            batch = []  # Generate all users at once
            for i in range(1, TOTAL_USERS + 1):
                batch.append(generate_user_data())  # Generate one user at a time and add to batch

                # Eksekusi batch jika sudah mencapai batas BATCH_SIZE
                if len(batch) == BATCH_SIZE or i == TOTAL_USERS:
                    connection.execute(insert_query, batch)
                    total_inserted += len(batch)
                    print(f"   [+] Executed batch: {total_inserted}/{TOTAL_USERS} users")
                    batch = []  # Reset batch container

        elapsed = datetime.now() - start_time
        print(f"\n✅ Seeding completed! Successfully inserted {total_inserted} users in {elapsed.total_seconds():.2f} seconds.")

    except OperationalError as e:
        print(f"\n❌ Database connection error: {e}")
    except Exception as e:
        print(f"\n❌ Seeding failed: {e}")

if __name__ == "__main__":
    seed_users()