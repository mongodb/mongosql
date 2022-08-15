pub(crate) mod agg_ir_to_mql;
pub(crate) mod ir_to_mql;
pub use crate::mapping_registry::MqlMappingRegistry;

use crate::{agg_ir, ir};

pub fn generate_mql_from_agg_ir(
    plan: agg_ir::Stage,
) -> Result<agg_ir_to_mql::MqlTranslation, agg_ir_to_mql::Error> {
    let cg = agg_ir_to_mql::MqlCodeGenerator { scope_level: 0u16 };

    cg.codegen_agg_ir_stage(plan)
}

pub fn generate_mql_from_ir(
    plan: ir::Stage,
) -> Result<ir_to_mql::MqlTranslation, ir_to_mql::Error> {
    let cg = ir_to_mql::MqlCodeGenerator {
        mapping_registry: MqlMappingRegistry::new(),
        scope_level: 0u16,
    };

    cg.codegen_stage(plan).map(|mql| mql.replace_bot())
}
