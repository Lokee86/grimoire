# Rejected post-mitigation LevelDB trial

Date: July 29, 2026  
Included in controlled benchmark comparison: **no**

This trial used the revised narrow-task Grimoire path and completed successfully at the model level, but Hermes did not route the requested Fast/Priority service tier. Its usage report contains `service_tier: null`, so it is retained as diagnostic evidence and excluded from the controlled Sol/High/Fast comparison.

## Diagnostic result

- API calls: 16
- Fresh input: 96,125
- Cache reads: 782,848
- Output: 16,881
- Reasoning: 10,340
- Total tokens: 895,854
- Grimoire calls: 2
- Retrieval trajectory: one narrow handle-only search followed by one inspection
- Model completed: yes
- Fast/Priority verified: **no**

The result suggested improvement over the earlier 18-to-19-call runs, but it was not accepted until Fast routing was corrected and independently smoke-tested. The accepted controlled result is in `../rerun4/`.
