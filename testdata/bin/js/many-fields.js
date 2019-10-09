// Dataset 3: Documents with many fields

const db = connect('127.0.0.1:27017/memtest');

const largedoc = (num, docName) => {
    let doc = {};
    for (let i = 0; i < num; i++) {
        let name = docName+"_"+i;
        doc[name] = i;
    }
    return doc;
};

// Half the documents have 2000 fields, the BIC sampling limit.
for (let i = 0; i < 500; i++) {
    const doc2000 = largedoc(2000, i);
    db.test.insert(doc2000);
}

// Half the documents have 4000 fields.
// The BIC should only sample the first 2000 fields.
for (let i = 500; i < 1000; i++) {
    const doc4000 = largedoc(4000, i);
    db.test.insert(doc4000);
}
