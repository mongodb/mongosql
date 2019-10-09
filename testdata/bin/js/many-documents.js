// Dataset 6: Many (1,000,000) documents

const db = connect('127.0.0.1:27017/memtest');

for (let i = 0; i < 1000000; i++){
    const doc = {};
    doc[i] = i;
    db.test.insert(doc);
}
