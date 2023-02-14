use crate::{analysis::EnumOrStruct, util::convert_to_snake_case};
use proc_macro2::TokenStream;
use quote::{format_ident, quote};

pub fn gen_visitor_mod(types: &[EnumOrStruct]) -> TokenStream {
    let visit_funcs = types.iter().map(|t| {
        let func_name = format_ident!("visit_{}", convert_to_snake_case(&t.get_name()));
        let type_name = format_ident!("{}", t.get_name());
        let full_type_name = quote!(super::#type_name);
        quote! {
            fn #func_name(&mut self, node: #full_type_name) -> #full_type_name {
                node.walk(self)
            }
        }
    });

    quote! {
        pub mod visitor {
            pub trait Visitor: Sized {
                #(#visit_funcs)*
            }
        }
    }
}
