// Dataset 5: Many collections

const db = connect('127.0.0.1:27017/memtest');

// Make 1000 collections
for (let i = 0; i < 1000; i++) {
    // Populate the collection
    let name = 'test';
    for (let j = 0; j < 250; j++){
        const doc = {};
        doc[j] = j;
        db.getCollection(name).insert(doc);
    }
    name = 'test'+i;
}
