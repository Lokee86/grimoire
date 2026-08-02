package relationships.contracts

open class Base

interface Contract

interface DirectContract

interface ChildContract : Contract

annotation class Marker

class Outer {
    interface NestedContract
}
