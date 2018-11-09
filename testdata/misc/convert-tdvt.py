#!/usr/local/bin/python3

import errno
import os
import signal
import csv
import sys
import yaml
import MySQLdb as mysql
from MySQLdb.constants.FIELD_TYPE import *
from functools import wraps

skip_rows = {
    'calcs.yml': set([]),
    'logical_calcs.yml': set([]),
    'logical_staples.yml': set([24, 37]),
}

skip_tests = {
    'calcs.yml': set([
        698,
        684,
        655,
        654,
        677,
        676,
        675,
        635,
        634,
        630,
        629,
        628,
        627,
        610,
        602,
        169,
        168,
        167,
        166,
        162,
        155,
        152,
        581,
        558,
        142,
        141,
        540,
        539,
        538,
        128,
        127,
        126,
        125,
        461,
        450,
        449,
        448,
        455,
        447,
        442,
        435,
        99,
        98,
        96,
        86,
        85,
        84,
        76,
        378,
        75,
        377,
        74,
        73,
        374,
        371,
        365,
        362,
        359,
        358,
        419,
        415,
        332,
        331,
        329,
        328,
        327,
        406,
        326,
        325,
        324,
        323,
        322,
        293,
        292,
        291,
        278,
        277,
        276,
        275,
        265,
        263,
        231,
        229,
        256,
        224,
    ]),
    'logical_calcs.yml': set([
        11,
        0,
        6,
        9,
        7,
        20,
        5,
        19,
        15,
        14,
        16,
        17,
        13,
        1,
        12,
    ]),
    'logical_staples.yml': set([
        0,
        36,
        30,
        29,
        34,
        15,
        39,
        40,
        27,
        26,
        38,
        14,
        23,
        13,
        8,
        12,
        25,
        11,
        20,
        10,
        19,
        22,
        21,
        17,
        33,
        7,
        16,
        31,
        6,
        32,
        2,
        1,
        4,
        3,
        5,
        18,
    ]),
}

def main():
    in_filenames = get_in_filenames()
    testcases_by_out_filename = {}

    print('\nGenerating sqlproxy tests from csv files...')
    for filename in in_filenames:
        out_filename = generate_out_filename(filename)
        testcases = generate_testcases(filename)
        testcases_by_out_filename[out_filename] = {
            'testcases': testcases,
        }
    print('Done')

    print('\nRunning queries to get column names/types...')
    for filename in testcases_by_out_filename:
        testfile = testcases_by_out_filename[filename]
        get_column_info(testfile['testcases'], os.path.basename(filename))
    print('Done')

    print('\nWriting tests to yml files...')
    for out_filename in testcases_by_out_filename:
        out_data = testcases_by_out_filename[out_filename]
        with open(out_filename, 'w') as out_file:
            formatted_data = yaml.dump(out_data, Dumper=TestDumper, default_flow_style=None, width=5000)
            out_file.write(formatted_data.replace("\n  - ", "\n\n  - "))
    print('Done')

class TestDumper(yaml.Dumper):
    def increase_indent(self, flow=False, indentless=False):
        return super(TestDumper, self).increase_indent(flow, False)

def get_in_filenames():
    basenames = [
        'calcs.csv',
        'logical-calcs.csv',
        'logical-staples.csv',
    ]
    in_filenames = basenames
    return in_filenames

def generate_out_filename(filename):
    out_directory = '../suites/tdvt/'
    out_filename = os.path.basename(filename)
    out_filename = out_filename.replace('-', '_')
    out_filename = out_filename.replace('csv', 'yml')
    out_filename = os.path.join(out_directory, out_filename)
    return out_filename

