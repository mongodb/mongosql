// Dataset 8: Deeply nested arrays (20 layers of nesting)

const db = connect('127.0.0.1:27017/memtest');

function nestarray (depth) {
    const arr = [];
    let cur = arr;
    for (let i = 0; i < depth; i++) {
        cur[0] = i;
        let tmp = [];
        cur[1] = {'a': tmp};
        cur = tmp;
    }
    cur[0] = depth;
    cur[1] = {'a': depth+1};
    return arr;
}

const array = nestarray(10)
for (let i = 0; i < 1000; i++) {
    const doc = {};
    doc[i] = array;
    db.test.insert(doc);
}
