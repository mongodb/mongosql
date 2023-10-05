char* translate(char *current_db, char *sql, char *catalog, int relax_schema_checking, int exclude_namespaces);
char* version();
char *get_namespaces(char *current_db, char *sql);
void delete_string(char *str);
