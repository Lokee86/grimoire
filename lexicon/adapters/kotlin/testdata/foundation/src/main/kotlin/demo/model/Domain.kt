package demo.model

import demo.support.Helper as SupportHelper
import kotlin.collections.*

sealed interface Result<out T>

sealed class Outcome

data class User(
    val id: String,
    var nickname: String?,
    age: Int,
) : Result<User> {
    companion object Factory {
        suspend fun String?.decode(limit: Int? = null): User? = null
    }

    constructor(id: String) : this(id, null, 0)

    val displayName: String? = nickname

    fun rename(value: String?): Unit {
        nickname = value
    }
}

enum class Mode {
    FAST,
    SAFE;

    fun label(): String = name
}

object Registry {
    val current: User? = null
}

@JvmInline
value class UserId(val value: String)

class Empty

interface Service {
    suspend fun load(id: UserId?): Result<User>?
}

fun topLevel(input: User?): String? = input?.id
