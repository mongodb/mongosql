use syn::{
    punctuated::Punctuated, token::Comma, visit::Visit, File, GenericArgument, ItemEnum,
    ItemStruct, PathArguments, Type,
};

/// EnumOrStruct represents either
/// an enum or struct in syn. Currently
/// we only generated visitors for enums and structs.
pub enum EnumOrStruct {
    Enum(Box<ItemEnum>),
    Struct(Box<ItemStruct>),
}

impl EnumOrStruct {
    pub fn get_name(&self) -> String {
        match self {
            EnumOrStruct::Enum(e) => e.ident.to_string(),
            EnumOrStruct::Struct(s) => s.ident.to_string(),
        }
    }
}

/// collect_types collects all the types in a syn::File
pub fn collect_types(file: &File) -> Vec<EnumOrStruct> {
    struct TypeVisitor {
        types: Vec<EnumOrStruct>,
    }

    impl<'ast> Visit<'ast> for TypeVisitor {
        fn visit_item_enum(&mut self, node: &'ast ItemEnum) {
            self.types.push(EnumOrStruct::Enum(Box::new(node.clone())));
        }

        fn visit_item_struct(&mut self, node: &'ast ItemStruct) {
            self.types
                .push(EnumOrStruct::Struct(Box::new(node.clone())));
        }
    }

    let mut type_vis = TypeVisitor { types: Vec::new() };
    type_vis.visit_file(file);
    type_vis.types
}

pub fn get_generic_name(g: &GenericArgument) -> String {
    let t = get_generic_type(g);
    let (type_name, _) = get_relevant_type_info(t);
    type_name
}

pub fn get_generic_type(g: &GenericArgument) -> &Type {
    match g {
        GenericArgument::Type(t) => t,
        _ => panic!("unsupported generic argument"),
    }
}

/// get_relevant_type_info gets all of the type info for a syn::Type
/// that we need for generating Visitor traits and walk methods.
pub fn get_relevant_type_info(ty: &Type) -> (String, Option<&Punctuated<GenericArgument, Comma>>) {
    match ty {
        Type::Path(p) => {
            let ty = p
                .path
                .segments
                .last()
                .expect("failed to get end of TypePath");
            let ty_name = ty.ident.to_string();
            let args = match &ty.arguments {
                PathArguments::AngleBracketed(x) => Some(&x.args),
                _ => None,
            };
            (ty_name, args)
        }
        _ => {
            panic!("currently we only support Type::Path based Types")
        }
    }
}
