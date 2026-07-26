use super::format::Manifest;

#[test]
fn manifest_accepts_null_files_as_empty() {
    let manifest: Manifest = serde_json::from_str(
        r#"{
            "version": 1,
            "state_commit": "state",
            "languages": [{
                "language": "interstack",
                "adapter_version": "1",
                "schema_version": 1,
                "repository": "example/repository",
                "analysis_config_id": "sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
                "shared_object_id": null,
                "files": null
            }]
        }"#,
    )
    .unwrap();
    assert!(manifest.languages[0].files.is_empty());
}
