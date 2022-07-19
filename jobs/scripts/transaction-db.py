import os
import time
import mysql.connector
from datetime import date, timedelta
from timeit import default_timer as timer

start = timer()

db = mysql.connector.connect(
  host=os.getenv('DB_URL'),
  user=os.getenv('DB_USER'),
  password=os.getenv('DB_PASSWORD'),
  database=os.getenv('DB_NAME')
)

cursor = db.cursor()

yesterday = date.today() - timedelta(days=1)
cursor.execute(f"SELECT * FROM transactions WHERE DATE(operation_date) > {yesterday}")
transactions = cursor.fetchall()

cursor.execute(f"SELECT count(*) FROM transactions WHERE DATE(operation_date) > {yesterday}")
number_transactions = cursor.fetchone()[0]
number_processed = 0

for transaction in transactions:
    if transaction[3] == 'IN':
        sql = "UPDATE clients SET balance = balance + %s WHERE code = %s"
    if transaction[3] == 'OUT':
        sql = "UPDATE clients SET balance = balance - %s WHERE code = %s"

    cursor.execute(sql, (transaction[2], transaction[1]))

    db.commit()
    number_processed += 1

end = timer()
elapsed = end - start

cursor.close()
db.close()

print('::', (number_processed / number_transactions) * 100, '%', 'Completed in', str(elapsed) + 's')