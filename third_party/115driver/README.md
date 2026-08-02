# 115driver compatibility patch

Based on `github.com/SheltonZhu/115driver` v1.3.5. Its upload client identifies
as 115Browser `27.0.5.7`; 115 now requires `36.0.0` and rejects uploads with
`请升级到最新版本`. Remove this local replacement after upstream updates both
`UA115Browser` and the signed upload `appVer`.
