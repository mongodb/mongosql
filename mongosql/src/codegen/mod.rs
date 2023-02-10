pub(crate) mod air_to_mql;
pub(crate) mod mir_to_mql;
pub use crate::mapping_registry::MqlMappingRegistry;

use crate::{air, mir};

pub fn generate_mql_from_air(
    plan: air::Stage,
) -> Result<air_to_mql::MqlTranslation, air_to_mql::Error> {
    let cg = air_to_mql::MqlCodeGenerator { scope_level: 0u16 };

    cg.codegen_air_stage(plan)
}

pub fn generate_mql_from_mir(
    plan: mir::Stage,
) -> Result<mir_to_mql::MqlTranslation, mir_to_mql::Error> {
    let cg = mir_to_mql::MqlCodeGenerator {
        mapping_registry: MqlMappingRegistry::new(),
        scope_level: 0u16,
    };

    cg.codegen_stage(plan).map(|mql| mql.replace_bot())
}
