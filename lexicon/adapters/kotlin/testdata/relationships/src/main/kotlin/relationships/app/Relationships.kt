package relationships.app

import relationships.contracts.Base as Parent
import relationships.contracts.Contract as AliasContract
import relationships.contracts.DirectContract
import relationships.contracts.Marker as AliasMarker
import relationships.contracts.Outer as AliasOuter
import relationships.wild.*

annotation class LocalMarker

open class LocalBase

@AliasMarker
class AliasChild : Parent(), AliasContract

class DirectChild : DirectContract

@WildMarker
class WildChild : WildContract

@LocalMarker
class LocalChild : LocalBase()

class ExactChild : relationships.contracts.Base()

class NestedAliasChild : AliasOuter.NestedContract

class Lexical {
    interface InnerContract
    annotation class InnerMarker

    @InnerMarker
    class InnerChild : InnerContract
}

object Delegating : AliasContract by delegate
