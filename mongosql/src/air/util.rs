use crate::air::FieldRef;

impl FieldRef {
    pub(crate) fn root_parent(&self) -> String {
        match &self.parent {
            Some(parent) => parent.root_parent(),
            None => self.name.clone(),
        }
    }

    pub(crate) fn path_components(&self) -> Vec<String> {
        match &self.parent {
            None => vec![self.name.clone()],
            Some(parent) => {
                let mut components = parent.path_components();
                components.push(self.name.clone());
                components
            }
        }
    }
}

mod test {
    macro_rules! test_path_components {
        ($func_name:ident, expected = $expected:expr, input = $input:expr) => {
            #[test]
            fn $func_name() {
                use crate::air::FieldRef;

                let input = $input;
                let expected = $expected;
                let actual = input.path_components();

                assert_eq!(expected, actual)
            }
        };
    }

    test_path_components!(
        no_parent,
        expected = vec!["foo".to_string()],
        input = FieldRef {
            parent: None,
            name: "foo".to_string()
        }
    );

    test_path_components!(
        one_parent,
        expected = vec!["parent".to_string(), "child".to_string()],
        input = FieldRef {
            parent: Some(Box::new(FieldRef {
                parent: None,
                name: "parent".to_string(),
            })),
            name: "child".to_string()
        }
    );

    test_path_components!(
        multiple_ancestors,
        expected = vec![
            "a".to_string(),
            "b".to_string(),
            "c".to_string(),
            "d".to_string()
        ],
        input = FieldRef {
            parent: Some(Box::new(FieldRef {
                parent: Some(Box::new(FieldRef {
                    parent: Some(Box::new(FieldRef {
                        parent: None,
                        name: "a".to_string(),
                    })),
                    name: "b".to_string(),
                })),
                name: "c".to_string(),
            })),
            name: "d".to_string()
        }
    );
}
