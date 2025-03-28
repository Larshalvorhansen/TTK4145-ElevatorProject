import std.algorithm, std.concurrency, std.format, std.range, std.stdio, std.traits;
import core.thread, core.sync.mutex, core.sync.condition;

immutable Duration tick = 33.msecs;

class Resource(T) {
    private {
        T value;
        Mutex mtx;
        Condition cond;
        PriorityQueue!int queue;
        bool opptatt = false;
    }

    this() {
        mtx = new Mutex();
        cond = new Condition(mtx);
    }

    T allocate(int id, int prioritet) {
        mtx.lock();

        // Legg deg i kø med gitt prioritet
        queue.insert(id, prioritet);

        // Vent til du er først og ressursen er ledig
        while (queue.front() != id || opptatt) {
            cond.wait();
        }

        // Du har fått tilgang
        opptatt = true;
        queue.popFront();

        mtx.unlock();
        return value;
    }

    void deallocate(T v) {
        mtx.lock();

        opptatt = false;
        value = v;

        // Gi beskjed til andre som venter
        cond.notifyAll();

        mtx.unlock();
    }
}

struct PriorityQueue(T) {
    private {
        struct Elem {
            T val;
            int priority;
            string toString() {
                return format!("%s:%s")(priority, val);
            }
        }
        Elem[] queue;
    }

    void insert(T value, int priority) {
        queue ~= Elem(value, priority);
        queue.sort!((a, b) => a.priority > b.priority, SwapStrategy.stable);
    }

    T front() {
        assert(!queue.empty,
            format("Henter fra tom kø av %s", T.stringof));
        return queue[0].val;
    }

    void popFront() {
        assert(!queue.empty,
            format("Fjerner fra tom kø av %s", T.stringof));
        queue = queue.remove(0);
    }

    bool empty() {
        return queue.empty;
    }

    string toString() {
        return format!("%s([%(%s, %)])")(typeof(this).stringof, queue);
    }
}


void main() {
    auto resource = new Resource!(int[])();

    executionStates = new ExecutionState ;

    auto cfgs = [
        ResourceUserConfig(0, 0, 1, 1),
        ResourceUserConfig(1, 0, 3, 1),
        ResourceUserConfig(2, 1, 5, 1),
        ResourceUserConfig(0, 1, 10, 2),
        ResourceUserConfig(1, 0, 11, 1),
        ResourceUserConfig(2, 1, 11, 1),
        ResourceUserConfig(3, 0, 11, 1),
        ResourceUserConfig(4, 1, 11, 1),
        ResourceUserConfig(5, 0, 11, 1),
        ResourceUserConfig(6, 1, 11, 1),
        ResourceUserConfig(7, 0, 11, 1),
        ResourceUserConfig(8, 1, 11, 1),
        ResourceUserConfig(0, 1, 25, 3),
        ResourceUserConfig(6, 0, 26, 2),
        ResourceUserConfig(7, 0, 26, 2),
        ResourceUserConfig(1, 1, 26, 2),
        ResourceUserConfig(2, 1, 27, 2),
        ResourceUserConfig(3, 1, 28, 2),
        ResourceUserConfig(4, 1, 29, 2),
        ResourceUserConfig(5, 1, 30, 2),
    ];

    spawn(&executionLogger);
    foreach (cfg; cfgs) {
        spawnLinked(&resourceUser, cfg, cast(shared)resource);
    }
    foreach (_; 0 .. cfgs.length) {
        receive((LinkTerminated lt) {});
    }
    Thread.sleep(tick * 2);

    auto val = resource.allocate(-1, 0);

    assert(val.length == cfgs.length,
        "Test feilet: ikke alle brukere ble kjørt");
    assert(val[0..3] == [0, 1, 2],
        format("Test 1 feilet: feil rekkefølge, fikk %s", val[0..3]));

    assert(val[3] == 0,
        format("Test 2 feilet: feil rekkefølge, fikk %s", val[3]));
    assert(val[4..8].all!("(a & 1) == 0"),
        format("Test 2 feilet: høyt prioriterte (partall) kom ikke først: %s", val[4..8]));
    assert(val[8..12].all!("a & 1"),
        format("Test 2 feilet: lavt prioriterte (oddetall) kom ikke til slutt: %s", val[8..12]));

    assert(val[12] == 0,
        format("Test 3 feilet: feil rekkefølge, fikk %s", val[12]));
    assert(val[13..18].all!("a >= 1") && val[13..18].all!("a <= 5"),
        format("Test 3 feilet: høyt prioriterte brukere mangler: %s", val[13..18]));
    assert(val[18..20].all!("a >= 6") && val[18..20].all!("a <= 7"),
        format("Test 3 feilet: lavt prioriterte brukere mangler: %s", val[18..20]));

    writeln("Alle tester passerer");
}


struct ResourceUserConfig {
    int id;
    int priority;
    int release;
    int execute;
}

void resourceUser(ResourceUserConfig cfg, shared Resource!(int[]) r) {
    Thread.getThis.isDaemon = true;
    auto resource = cast(Resource!(int[]))r;

    Thread.sleep(cfg.release * tick);

    executionStates[cfg.id] = ExecutionState.waiting;
    auto val = resource.allocate(cfg.id, cfg.priority);

    executionStates[cfg.id] = ExecutionState.executing;

    Thread.sleep(cfg.execute * tick);
    val ~= cfg.id;
    resource.deallocate(val);

    executionStates[cfg.id] = ExecutionState.done;
}


version(Windows) {
    enum ExecutionState : char {
        none = ' ',
        waiting = cast(char)177,
        executing = cast(char)178,
        done = cast(char)223,
    }
    enum Grid : char {
        none = ' ',
        horizontal = cast(char)196,
    }
} else {
    enum ExecutionState : wchar {
        none = ' ',
        waiting = '\u2592',
        executing = '\u2593',
        done = '\u2580',
    }
    enum Grid : wchar {
        none = ' ',
        horizontal = '\u2500',
    }
}

__gshared ExecutionState[] executionStates;

void executionLogger() {
    Thread.getThis.isDaemon = true;
    Thread.sleep(tick / 2);
    auto t = 0;

    writefln("  id:%(%3d%)", iota(0, executionStates.length));

    while (true) {
        writef("%04d : ", t);
        foreach (id, ref state; executionStates) {
            auto grid = (t % 5 == 0) ? Grid.horizontal : Grid.none;
            writef("%c%c%c", cast(OriginalType!ExecutionState)state, grid, grid);
            if (state == ExecutionState.done) {
                state = ExecutionState.none;
            }
        }
        writeln;
        t++;
        Thread.sleep(tick);
    }
}