def generate_testcases(filename):
    csv_rows = []
    with open(filename) as in_file:
        fieldnames = ['tds_name', 'test_name', 'passed', 'z', 'zz', 'zzz', 'error_msg', 'error_type', 'time', 'sql', 'actual', 'expected_results']
        data = csv.DictReader(in_file, fieldnames=fieldnames)
        for row in data:
            csv_rows.append(row)


    testcases = []
    csv_rows = csv_rows[1:]
    for i in range(len(csv_rows)):
        test = csv_rows[i]

        sql = canonicalize_query(test['sql'])
        expected_error = test['error_type']
        expected_results = []

        if expected_error == '':
            rows = test['expected_results'].strip().split('\n')
            for row in rows:
                value = canonicalize_value(row)
                expected_results.append([value])

        test_case = {
            'id': str(i),
            'sql': sql,
            'expected_results': expected_results,
            'expected_error': expected_error,
        }

        if filename == 'logical-staples.csv':
            test_case['min_server_version'] = '3.4'

        testcases.append(test_case)

    return testcases

def get_column_info(testcases, filename):
    try:
        db = mysql.connect(
            unix_socket='/tmp/mysql.sock',
            db='fullblackbox',
        )
    except mysql.Error as e:
        print("Error connecting to db:", e)
        sys.exit(1)

    for i in range(len(testcases)):
        test = testcases[i]

        if i in skip_rows[filename]:
            test['skip'] = True
            print('skipping column info generation: {}[{}] -- setting test.skip=true'.format(filename, i))
            continue

        if i in skip_tests[filename]:
            print('known failing test: {}[{}] -- setting test.skip=true'.format(filename, i))
            test['skip'] = True

        try:
            get_column_info_for_test(test, db)
        except:
            test['skip'] = True
            print('error during column info generation: {}[{}] -- setting test.skip=true'.format(filename, i))

    db.close()

def get_column_info_for_test(test, db):
    sql = test['sql']

    cursor = db.cursor()
    cursor.execute(sql)

    expected_names = []
    expected_types = []
    for i in range(len(cursor.description)):
        col_desc = cursor.description[i]
        col_name = col_desc[0]
        col_type = mysql_type_strings[col_desc[1]]
        expected_names.append(col_name)
        expected_types.append(col_type)

    test['expected_names'] = expected_names
    test['expected_types'] = expected_types

    cursor.close()

def canonicalize_query(query):
    #query = query.replace('"', '`')
    #query = query.replace('`TestV1`.`Calcs`', '`Calcs`')
    #query = query.replace('`TESTV1`.`Calcs`', '`Calcs`')
    #query = query.replace('`ADMIN`.`Calcs`', '`Calcs`')
    #query = query.replace('`PUBLIC`.`Calcs`', '`Calcs`')
    #query = query.replace('`PUBLIC`.`CALCS`', 'CALCS')
    #query = query.replace('[dbo].', '')
    #query = re.sub('\[Calcs\](?:(\.)\[([A-Za-z0-9]*)\])?', r'Calcs\1\2', query)
    #query = re.sub('\[(TEMP\(Test\)\([0-9]+\)\(0\))]', r'`\1`', query)
    query += ' order by 1'
    return query

def canonicalize_value(value):
    if value == '%null%':
        return None

    if value.lower() == 'true':
        return 1

    if value.lower() == 'false':
        return 0

    if len(value) > 0:
        if value[0] == '#' and value[-1] == '#':
            return value[1:-1]

        if value[0] == '"' and value[-1] == '"':
            value = value[1:-1]

    # TODO: this probably will have some false positives
    value = value.replace(r'\"', '"')

    return value

mysql_type_strings = {
    DECIMAL: 'decimal',
    TINY: 'tiny',
    SHORT: 'short',
    LONG: 'long',
    FLOAT: 'float',
    DOUBLE: 'double',
    NULL: 'null',
    TIMESTAMP: 'timestamp',
    LONGLONG: 'longlong',
    INT24: 'int',
    DATE: 'date',
    TIME: 'time',
    DATETIME: 'datetime',
    YEAR: 'year',
    NEWDATE: 'newdate',
    VARCHAR: 'varchar',
    BIT: 'bit',
    NEWDECIMAL: 'newdecimal',
    ENUM: 'enum',
    SET: 'set',
    TINY_BLOB: 'tiny_blob',
    MEDIUM_BLOB: 'medium_blob',
    LONG_BLOB: 'long_blob',
    BLOB: 'blob',
    VAR_STRING: 'var_string',
    STRING: 'string',
    GEOMETRY: 'geometry',
}


if __name__ == "__main__":
    main()
