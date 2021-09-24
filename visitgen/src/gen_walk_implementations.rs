use lazy_static::lazy_static;
use syn::{
    punctuated::Punctuated, token::Comma, Fields, FieldsNamed, FieldsUnnamed, GenericArgument,
    ItemEnum, ItemStruct, Type,
};

use crate::{
    analysis::{get_generic_name, get_generic_type, get_relevant_type_info, EnumOrStruct},
    util::convert_to_snake_case,
};

use std::collections::HashSet;

lazy_static! {
    static ref COMPOUND_TYPES: HashSet<&'static str> = {
        let mut s = HashSet::new();
        s.insert("Box");
        s.insert("HashMap");
        s.insert("BTreeMap");
        s.insert("Option");
        s.insert("Vec");
        s
    };
}

/// gen_walk_implementations generates all the walk method implementations for
/// an entire file.
pub fn gen_walk_implementations(module_path: &[&str], types: &[EnumOrStruct]) {
    use std::env;
    use std::fs;
    use std::path::PathBuf;

    let out_dir = env::var_os("OUT_DIR").unwrap();
    let mut dest_path = PathBuf::new();
    dest_path.push(&out_dir);
    for dir in module_path[..module_path.len() - 1].iter() {
        dest_path.push(dir);
    }
    dest_path.push("walk.rs");

    let target_module_path = module_path.join("::");
    let self_module_path = module_path[..module_path.len() - 1].join("::");
    let mut out = format!(
        "pub mod walk {{
     use crate::{}::*;
     use crate::{}::visitor::Visitor;\n\n",
        target_module_path, self_module_path,
    );

    let type_set = types.iter().map(|x| x.get_name()).collect::<HashSet<_>>();
    for t in types.iter() {
        out.push_str(gen_walk_impelementation(&type_set, t).as_str());
    }

    out.push('}');
    fs::write(&dest_path, &out).expect("failed to write walk implementations");
}

