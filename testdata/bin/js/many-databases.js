// Dataset 4: Many databases

let db = connect('127.0.0.1:27017/memtest');

// Make 1000 dbs
let name = 'memtest';
for (let i = 0; i < 1000; i++) {
    db = db.getSiblingDB(name);
    // Populate the db with a collection
    for (let j = 0; j < 250; j++){
        const doc = {};
        doc[j] = j;
        db.test.insert(doc);
    }
    name = 'memtest' + i
}
