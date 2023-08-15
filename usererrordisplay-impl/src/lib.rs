extern crate proc_macro;
use proc_macro::TokenStream;
use quote::quote;
use syn::{parse_macro_input, Data, DeriveInput};

#[proc_macro_derive(UserErrorDisplay)]
pub fn formatted_error_derive(input: TokenStream) -> TokenStream {
    let ast = parse_macro_input!(input as DeriveInput);
    if let Data::Enum(_) = &ast.data {
        let name = &ast.ident;
        let expanded = quote! {
            impl std::fmt::Display for #name {
                fn fmt(&self, f: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                    let message = if let Some(user_message) = self.user_message() {
                        format!("{}\n\tCaused by:\n\t{}", user_message, self.technical_message())
                    } else {
                        self.technical_message()
                    };
                    write!(f, "Error {}: {}", self.code(), message)

                }
            }

            impl std::error::Error for #name { }

        };
        TokenStream::from(expanded)
    } else {
        panic!("UserErrorDispaly can only be derived for enums.");
    }
}
