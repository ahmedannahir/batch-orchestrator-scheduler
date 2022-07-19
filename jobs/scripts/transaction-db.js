var mysql = require('mysql2/promise');

(async () => {
  var start = performance.now()

  var conn = await mysql.createConnection({
    host: process.env.DB_URL,
    user: process.env.DB_USER,
    password: process.env.DB_PASSWORD,
    database: process.env.DB_NAME,
  });
  
  const [transactions, _] = await conn.execute(
    `SELECT * from transactions WHERE DATE(operation_date) > '${((yesterdayDate) => {
      return `${new Date(yesterdayDate
        .setDate(yesterdayDate.getDate() - 1))
        .toISOString()
        .slice(0, 19)
        .replace('T', ' ')}';`
    })(new Date)}`);

  let number_processed = 0;
  let number_transactions = transactions.length;

  for (const transaction of transactions) {
    await conn.execute(`UPDATE clients SET balance = balance ${transaction.operation_type === 'IN' ? '+' : '-'} ${transaction.amount} WHERE code = ${transaction.client_id}`);
    number_processed++;
  }

  var end = performance.now();
  var elapsed = end - start;

  console.log("::", (number_processed / number_transactions) * 100, "% Completed in", elapsed/1000 + 's');
})();