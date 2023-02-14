use quote::quote;
use syn::File;

mod analysis;
mod gen_visitor_trait;
mod gen_walk_implementations;
mod util;

use gen_visitor_trait::gen_visitor_mod;
use gen_walk_implementations::gen_walk_mod;

#[cfg(test)]
mod test;

#[proc_macro]
pub fn generate_visitors(input: proc_macro::TokenStream) -> proc_macro::TokenStream {
    let tokens: File = syn::parse(input).expect("failed to parse macro body");
    let types = analysis::collect_types(&tokens);
    let visitor_mod = gen_visitor_mod(&types);
    let walk_mod = gen_walk_mod(&types);

    let expanded = quote! {
        #tokens
        #visitor_mod
        #walk_mod
    };

    proc_macro::TokenStream::from(expanded)
}
