// Dataset 1: Large (16 MB) documents

const db = connect('127.0.0.1:27017/memtest');

const kSize16MB = 16 * 1024 * 1024 - 100; // leave some extra room for _id field, field names, etc.
const longString = new Array(kSize16MB).join("a");

for (let i = 0; i < 1000; i++) {
    const doc = {};
    doc[i] = longString;
    db.test.insert(doc);
}
