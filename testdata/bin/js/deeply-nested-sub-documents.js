// Dataset 2: Deeply nested (100 levels of nesting) documents

const db = connect('127.0.0.1:27017/memtest');

const nestobj = (depth) => {
    const doc = {};
    let cur = doc;
    for (let i = 0; i < depth; i++) {
        cur[i] = {};
        cur = cur[i];
    }
    cur['a'] = 'foo';
    return doc;
};

const nestedDoc = nestobj(99);
for (let i = 0; i < 1000; i++) {
    const doc = {};
    doc[i] = nestedDoc;
    db.test.insert(doc);
}
