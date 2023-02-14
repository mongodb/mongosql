use lazy_static::lazy_static;
use proc_macro2::TokenStream;
use quote::{format_ident, quote};
use syn::{
    punctuated::Punctuated, token::Comma, Fields, GenericArgument, ItemEnum, ItemStruct, Type,
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
        s.insert("LinkedHashMap");
        s.insert("UniqueLinkedHashMap");
        s.insert("BindingTuple");
        s.insert("BTreeMap");
        s.insert("Option");
        s.insert("Vec");
        s
    };
}

pub fn gen_walk_mod(types: &[EnumOrStruct]) -> TokenStream {
    let type_set = types.iter().map(|x| x.get_name()).collect::<HashSet<_>>();
    let walk_impls = types.iter().map(|t| gen_walk_implementation(&type_set, t));

    quote! {
        pub mod walk {
            use super::{*, visitor::Visitor};
            #(#walk_impls)*
        }
    }
}

fn gen_walk_implementation(type_set: &HashSet<String>, t: &EnumOrStruct) -> TokenStream {
    let type_name = format_ident!("{}", t.get_name());
    let walk_body = match t {
        EnumOrStruct::Enum(e) => gen_walk_for_enum(type_set, e),
        EnumOrStruct::Struct(s) => gen_walk_for_struct(type_set, s),
    };
    quote! {
        impl #type_name {
            pub fn walk<V>(self, _visitor: &mut V) -> Self where V: Visitor {
                #walk_body
            }
        }
    }
}

