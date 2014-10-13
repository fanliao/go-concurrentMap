package main

const (
	/**
	 * The default initial capacity for this table,
	 * used when not otherwise specified in a constructor.
	 */
	DEFAULT_INITIAL_CAPACITY int = 16

	/**
	 * The default load factor for this table, used when not
	 * otherwise specified in a constructor.
	 */
	DEFAULT_LOAD_FACTOR float32 = 0.75

	/**
	 * The default concurrency level for this table, used when not
	 * otherwise specified in a constructor.
	 */
	DEFAULT_CONCURRENCY_LEVEL int = 16

	/**
	 * The maximum capacity, used if a higher value is implicitly
	 * specified by either of the constructors with arguments.  MUST
	 * be a power of two <= 1<<30 to ensure that entries are indexable
	 * using ints.
	 */
	MAXIMUM_CAPACITY int = 1 << 30

	/**
	 * The maximum number of segments to allow; used to bound
	 * constructor arguments.
	 */
	MAX_SEGMENTS int = 1 << 16 // slightly conservative

	/**
	 * Number of unsynchronized retries in size and containsValue
	 * methods before resorting to locking. This is used to avoid
	 * unbounded retries if tables undergo continuous modification
	 * which would make it impossible to obtain an accurate result.
	 */
	RETRIES_BEFORE_LOCK int = 2
)

type concurrentMap struct {
    /**
     * Mask value for indexing into segments. The upper bits of a
     * key's hash code are used to choose the segment.
     */
    segmentMask int

    /**
     * Shift value for indexing within segments.
     */
    segmentShift int

    /**
     * The segments, each of which is a specialized hash table
     */
    segments []Segment

    keySet []interface{}
    entrySet []*Entry
    values []interface{}
}

type Entry 
