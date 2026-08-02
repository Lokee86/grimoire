use super::model::{NodeKey, RelationKind, SourceSpan};

/// Why an adapter could not resolve a symbolic relationship target.
#[derive(Clone, Debug, Eq, Ord, PartialEq, PartialOrd, Hash)]
pub enum UnresolvedReason {
    MissingTarget,
    AmbiguousTarget,
    UnsupportedForm,
    DynamicTarget,
    ExternalTarget,
    BuiltinTarget,
    GeneratedTarget,
    TypeConversion,
    SelfTarget,
    UnsupportedMacroExpansion,
    MacroArgumentMismatch,
    MacroExpansionCycle,
    MacroExpansionDepth,
    CompilerAnalysisFailed,
    CompilerIdentityMismatch,
    Unknown(String),
}

impl UnresolvedReason {
    pub(crate) fn as_str(&self) -> &str {
        match self {
            Self::MissingTarget => "missing-target",
            Self::AmbiguousTarget => "ambiguous-target",
            Self::UnsupportedForm => "unsupported-form",
            Self::DynamicTarget => "dynamic-target",
            Self::ExternalTarget => "external-target",
            Self::BuiltinTarget => "builtin-target",
            Self::GeneratedTarget => "generated-target",
            Self::TypeConversion => "type-conversion",
            Self::SelfTarget => "self-target",
            Self::UnsupportedMacroExpansion => "unsupported-macro-expansion",
            Self::MacroArgumentMismatch => "macro-argument-mismatch",
            Self::MacroExpansionCycle => "macro-expansion-cycle",
            Self::MacroExpansionDepth => "macro-expansion-depth",
            Self::CompilerAnalysisFailed => "compiler-analysis-failed",
            Self::CompilerIdentityMismatch => "compiler-identity-mismatch",
            Self::Unknown(value) => value,
        }
    }

    pub(crate) fn parse(value: &str) -> Option<Self> {
        if value.is_empty() {
            return None;
        }
        Some(match value {
            "missing-target" => Self::MissingTarget,
            "ambiguous-target" => Self::AmbiguousTarget,
            "unsupported-form" => Self::UnsupportedForm,
            "dynamic-target" => Self::DynamicTarget,
            "external-target" => Self::ExternalTarget,
            "builtin-target" => Self::BuiltinTarget,
            "generated-target" => Self::GeneratedTarget,
            "type-conversion" => Self::TypeConversion,
            "self-target" => Self::SelfTarget,
            "unsupported-macro-expansion" => Self::UnsupportedMacroExpansion,
            "macro-argument-mismatch" => Self::MacroArgumentMismatch,
            "macro-expansion-cycle" => Self::MacroExpansionCycle,
            "macro-expansion-depth" => Self::MacroExpansionDepth,
            "compiler-analysis-failed" => Self::CompilerAnalysisFailed,
            "compiler-identity-mismatch" => Self::CompilerIdentityMismatch,
            other => Self::Unknown(other.to_owned()),
        })
    }

    pub(crate) fn is_unknown(&self) -> bool {
        matches!(self, Self::Unknown(_))
    }
}

/// A symbolic relationship that an adapter observed but could not resolve safely.
#[derive(Clone, Debug, Eq, Ord, PartialEq, PartialOrd, Hash)]
pub struct UnresolvedReferenceFact {
    pub source: NodeKey,
    pub relation: RelationKind,
    pub expression: String,
    pub candidate_namespace: Option<String>,
    pub candidate_name: Option<String>,
    pub reason: UnresolvedReason,
    pub span: Option<SourceSpan>,
}

#[cfg(test)]
mod tests {
    use super::UnresolvedReason;

    #[test]
    fn parses_known_macro_reasons() {
        for (value, expected) in [
            (
                "unsupported-macro-expansion",
                UnresolvedReason::UnsupportedMacroExpansion,
            ),
            (
                "macro-argument-mismatch",
                UnresolvedReason::MacroArgumentMismatch,
            ),
            (
                "macro-expansion-cycle",
                UnresolvedReason::MacroExpansionCycle,
            ),
            (
                "macro-expansion-depth",
                UnresolvedReason::MacroExpansionDepth,
            ),
        ] {
            let reason = UnresolvedReason::parse(value).unwrap();
            assert_eq!(reason, expected);
            assert_eq!(reason.as_str(), value);
            assert!(!reason.is_unknown());
        }
    }

    #[test]
    fn parses_known_compiler_reasons() {
        for (value, expected) in [
            (
                "compiler-analysis-failed",
                UnresolvedReason::CompilerAnalysisFailed,
            ),
            (
                "compiler-identity-mismatch",
                UnresolvedReason::CompilerIdentityMismatch,
            ),
        ] {
            let reason = UnresolvedReason::parse(value).unwrap();
            assert_eq!(reason, expected);
            assert_eq!(reason.as_str(), value);
            assert!(!reason.is_unknown());
        }
    }

    #[test]
    fn preserves_unknown_reason_labels() {
        let reason = UnresolvedReason::parse("future-adapter-signal").unwrap();
        assert_eq!(reason.as_str(), "future-adapter-signal");
        assert!(reason.is_unknown());
    }
}