fn gen_walk_impelementation(type_set: &HashSet<String>, t: &EnumOrStruct) -> String {
    let type_name = t.get_name();
    let mut out = format!(
        "
    impl {} {{
        #[allow(dead_code, unused_parens, clippy::double_parens)]
        pub fn walk<V>(self, _visitor: &mut V) -> Self
        where
            V: Visitor
        {{\n",
        type_name
    );

    out.push_str(
        match t {
            EnumOrStruct::Enum(e) => gen_walk_for_enum(type_set, e),
            EnumOrStruct::Struct(s) => gen_walk_for_struct(type_set, s),
        }
        .as_str(),
    );

    out.push_str(
        "        }
    }\n",
    );
    out
}

fn gen_walk_for_enum(type_set: &HashSet<String>, e: &ItemEnum) -> String {
    let mut out = "            match self {\n".to_string();
    let enum_name = e.ident.to_string();
    for v in e.variants.iter() {
        match &v.fields {
            // TODO: Make this support all forms (no necessary for mongosql, likely)
            Fields::Named(_) => {
                panic!(
                    "Not supporting named enum variants, please use a separate struct definition"
                )
            }
            Fields::Unnamed(u) => {
                if u.unnamed.len() != 1 {
                    panic!("enum variants must have either 1 or no arguments, please refactor your code")
                }
                let ty = &u.unnamed.first().expect("impossible failure").ty;
                out.push_str("                ");
                out.push_str(enum_name.as_str());
                out.push_str("::");
                out.push_str(v.ident.to_string().as_str());
                out.push_str("(x)");
                out.push_str(" => ");
                out.push_str(enum_name.as_str());
                out.push_str("::");
                out.push_str(v.ident.to_string().as_str());
                out.push('(');
                out.push_str(gen_walk_visit_type(type_set, ty, "x").as_str());
                out.push_str("),\n");
            }
            Fields::Unit => {
                out.push_str("                ");
                out.push_str(enum_name.as_str());
                out.push_str("::");
                out.push_str(v.ident.to_string().as_str());
                out.push_str(" => ");
                out.push_str(enum_name.as_str());
                out.push_str("::");
                out.push_str(v.ident.to_string().as_str());
                out.push_str(",\n");
            }
        }
    }
    out.push_str("            }\n");
    out
}

fn gen_walk_for_struct(type_set: &HashSet<String>, s: &ItemStruct) -> String {
    let type_name = s.ident.to_string();
    let mut out = "            ".to_string();
    out.push_str(&type_name);
    match &s.fields {
        Fields::Named(f) => {
            out.push_str("{\n");
            out.push_str(gen_named_fields(type_set, f).as_str());
            out.push_str("            }\n");
        }
        Fields::Unnamed(u) => {
            out.push_str("(\n");
            out.push_str(gen_unnamed_fields(type_set, u).as_str());
            out.push_str("            )\n");
        }
        Fields::Unit => out.push('\n'),
    }
    out
}

fn gen_walk_visit_type(type_set: &HashSet<String>, ty: &Type, field_name: &str) -> String {
    let (type_name, generic_args) = get_relevant_type_info(ty);

    if type_set.contains(&type_name) {
        format!(
            "_visitor.visit_{}({})",
            convert_to_snake_case(&type_name),
            field_name
        )
    } else {
        match type_name.as_str() {
            "Box" => gen_walk_visit_box(type_set, field_name, generic_args),
            "BTreeMap" => gen_walk_visit_map(
                type_set,
                field_name,
                generic_args,
                "std::collections::BTreeMap",
            ),
            "HashMap" => gen_walk_visit_map(
                type_set,
                field_name,
                generic_args,
                "std::collections::HashMap",
            ),
            "LinkedHashMap" => gen_walk_visit_map(
                type_set,
                field_name,
                generic_args,
                "linked_hash_map::LinkedHashMap",
            ),
            "UniqueLinkedHashMap" => gen_walk_visit_unique_map(
                type_set,
                field_name,
                generic_args,
                "crate::util::unique_linked_hash_map::UniqueLinkedHashMap",
            ),
            "Option" => gen_walk_visit_option(type_set, field_name, generic_args),
            "Vec" => gen_walk_visit_vec(type_set, field_name, generic_args),
            // We just move this type as is, we don't have a way to visit it
            _ => field_name.to_owned(),
        }
    }
}

fn gen_walk_visit_box(
    type_set: &HashSet<String>,
    field_name: &str,
    generic_args: Option<&Punctuated<GenericArgument, Comma>>,
) -> String {
    let generic_args = generic_args.expect("Box found with no generic arguments");
    if generic_args.len() != 1 {
        panic!("nonsensical Box definition found with more than one generic argument")
    }
    let box_generic = generic_args.first().expect("impossible failure");
    let box_type_name = get_generic_name(box_generic);
    if type_set.contains(&box_type_name) || COMPOUND_TYPES.contains(&box_type_name as &str) {
        let box_type = get_generic_type(box_generic);
        format!(
            "Box::new({})",
            gen_walk_visit_type(type_set, box_type, &(format!("(*{})", field_name))),
        )
    } else {
        field_name.to_owned()
    }
}

fn gen_walk_visit_unique_map(
    type_set: &HashSet<String>,
    field_name: &str,
    generic_args: Option<&Punctuated<GenericArgument, Comma>>,
    map_type_name: &str,
) -> String {
    let generic_args = generic_args.expect("HashMap found with no generic arguments");
    if generic_args.len() != 2 {
        panic!("nonsensical HashMap definition without two generic arguments")
    }
    let key_generic = generic_args.first().expect("impossible failure");
    let key_type_name = get_generic_name(key_generic);
    let key_special =
        type_set.contains(&key_type_name) || COMPOUND_TYPES.contains(&key_type_name as &str);

    let value_generic = generic_args.last().expect("impossible failure");
    let value_type_name = get_generic_name(value_generic);
    let value_special =
        type_set.contains(&value_type_name) || COMPOUND_TYPES.contains(&value_type_name as &str);

    if key_special {
        let key_type = get_generic_type(key_generic);
        if value_special {
            let value_type = get_generic_type(value_generic);
            format!(
                "{{let mut out = {}::new(); out.insert_many({}.into_iter().map(|(map_k, map_v)| ({}, {}))).unwrap(); out}}",
                map_type_name,
                field_name,
                gen_walk_visit_type(type_set, key_type, "map_k"),
                gen_walk_visit_type(type_set, value_type, "map_v"),
            )
        } else {
            format!(
                "{{let mut out = {}::new(); out.insert_many({}.into_iter().map(|(map_k, map_v)| ({}, map_v))).unwrap(); out}}",
                map_type_name,
                field_name,
                gen_walk_visit_type(type_set, key_type, "map_k"),
            )
        }
    } else if value_special {
        let value_type = get_generic_type(value_generic);
        format!("{{let mut out = {}::new(); out.insert_many({}.into_iter().map(|(map_k, map_v)| (map_k, {}))).unwrap(); out}}",
            map_type_name,
            field_name,
            gen_walk_visit_type(type_set, value_type, "map_v"),
        )
    } else {
        field_name.to_owned()
    }
}

fn gen_walk_visit_map(
    type_set: &HashSet<String>,
    field_name: &str,
    generic_args: Option<&Punctuated<GenericArgument, Comma>>,
    map_type_name: &str,
) -> String {
    let generic_args = generic_args.expect("HashMap found with no generic arguments");
    if generic_args.len() != 2 {
        panic!("nonsensical HashMap definition without two generic arguments")
    }
    let key_generic = generic_args.first().expect("impossible failure");
    let key_type_name = get_generic_name(key_generic);
    let key_special =
        type_set.contains(&key_type_name) || COMPOUND_TYPES.contains(&key_type_name as &str);

    let value_generic = generic_args.last().expect("impossible failure");
    let value_type_name = get_generic_name(value_generic);
    let value_special =
        type_set.contains(&value_type_name) || COMPOUND_TYPES.contains(&value_type_name as &str);

    if key_special {
        let key_type = get_generic_type(key_generic);
        if value_special {
            let value_type = get_generic_type(value_generic);
            format!(
                "{}.into_iter().map(|(map_k, map_v)| ({}, {})).collect::<{}<_,_>>()",
                field_name,
                gen_walk_visit_type(type_set, key_type, "map_k"),
                gen_walk_visit_type(type_set, value_type, "map_v"),
                map_type_name
            )
        } else {
            format!(
                "{}.into_iter().map(|(map_k, map_v)| ({}, map_v)).collect::<{}<_,_>>()",
                field_name,
                gen_walk_visit_type(type_set, key_type, "map_k"),
                map_type_name
            )
        }
    } else if value_special {
        let value_type = get_generic_type(value_generic);
        format!(
            "{}.into_iter().map(|(map_k, map_v)| (map_k, {})).collect::<{}<_,_>>()",
            field_name,
            gen_walk_visit_type(type_set, value_type, "map_v"),
            map_type_name
        )
    } else {
        field_name.to_owned()
    }
}

fn gen_walk_visit_option(
    type_set: &HashSet<String>,
    field_name: &str,
    generic_args: Option<&Punctuated<GenericArgument, Comma>>,
) -> String {
    let generic_args = generic_args.expect("Option found with no generic arguments");
    if generic_args.len() != 1 {
        panic!("nonsensical Option definition found with more than one generic argument")
    }
    let option_generic = generic_args.first().expect("impossible failure");
    let option_type_name = get_generic_name(option_generic);
    if type_set.contains(&option_type_name) || COMPOUND_TYPES.contains(&option_type_name as &str) {
        let option_type = get_generic_type(option_generic);
        format!(
            "{}.map(|opt_x| {})",
            field_name,
            gen_walk_visit_type(type_set, option_type, "opt_x"),
        )
    } else {
        field_name.to_owned()
    }
}

fn gen_walk_visit_vec(
    type_set: &HashSet<String>,
    field_name: &str,
    generic_args: Option<&Punctuated<GenericArgument, Comma>>,
) -> String {
    let generic_args = generic_args.expect("Vec found with no generic arguments");
    if generic_args.len() != 1 {
        panic!("nonsensical Vec definition found with more than one generic argument")
    }
    let vec_generic = generic_args.first().expect("impossible failure");
    let vec_type_name = get_generic_name(vec_generic);
    if type_set.contains(&vec_type_name) || COMPOUND_TYPES.contains(&vec_type_name as &str) {
        let vec_type = get_generic_type(vec_generic);
        format!(
            "{}.into_iter().map(|vec_x| {}).collect::<Vec<_>>()",
            field_name,
            gen_walk_visit_type(type_set, vec_type, "vec_x"),
        )
    } else {
        field_name.to_owned()
    }
}

fn gen_named_fields(type_set: &HashSet<String>, fields: &FieldsNamed) -> String {
    let mut out = String::new();
    for f in fields.named.iter() {
        let name = f.ident.as_ref().expect("impossible failure").to_string();
        out.push_str("                ");
        out.push_str(name.as_str());
        out.push_str(": ");
        let field_name = format!("self.{}", name.as_str());
        out.push_str(gen_walk_visit_type(type_set, &f.ty, field_name.as_str()).as_str());
        out.push_str(",\n");
    }
    out
}

fn gen_unnamed_fields(type_set: &HashSet<String>, fields: &FieldsUnnamed) -> String {
    let mut out = String::new();
    for (i, f) in fields.unnamed.iter().enumerate() {
        out.push_str("                ");
        let field_name = format!("self.{}", i);
        out.push_str(gen_walk_visit_type(type_set, &f.ty, field_name.as_str()).as_str());
        out.push_str(",\n");
    }
    out
}
