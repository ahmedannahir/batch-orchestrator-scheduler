const fs = require('fs');
const mysql = require('mysql2/promise');

(async () => {
    var start = performance.now();

    var conn = await mysql.createConnection({
      host: process.env.DB_URL,
      user: process.env.DB_USER,
      password: process.env.DB_PASSWORD,
      database: process.env.DB_NAME,
    });
    
    
    let transactions = JSON.parse(fs.readFileSync('jobs/batches/transactions.json'));
    
    let number_processed = 0;
    let number_transactions = transactions.length;

    for (const transaction of transactions) {
        await conn.execute(`UPDATE clients SET balance = balance ${transaction.OperationType === 'IN' ? '+' : '-'} ${transaction.Amount} WHERE code = ${transaction.ClientID}`);
        number_processed++;
    }

    var end = performance.now();
    var elapsed = end - start;

    console.log("::", (number_processed / number_transactions) * 100, "% Completed in", elapsed/1000 + 's');    
})();