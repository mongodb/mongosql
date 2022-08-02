pub(crate) mod agg_ir_to_mql;
pub(crate) mod ir_to_mql;
use ir_to_mql::MqlCodeGenerator;
pub use ir_to_mql::{MqlMappingRegistry, MqlTranslation};

use crate::{agg_ir, ir};

pub fn generate_mql_from_agg_ir(
    _plan: agg_ir::Stage,
) -> Result<MqlTranslation, agg_ir_to_mql::Error> {
    type Error = agg_ir_to_mql::Error;
    Err(Error::UnimplementedAggIR)
}

pub fn generate_mql_from_ir(plan: ir::Stage) -> Result<MqlTranslation, ir_to_mql::Error> {
    let cg = MqlCodeGenerator {
        mapping_registry: MqlMappingRegistry::new(),
        scope_level: 0u16,
    };

    cg.codegen_stage(plan).map(|mql| mql.replace_bot())
}
