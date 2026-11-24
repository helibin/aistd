# Pinned Admin source templates

These MIT-licensed, locally sanitized upstream source archives are embedded in the CLI.
They are application templates, separate from the pinned engineering/visual standards.

| Template | Version | Commit | Archive SHA256 |
| --- | --- | --- | --- |
| Vben Naive UI | 5.7.0 | `63a38dce49ba109f61607994e21ba921d8e970e9` | `afcd73a7fe5857f8dc2b44022350f6495ec6f0ad4b061914714ba257d0597fad` |
| Soybean-admin | 2.0.2 | `5e40b850180804aa51573590eb2e058d1fd8eb06` | `2007bc56ac5779f2a3da20e8720159277cc67567ee08af808458e070d99c9606` |

Based on GitHub codeload ZIPs for the above exact commits in
`vbenjs/vue-vben-admin` and `soybeanjs/soybean-admin`. Their LICENSE files remain
inside each archive and every generated Admin workspace.
The hashes identify the bundles shipped by aistd, not the original downloads.
The public bundles omit the Soybean demo token and replace its demo avatar;
the Vben bundle clears two playground login IDs and omits two community QR images.

Vben's workspace structure is retained; aistd invokes only the Naive UI application.
Soybean 2.0.2 uses Vite with a pinned lockfile, avoiding an unpinned Vite+/TypeScript
catalog migration. Neither template is mixed into the generated root npm workspaces.
The loader adapts the selected base URL to `/admin/`, uses the bundled Vben favicon
as the selected application's logo instead of a CDN URL, changes only the local registry
to npmjs.org, fixes Soybean's pnpm metadata to 10.33.4, and disables its browser auto-open.
It records original provenance in `.aistd/admin-template.json` and integration notes
in `apps/admin/AISTD-TEMPLATE.md`. No network requests occur during project generation.

To revise a source, download an exact upstream commit, retain its license, update the
archive/hash/version/commit together, review integration replacements, and run unit,
generated-project install/typecheck/build, and browser checks for both templates.
Existing generated projects are never upgraded automatically. `doctor` verifies only
the two installed standard sources, not these editable Admin application sources.
