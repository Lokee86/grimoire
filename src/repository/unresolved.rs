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
}

impl UnresolvedReason {
    pub(crate) fn as_str(&self) -> &'static str {
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
        }
    }

    pub(crate) fn parse(value: &str) -> Option<Self> {
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
            _ => return None,
        })
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
    fn parses_and_formats_generated_target() {
        let reason = UnresolvedReason::parse("generated-target").unwrap();
        assert_eq!(reason, UnresolvedReason::GeneratedTarget);
        assert_eq!(reason.as_str(), "generated-target");
    }
}
