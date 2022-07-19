import os
import mysql.connector
import json
from timeit import default_timer as timer

start = timer()

mydb = mysql.connector.connect(
  host=os.getenv('DB_URL'),
  user=os.getenv('DB_USER'),
  password=os.getenv('DB_PASSWORD'),
  database=os.getenv('DB_NAME')
)

transactions_file = open('jobs/batches/transactions.json')

transactions = json.load(transactions_file)

mycursor = mydb.cursor()
number_transactions = len(transactions)
number_processed = 0
for transaction in transactions:
    sql = f"UPDATE clients SET balance = balance {'+' if transaction['OperationType']=='IN' else '-'} {transaction['Amount']} WHERE code = {transaction['ClientID']}"

    mycursor.execute(sql)

    mydb.commit()

    number_processed += 1

end = timer()
elapsed = end - start

mycursor.close()
mydb.close()

print('::', (number_processed / number_transactions) * 100, '%', 'Completed in', str(elapsed) + 's')
  

