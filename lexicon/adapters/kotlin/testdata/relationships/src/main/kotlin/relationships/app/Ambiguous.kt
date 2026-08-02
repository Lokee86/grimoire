package relationships.app

import relationships.ambiguous.one.*
import relationships.ambiguous.two.*

class AmbiguousChild : Shared

@SharedMarker
class AmbiguousAnnotated

class ExternalChild : external.Base(), ExternalContract by externalDelegate

@ExternalMarker
class ExternalAnnotated
