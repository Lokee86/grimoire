using System;
using System.Collections.Generic;

namespace Smoke;

public interface IWorker
{
    int Work(int value);
}

public class BaseWorker
{
    protected int Count;

    public virtual int Work(int value)
    {
        Count += value;
        return Count;
    }
}

public sealed class Worker : BaseWorker, IWorker
{
    private readonly List<int> values = new();

    public override int Work(int value)
    {
        values.Add(value);
        Count = base.Work(value);
        return values.Count + Count;
    }
}

public static class Entry
{
    public static int Run()
    {
        IWorker worker = new Worker();
        var result = worker.Work(2);
        return result;
    }
}
