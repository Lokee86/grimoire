package runtime.slice

fun top(value: Int): Int = value

fun choose(value: Int): Int = value

fun choose(value: String): String = value

open class Base(seed: Int) {
    open fun compute(value: Int): Int = value

    open fun String.render(value: Int): Int = value
}

interface Contract {
    fun compute(value: Int): Int
}

class Child(var count: Int) : Base(count), Contract {
    private val delegated: Int by lazy { 1 }

    override fun compute(value: Int): Int = value

    override fun String.render(value: Int): Int = value

    fun helper(value: Int): Int = value

    fun run(input: Int, worker: ExternalWorker) {
        helper(input)
        this.helper(input)
        top(input)
        Helpers.work(input)
        Factory.create(input)
        Child(input)
        choose(input)
        worker.run()
        println(input)
        worker?.unsupported()
        count = input
        this.count += input
        Helpers.state++
        delegated
    }

    fun shadows(input: Int) {
        var count = input
        count++
    }

    fun destructured(pair: Pair<Int, Int>) {
        var (count, other) = pair
        count++
        other++
    }

    fun delegatedLocal() {
        val count by lazy { 1 }
        count
    }

    fun nestedLocal() {
        fun local() {
            count++
        }
        local()
    }

    fun nestedLambda() {
        top(1).also { count++ }
    }
}

object Helpers {
    var state: Int = 0

    fun work(value: Int) {
        state += value
    }
}

class Factory {
    companion object {
        fun create(value: Int): Child = Child(value)
    }
}

open class Parent(value: Int)

class Secondary : Parent {
    constructor(value: Int) : super(value) {
        Helpers.work(value)
    }

    constructor() : this(0)
}
