package site
{
	content: {
		docs: {
			reference: {
				cli: {
					"cue-mod-tidy": {
						page: {
							cache: {
								multi_step: {
									"NS3B2POAQH36V3U5MJ4FJHAKB6APHGATMK7T2VCG5ALQP1V539M0====": [{
										doc:      ""
										cmd:      "cue help mod tidy"
										exitCode: 0
										output: """
												WARNING: THIS COMMAND IS EXPERIMENTAL.

												Tidy resolves all module dependencies in the current module and updates
												the cue.mod/module.cue file to reflect them.

												It also removes dependencies that are not needed.

												It will attempt to fetch modules that aren't yet present in the
												dependencies by fetching the latest available version from
												a registry.

												See "cue help environment" for details on how $CUE_REGISTRY is used to
												determine the modules registry.

												Note: you must enable the modules experiment with:
												\texport CUE_EXPERIMENT=modules
												for this command to work.

												Usage:
												  cue mod tidy [flags]

												Flags:
												      --check   check for tidiness only; do not update module.cue file

												Global Flags:
												  -E, --all-errors   print all available errors
												  -i, --ignore       proceed in the presence of errors
												  -s, --simplify     simplify output
												      --strict       report errors for lossy mappings
												      --trace        trace computation
												  -v, --verbose      print information about progress

												"""
									}]
								}
							}
						}
					}
				}
			}
		}
	}
}
