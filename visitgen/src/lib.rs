use quote::quote;
use syn::File;

mod analysis;
mod gen_visitor_trait;
mod gen_walk_implementations;
mod util;

mod gen_ref_visitor_trait;
mod gen_ref_walk_implementaion;
#[cfg(test)]
mod test;

#[proc_macro]
pub fn generate_visitors(input: proc_macro::TokenStream) -> proc_macro::TokenStream {
    let tokens: File = syn::parse(input).expect("failed to parse macro body");
    let types = analysis::collect_types(&tokens);
    let visitor_mod = gen_visitor_trait::gen_visitor_mod(&types);
    let walk_mod = gen_walk_implementations::gen_walk_mod(&types);

    let visitor_ref_mod = gen_ref_visitor_trait::gen_visitor_mod(&types);
    let walk_ref_mod = gen_ref_walk_implementaion::gen_walk_mod(&types);

    let expanded = quote! {
        #tokens
        #visitor_mod
        #walk_mod
        #visitor_ref_mod
        #walk_ref_mod
    };

    #[cfg(feature = "debug-visitor-output")]
    {
        let visitor_folder_path = format!("./target/{}_visit", types.get(0).unwrap().get_name());
        let walk_folder_path = format!("./target/{}_walk", types.get(0).unwrap().get_name());
        procout::procout(&visitor_ref_mod, None, Some(visitor_folder_path.as_str()));
        procout::procout(&walk_ref_mod, None, Some(walk_folder_path.as_str()));
    }

    proc_macro::TokenStream::from(expanded)
}
