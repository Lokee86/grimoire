package extensions.app

import extensions.api.ambiguous
import extensions.api.collision
import extensions.api.defaulted
import extensions.api.direct
import extensions.api.externalOnly
import extensions.api.imported as renamed
import extensions.api.spread
import extensions.model.Item as ModelItem
import extensions.model.Other
import extensions.wild.*
import external.library.thirdParty

fun ModelItem.samePackage(): ModelItem = this

class Usage(seed: ModelItem) {
    private val property: ModelItem = seed

    private fun ModelItem.lexical(value: Int): ModelItem = this

    fun throughParameter(item: ModelItem) {
        item.direct(1)
        item.defaulted()
        item.spread(1, 2)
    }

    fun throughLocal() {
        val item: ModelItem = property
        item.direct(2)
    }

    fun throughProperty() {
        property.direct(3)
    }

    fun throughImports(item: ModelItem) {
        item.renamed()
        item.wild()
        item.lexical(4)
        item.samePackage()
    }

    fun ambiguous(item: ModelItem) {
        item.ambiguous(1)
    }

    fun external(text: String, item: ModelItem) {
        text.externalOnly()
        item.thirdParty()
    }

    fun memberWins(item: ModelItem) {
        item.collision()
    }

    fun unsupported(item: ModelItem, dynamic: Any) {
        item.direct(1).fluent()
        item?.direct(2)
        dynamic.direct(3)
        val inferred = item
        inferred.direct(4)
    }

    fun shadowing(item: ModelItem, flag: Boolean) {
        item.direct(5)
        if (flag) {
            val item = property
            item.direct(6)
            val property: Other = Other()
            property.direct(7)
        }
        item.direct(8)
        property.direct(9)
    }
}
