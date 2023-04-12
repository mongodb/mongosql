use super::MqlCodeGenerator;
use crate::air::{AggregationFunction, MQLOperator, SQLOperator};

impl MqlCodeGenerator {
    pub(crate) fn agg_func_to_mql_op(mqla: AggregationFunction) -> &'static str {
        use AggregationFunction::*;
        match mqla {
            AddToArray => "$push",
            Avg => "$avg",
            Count => unreachable!(),
            First => "$first",
            Last => "$last",
            Max => "$max",
            MergeDocuments => "$mergeObjects",
            Min => "$min",
            StddevPop => "$stdDevPop",
            StddevSamp => "$stdDevSamp",
            Sum => "$sum",
        }
    }

    pub(crate) fn agg_func_to_sql_op(mqla: AggregationFunction) -> &'static str {
        use AggregationFunction::*;
        match mqla {
            AddToArray => "$sqlPush",
            Avg => "$sqlAvg",
            Count => "$sqlCount",
            First => "$sqlFirst",
            Last => "$sqlLast",
            Max => "$sqlMax",
            MergeDocuments => "$sqlMergeObjects",
            Min => "$sqlMin",
            StddevPop => "$sqlStdDevPop",
            StddevSamp => "$sqlStdDevSamp",
            Sum => "$sqlSum",
        }
    }

    pub(crate) fn to_mql_op(mqlo: MQLOperator) -> &'static str {
        use MQLOperator::*;
        match mqlo {
            // String operators
            Concat => "$concat",

            // Conditional operators
            Cond => "$cond",

            // Arithmetic operators
            Add => "$add",
            Subtract => "$subtract",
            Multiply => "$multiply",
            Divide => "$divide",

            // Comparison operators
            Lt => "$lt",
            Lte => "$lte",
            Ne => "$ne",
            Eq => "$eq",
            Gt => "$gt",
            Gte => "$gte",

            // Boolean operators
            Not => "$not",
            And => "$and",
            Or => "$or",

            // Array scalar functions
            Slice => "$slice",
            Size => "$size",

            // Numeric value scalar functions
            IndexOfCP => "$indexOfCP",
            IndexOfBytes => "$indexOfBytes",
            StrLenCP => "$strLenCP",
            StrLenBytes => "$strLenBytes",
            Abs => "$abs",
            Ceil => "$ceil",
            Cos => "$cos",
            DegreesToRadians => "$degreesToRadians",
            Floor => "$floor",
            Log => "$log",
            Mod => "$mod",
            Pow => "$pow",
            RadiansToDegrees => "$radiansToDegrees",
            Round => "$round",
            Sin => "$sin",
            Tan => "$tan",
            Sqrt => "$sqrt",

            // String value scalar functions
            SubstrCP => "$substrCP",
            SubstrBytes => "$substrBytes",
            ToUpper => "$toUpper",
            ToLower => "$toLower",
            Trim => "$trim",
            LTrim => "$ltrim",
            RTrim => "$rtrim",
            Split => "$split",

            // Datetime value scalar function
            Year => "$year",
            Month => "$month",
            DayOfMonth => "$dayOfMonth",
            Hour => "$hour",
            Minute => "$minute",
            Second => "$second",
            Week => "$week",
            DayOfYear => "$dayOfYear",
            IsoWeek => "$isoWeek",
            IsoDayOfWeek => "$isoDayOfWeek",
            DateAdd => "$dateAdd",
            DateDiff => "$dateDiff",
            DateTrunc => "$dateTrunc",

            // MergeObjects merges an array of objects
            MergeObjects => "$mergeObjects",
        }
    }

    pub(crate) fn to_sql_op(sqlo: SQLOperator) -> Option<&'static str> {
        use SQLOperator::*;
        Some(match sqlo {
            // Arithmetic operators
            Divide => "$sqlDivide",
            Pos => "$sqlPos",
            Neg => "$sqlNeg",

            // Comparison operators
            Lt => "$sqlLt",
            Lte => "$sqlLte",
            Ne => "$sqlNe",
            Eq => "$sqlEq",
            Gt => "$sqlGt",
            Gte => "$sqlGte",
            Between => "$sqlBetween",

            // Boolean operators
            Not => "$sqlNot",
            And => "$sqlAnd",
            Or => "$sqlOr",

            // Conditional scalar functions
            NullIf => "$nullIf",
            Coalesce => "$coalesce",

            // Array scalar functions
            Slice => "$sqlSlice",
            Size => "$sqlSize",

            // Numeric value scalar functions
            IndexOfCP => "$sqlIndexOfCP",
            StrLenCP => "$sqlStrLenCP",
            StrLenBytes => "$sqlStrLenBytes",
            BitLength => "$sqlBitLength",
            Cos => "$sqlCos",
            Log => "$sqlLog",
            Mod => "$sqlMod",
            Round => "$sqlRound",
            Sin => "$sqlSin",
            Sqrt => "$sqlSqrt",
            Tan => "$sqlTan",

            // String value scalar functions
            SubstrCP => "$sqlSubstrCP",
            ToUpper => "$sqlToUpper",
            ToLower => "$sqlToLower",
            Split => "$sqlSplit",

            // ComputedFieldAccess, CurrentTimestamp
            _ => return None,
        })
    }
}
