import os
import random
from datetime import datetime

from faker import Faker
from sqlalchemy import create_engine, text

fake = Faker()

DATABASE_URL = os.getenv("DATABASE_URL", "postgresql://postgres:postgres@localhost:5432/my_db")
TOTAL_POSTS = int(os.getenv("POST_COUNT", "100000"))
BATCH_SIZE = int(os.getenv("BATCH_SIZE", "1000"))


def generate_post_data(user):
    return {
        "user_id": user["id"],
        "username": user["username"],
        "title": fake.sentence(nb_words=8)[:255],
        "content": fake.paragraph(nb_sentences=5),
        "tags": fake.words(nb=3, unique=True),
    }


def seed_posts():
    if TOTAL_POSTS < 1:
        raise ValueError("POST_COUNT must be greater than zero")
    if BATCH_SIZE < 1:
        raise ValueError("BATCH_SIZE must be greater than zero")

    print("Starting database connection...")
    engine = create_engine(DATABASE_URL, echo=False)
    users_query = text("SELECT id, username FROM users ORDER BY id")
    insert_query = text(
        """
        INSERT INTO posts (user_id, username, title, content, tags)
        VALUES (:user_id, :username, :title, :content, :tags)
        """
    )

    start_time = datetime.now()
    total_inserted = 0

    with engine.begin() as connection:
        users = [dict(user) for user in connection.execute(users_query).mappings()]
        if not users:
            raise RuntimeError("No users found. Run make seed-users first.")

        print(f"Generating and inserting {TOTAL_POSTS} posts...")
        batch = []
        for index in range(TOTAL_POSTS):
            batch.append(generate_post_data(random.choice(users)))
            if len(batch) == BATCH_SIZE or index == TOTAL_POSTS - 1:
                connection.execute(insert_query, batch)
                total_inserted += len(batch)
                print(f"   Inserted batch: {total_inserted}/{TOTAL_POSTS} posts")
                batch = []

    elapsed = datetime.now() - start_time
    print(
        f"Seeding completed: inserted {total_inserted} posts "
        f"in {elapsed.total_seconds():.2f} seconds."
    )


if __name__ == "__main__":
    seed_posts()