use crate::ast::{
    self,
    definitions::{
        Datasource, Expression, ExtendedUnwindOption, ExtendedUnwindSource, UnwindOption,
        UnwindPathPart, UnwindPathPartOption, UnwindSource,
    },
    rewrites::{Error, Pass, Result},
    visitor::Visitor,
    SubpathExpr,
};

pub struct ExtendedUnwindRewritePass;

impl Pass for ExtendedUnwindRewritePass {
    fn apply(&self, query: ast::Query) -> Result<ast::Query> {
        let mut visitor = ExtendedUnwindRewriteVisitor::default();
        let res = query.walk(&mut visitor);
        if let Some(error) = visitor.error {
            return Err(error);
        }
        Ok(res)
    }
}

/// The visitor that performs the rewrites for the `ExtendedUnwindRewriteRewritePass`.
#[derive(Default)]
struct ExtendedUnwindRewriteVisitor {
    error: Option<Error>,
}

impl Visitor for ExtendedUnwindRewriteVisitor {
    fn visit_datasource(&mut self, data_source: ast::Datasource) -> ast::Datasource {
        match data_source {
            ast::Datasource::ExtendedUnwind(ExtendedUnwindSource {
                datasource: source,
                options,
            }) => {
                let source = Box::new(self.visit_datasource(*source));
                let (mut paths, mut global_index, mut global_outer) = (None, None, false);
                let (mut path_count, mut index_count, mut outer_count) = (0, 0, 0);
                for option in options {
                    match option {
                        ExtendedUnwindOption::Paths(path) => {
                            paths = Some(path);
                            path_count += 1;
                        }
                        ExtendedUnwindOption::Index(i) => {
                            global_index = Some(i);
                            index_count += 1;
                        }
                        ExtendedUnwindOption::Outer(o) => {
                            global_outer = o;
                            outer_count += 1;
                        }
                    }
                }
                macro_rules! check_count {
                    ($count:ident, $name:literal) => {
                        if $count > 1 {
                            self.error = Some(Error::DuplicateOptionInUnwind($name));
                            return ast::Datasource::ExtendedUnwind(ExtendedUnwindSource {
                                datasource: source,
                                options: Vec::new(),
                            });
                        }
                    };
                }
                check_count!(path_count, "PATHS");
                check_count!(index_count, "INDEX");
                check_count!(outer_count, "OUTER");
                if paths.is_none() {
                    self.error = Some(Error::UnwindSourceWithoutPath);
                    return ast::Datasource::ExtendedUnwind(ExtendedUnwindSource {
                        datasource: source,
                        options: Vec::new(),
                    });
                }
                let mut paths = paths.unwrap();
                // if there is only one path and there are no local unwind options, this is an old
                // style UnwindSource that we can immediately convert to a normal UnwindSource.
                // In particular, this keeps us from prefixing the INDEX field with the path.
                if paths.len() == 1 && paths[0].iter().all(|p| p.options.is_empty()) {
                    create_simple_unwind_datasource(
                        source,
                        paths.remove(0),
                        global_index,
                        global_outer,
                    )
                } else {
                    create_unwind_datasource(*source, paths, global_index, global_outer)
                }
            }
            _ => data_source.walk(self),
        }
    }
}

// This converts a vector of strings representing a path into a nested SubpathExpr.
// Each element can be thought of as separated by `.` in the path. By representing it as a
// Vec, we can not worry about parts that contain `.` in them.
fn path_vec_to_path(mut path: Vec<String>) -> Expression {
    if path.len() == 1 {
        return Expression::Identifier(path.remove(0));
    }
    let mut ret = Expression::Identifier(path.remove(0));
    for p in path.into_iter() {
        ret = Expression::Subpath(SubpathExpr {
            expr: Box::new(ret),
            subpath: p.to_string(),
        });
    }
    ret
}

// Get options takes the ParthPartOptions and the path and returns a vector of UnwindOptions for
// normal UnwindSources.
fn get_options(
    options: Vec<UnwindPathPartOption>,
    path: Vec<String>,
    index_prefix: &str,
    global_index: Option<&String>,
    global_outer: bool,
) -> Vec<UnwindOption> {
    let mut ret = Vec::new();
    let mut found_index = false;
    let mut found_outer = false;
    for option in options.into_iter() {
        match option {
            UnwindPathPartOption::Index(i) => {
                ret.push(UnwindOption::Index(i));
                found_index = true;
            }
            UnwindPathPartOption::Outer(o) => {
                ret.push(UnwindOption::Outer(o));
                found_outer = true;
            }
        }
    }

    if !found_index && global_index.is_some() {
        ret.push(UnwindOption::Index(format!(
            "{}_{}",
            index_prefix,
            global_index.unwrap()
        )));
    }
    if !found_outer && global_outer {
        // there is no need to push Outer(false)
        ret.push(UnwindOption::Outer(true));
    }
    ret.push(UnwindOption::Path(path_vec_to_path(path)));
    ret
}

fn create_simple_unwind_datasource(
    source: Box<Datasource>,
    path: Vec<UnwindPathPart>,
    global_index: Option<String>,
    global_outer: bool,
) -> Datasource {
    let mut options = vec![UnwindOption::Path(path_vec_to_path(
        path.iter().map(|p| p.field.clone()).collect(),
    ))];
    if let Some(global_index) = global_index {
        options.push(UnwindOption::Index(global_index));
    }
    if global_outer {
        options.push(UnwindOption::Outer(true));
    }
    ast::Datasource::Unwind(ast::UnwindSource {
        datasource: source,
        options,
    })
}

// This simply loops over every path and calls create_unwind_datasource_for_path for each path
// to simplify the iteration.
fn create_unwind_datasource(
    source: Datasource,
    paths: Vec<Vec<UnwindPathPart>>,
    global_index: Option<String>,
    global_outer: bool,
) -> Datasource {
    let mut ret = source;
    for path in paths {
        ret = create_unwind_datasource_for_path(ret, path, global_index.clone(), global_outer);
    }
    ret
}

fn create_unwind_datasource_for_path(
    source: Datasource,
    path: Vec<UnwindPathPart>,
    global_index: Option<String>,
    global_outer: bool,
) -> Datasource {
    let mut ret = source;
    let mut subpath = Vec::new();
    let mut last_path_part_missing_options = true;
    for path_part in path.into_iter() {
        last_path_part_missing_options = path_part.options.is_empty();
        subpath.push(path_part.field);
        // if the options are empty, we are not unwinding at this point in the path
        if path_part.options.is_empty() {
            continue;
        }
        let mut index_prefix = None;
        for (i, options) in path_part.options.into_iter().enumerate() {
            if let Some(ip) = index_prefix {
                index_prefix = Some(format!("{}{}", ip, i - 1));
            } else {
                index_prefix = Some(subpath.join("_"));
            }
            let options = get_options(
                options,
                subpath.clone(),
                index_prefix.as_ref().unwrap(),
                global_index.as_ref(),
                global_outer,
            );
            ret = Datasource::Unwind(UnwindSource {
                datasource: Box::new(ret),
                options,
            });
        }
    }
    // for backward compatibility, if the last path part has no options, we need to unwind it
    // anyway. Even if the path part was specified as part[], the options will still consist of
    // one empty vector, rather than being and empty vector, so we do not insert an extra unwind.
    if last_path_part_missing_options {
        let options = get_options(
            Vec::new(),
            subpath.clone(),
            subpath.join("_").as_str(),
            global_index.as_ref(),
            global_outer,
        );
        ret = Datasource::Unwind(UnwindSource {
            datasource: Box::new(ret),
            options,
        });
    }
    ret
}