fn gen_walk_for_enum(type_set: &HashSet<String>, e: &ItemEnum) -> TokenStream {
    let enum_name = format_ident!("{}", e.ident);
    let variant_arms = e.variants.iter().map(|v| match &v.fields {
        Fields::Unnamed(u) if u.unnamed.len() == 1 => {
            let variant_name = format_ident!("{}", v.ident);
            let ty = &u.unnamed.first().unwrap().ty;
            let x = format_ident!("x");
            let x = quote!(#x);
            let visit_type = gen_walk_visit_type(type_set, ty, &x);
            quote! {
                #enum_name::#variant_name(#x) => #enum_name::#variant_name(#visit_type),
            }
        }
        Fields::Unit => {
            let variant_name = format_ident!("{}", v.ident);
            quote! {
                #enum_name::#variant_name => #enum_name::#variant_name,
            }
        }
        Fields::Unnamed(_) => {
            panic!("enum variants must have either 1 or no arguments, please refactor your code")
        }
        Fields::Named(_) => {
            panic!("Not supporting named enum variants, please use a separate struct definition")
        }
    });
    quote! {
        match self {
            #(#variant_arms)*
        }
    }
}

fn gen_walk_for_struct(type_set: &HashSet<String>, s: &ItemStruct) -> TokenStream {
    let type_name = format_ident!("{}", s.ident);
    match &s.fields {
        Fields::Named(fields) => {
            let field_assignments = fields.named.iter().map(|f| {
                let name = format_ident!("{}", f.ident.clone().unwrap());
                let visited_expr = quote!(self.#name);
                let visit_type = gen_walk_visit_type(type_set, &f.ty, &visited_expr);
                quote!(#name: #visit_type,)
            });
            quote!(#type_name { #(#field_assignments)* })
        }
        Fields::Unnamed(fields) => {
            let field_assignments = fields.unnamed.iter().enumerate().map(|(i, f)| {
                let visited_expr = quote!(self.#i);
                let visit_type = gen_walk_visit_type(type_set, &f.ty, &visited_expr);
                quote!(#visit_type,)
            });
            quote!(#type_name ( #(#field_assignments)* ))
        }
        Fields::Unit => quote!(),
    }
}

fn gen_walk_visit_type(
    type_set: &HashSet<String>,
    ty: &Type,
    visited_expr: &TokenStream,
) -> TokenStream {
    let (type_name, generic_args) = get_relevant_type_info(ty);

    if type_set.contains(&type_name) {
        let funcname = format_ident!("visit_{}", convert_to_snake_case(&type_name));
        quote!(_visitor.#funcname(#visited_expr))
    } else {
        match type_name.as_str() {
            "Box" => gen_walk_visit_box(type_set, visited_expr, generic_args),
            "BTreeMap" => gen_walk_visit_map(
                type_set,
                visited_expr,
                generic_args,
                &quote!(std::collections::BTreeMap),
            ),
            "HashMap" => gen_walk_visit_map(
                type_set,
                visited_expr,
                generic_args,
                &quote!(std::collections::HashMap),
            ),
            "LinkedHashMap" => gen_walk_visit_map(
                type_set,
                visited_expr,
                generic_args,
                &quote!(linked_hash_map::LinkedHashMap),
            ),
            "UniqueLinkedHashMap" => gen_walk_visit_unique_map(
                type_set,
                visited_expr,
                generic_args,
                &quote!(mongosql_datastructures::unique_linked_hash_map::UniqueLinkedHashMap),
            ),
            "Option" => gen_walk_visit_option(type_set, visited_expr, generic_args),
            "Vec" => gen_walk_visit_vec(type_set, visited_expr, generic_args),
            "BindingTuple" => gen_walk_visit_binding_tuple(type_set, visited_expr, generic_args),
            // We just move this type as is, we don't have a way to visit it
            _ => visited_expr.clone(),
        }
    }
}

fn gen_walk_visit_box(
    type_set: &HashSet<String>,
    visited_expr: &TokenStream,
    generic_args: Option<&Punctuated<GenericArgument, Comma>>,
) -> TokenStream {
    let generic_args = generic_args.expect("Box found with no generic arguments");
    if generic_args.len() != 1 {
        panic!("nonsensical Box definition found with more than one generic argument")
    }
    let box_generic = generic_args.first().expect("impossible failure");
    let box_type_name = get_generic_name(box_generic);
    if type_set.contains(&box_type_name) || COMPOUND_TYPES.contains(&box_type_name as &str) {
        let box_type = get_generic_type(box_generic);
        let visit_type = gen_walk_visit_type(type_set, box_type, &quote!((*#visited_expr)));
        quote!(Box::new(#visit_type))
    } else {
        visited_expr.clone()
    }
}

fn gen_walk_visit_unique_map(
    type_set: &HashSet<String>,
    visited_expr: &TokenStream,
    generic_args: Option<&Punctuated<GenericArgument, Comma>>,
    map_type_name: &TokenStream,
) -> TokenStream {
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
            let map_k = format_ident!("map_k");
            let map_k = quote!(#map_k);
            let visit_type_key = gen_walk_visit_type(type_set, key_type, &map_k);
            let map_v = format_ident!("map_v");
            let map_v = quote!(#map_v);
            let visit_type_value = gen_walk_visit_type(type_set, value_type, &map_v);
            quote!({
                let mut out = #map_type_name::new();
                out.insert_many(#visited_expr.into_iter().map(|(#map_k, #map_v)| (#visit_type_key, #visit_type_value))).unwrap();
                out
            })
        } else {
            let map_k = format_ident!("map_k");
            let map_k = quote!(#map_k);
            let visit_type = gen_walk_visit_type(type_set, key_type, &map_k);
            quote!({
                let mut out = #map_type_name::new();
                out.insert_many(#visited_expr.into_iter().map(|(#map_k, map_v)| (#visit_type, map_v))).unwrap();
                out
            })
        }
    } else if value_special {
        let value_type = get_generic_type(value_generic);
        let map_v = format_ident!("map_v");
        let map_v = quote!(#map_v);
        let visit_type = gen_walk_visit_type(type_set, value_type, &map_v);
        quote!({
            let mut out = #map_type_name::new();
            out.insert_many(#visited_expr.into_iter().map(|(map_k, #map_v)| (map_k, #visit_type))).unwrap();
            out
        })
    } else {
        visited_expr.clone()
    }
}

fn gen_walk_visit_map(
    type_set: &HashSet<String>,
    visited_expr: &TokenStream,
    generic_args: Option<&Punctuated<GenericArgument, Comma>>,
    map_type_name: &TokenStream,
) -> TokenStream {
    let generic_args = generic_args.expect("Map type found with no generic arguments");
    if generic_args.len() != 2 {
        panic!("nonsensical Map definition without two generic arguments")
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
            let map_k = format_ident!("map_k");
            let map_k = quote!(#map_k);
            let visit_type_key = gen_walk_visit_type(type_set, key_type, &map_k);
            let map_v = format_ident!("map_v");
            let map_v = quote!(#map_v);
            let visit_type_value = gen_walk_visit_type(type_set, value_type, &map_v);
            quote! {
                #visited_expr.into_iter()
                    .map(|(#map_k, #map_v)| (#visit_type_key, #visit_type_value))
                    .collect::<#map_type_name<_,_>>()
            }
        } else {
            let map_k = format_ident!("map_k");
            let map_k = quote!(#map_k);
            let visit_type_key = gen_walk_visit_type(type_set, key_type, &map_k);
            quote! {
                #visited_expr.into_iter()
                    .map(|(#map_k, map_v)| (#visit_type_key, map_v))
                    .collect::<#map_type_name<_,_>>()
            }
        }
    } else if value_special {
        let value_type = get_generic_type(value_generic);
        let map_v = format_ident!("map_v");
        let map_v = quote!(#map_v);
        let visit_type_value = gen_walk_visit_type(type_set, value_type, &map_v);
        quote! {
            #visited_expr.into_iter()
                .map(|(map_k, #map_v)| (map_k, #visit_type_value))
                .collect::<#map_type_name<_,_>>()
        }
    } else {
        visited_expr.clone()
    }
}

fn gen_walk_visit_option(
    type_set: &HashSet<String>,
    visited_expr: &TokenStream,
    generic_args: Option<&Punctuated<GenericArgument, Comma>>,
) -> TokenStream {
    let generic_args = generic_args.expect("Option found with no generic arguments");
    if generic_args.len() != 1 {
        panic!("nonsensical Option definition found with more than one generic argument")
    }
    let option_generic = generic_args.first().expect("impossible failure");
    let option_type_name = get_generic_name(option_generic);
    if type_set.contains(&option_type_name) || COMPOUND_TYPES.contains(&option_type_name as &str) {
        let option_type = get_generic_type(option_generic);
        let opt_x = format_ident!("opt_x");
        let opt_x = quote!(#opt_x);
        let visit_type = gen_walk_visit_type(type_set, option_type, &opt_x);
        quote!( #visited_expr.map(|#opt_x| #visit_type) )
    } else {
        visited_expr.clone()
    }
}

fn gen_walk_visit_vec(
    type_set: &HashSet<String>,
    visited_expr: &TokenStream,
    generic_args: Option<&Punctuated<GenericArgument, Comma>>,
) -> TokenStream {
    let generic_args = generic_args.expect("Vec found with no generic arguments");
    if generic_args.len() != 1 {
        panic!("nonsensical Vec definition found with more than one generic argument")
    }
    let vec_generic = generic_args.first().expect("impossible failure");
    let vec_type_name = get_generic_name(vec_generic);
    if type_set.contains(&vec_type_name) || COMPOUND_TYPES.contains(&vec_type_name as &str) {
        let vec_type = get_generic_type(vec_generic);
        let vec_x = format_ident!("vec_x");
        let vec_x = quote!(#vec_x);
        let visit_type = gen_walk_visit_type(type_set, vec_type, &vec_x);
        quote! {
            #visited_expr.into_iter()
                .map(|#vec_x| #visit_type)
                .collect::<Vec<_>>()
        }
    } else {
        visited_expr.clone()
    }
}

fn gen_walk_visit_binding_tuple(
    type_set: &HashSet<String>,
    visited_expr: &TokenStream,
    generic_args: Option<&Punctuated<GenericArgument, Comma>>,
) -> TokenStream {
    let generic_args = generic_args.expect("BindingTuple found with no generic arguments");
    if generic_args.len() != 1 {
        panic!("nonsensical BindingTuple definition found with more than one generic argument")
    }
    let bt_generic = generic_args.first().expect("impossible failure");
    let bt_type_name = get_generic_name(bt_generic);
    if type_set.contains(&bt_type_name) || COMPOUND_TYPES.contains(&bt_type_name as &str) {
        let bt_type = get_generic_type(bt_generic);
        let bt_x = format_ident!("bt_x");
        let bt_x = quote!(#bt_x);
        let visit_type = gen_walk_visit_type(type_set, bt_type, &bt_x);
        quote! {
            #visited_expr.into_iter()
                .map(|(k, #bt_x)| (k, #visit_type))
                .collect::<mongosql_datastructures::binding_tuple::BindingTuple<_>>()
        }
    } else {
        visited_expr.clone()
    }
}
