// Dataset 4: Many databases

let db = connect('127.0.0.1:27017/memtest');

// Make 1000 dbs
for (let i = 0; i < 1000; i++) {
    const name = 'memtest'+i;
    db = db.getSiblingDB(name);
    // Populate the db with a collection
    for (let j = 0; j < 1000; j++){
        const doc = {};
        doc[j] = j;
        db.test.insert(doc);
    }
}
