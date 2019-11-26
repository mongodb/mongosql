// Dataset 5: Many collections

const db = connect('127.0.0.1:27017/memtest');

// Make 1000 collections
let name = 'test';
for (let i = 0; i < 4000; i++) {
    // Populate the collection
    for (let j = 0; j < 60; j++){
        const doc = {};
        doc[j] = j;
        db.getCollection(name).insert(doc);
    }
    name = 'test'+i;
}
