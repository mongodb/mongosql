// Dataset 8: Deeply nested arrays (30 layers of nesting), across 4 collections in 2 databases.

const db1 = connect('127.0.0.1:27017/memtest');
const db2 = connect('127.0.0.1:27017/memtest2');

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

function insert_in_collection(col, num) {
    for (let i = 0; i < num; i++) {
        const doc = {};
        doc[i] = array;
        col.insert(doc);
    }
}

const array = nestarray(30)
insert_in_collection(db1.test, 10)
insert_in_collection(db1.test2, 10)
insert_in_collection(db2.test, 10)
insert_in_collection(db2.test2, 10)
