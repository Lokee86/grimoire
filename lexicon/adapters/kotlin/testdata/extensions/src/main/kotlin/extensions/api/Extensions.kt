package extensions.api

import extensions.model.Item

fun Item.direct(value: Int): Item = this

fun Item.imported(): Item = this

fun Item.defaulted(value: Int = 0): Item = this

fun Item.spread(vararg values: Int): Item = this

fun Item.ambiguous(value: Int): Item = this

fun Item.ambiguous(value: String): Item = this

fun Item.fluent(): Item = this

fun Item.collision(): Item = this

fun String.externalOnly(): String = this
