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

type ConcurrentMap struct {
	/**
	 * Mask value for indexing into segments. The upper bits of a
	 * key's hash code are used to choose the segment.
	 */
	segmentMask int

	/**
	 * Shift value for indexing within segments.
	 */
	segmentShift uint

	/**
	 * The segments, each of which is a specialized hash table
	 */
	segments []*Segment

	keySet   []interface{}
	entrySet []*Entry
	values   []interface{}
}

/**
 * Returns the segment that should be used for key with given hash
 * @param hash the hash code for the key
 * @return the segment
 */
func (this *ConcurrentMap) segmentFor(hash uint32) *Segment {
	return this.segments[(hash>>this.segmentShift)&uint32(this.segmentMask)]
}

/**
 * ConcurrentHashMap list entry. Note that this is never exported
 * out as a user-visible Map.Entry.
 *
 * Because the value field is volatile, not final, it is legal wrt
 * the Java Memory Model for an unsynchronized reader to see null
 * instead of initial value when read via a data race.  Although a
 * reordering leading to this is not likely to ever actually
 * occur, the Segment.readValueUnderLock method is used as a
 * backup in case a null (pre-initialized) value is ever seen in
 * an unsynchronized access method.
 */
type Entry struct {
	key   interface{}
	hash  uint32
	value interface{}
	next  *Entry
}

type Segment struct {
	/**
	 * The number of elements in this segment's region.
	 */
	count int

	/**
	 * Number of updates that alter the size of the table. This is
	 * used during bulk-read methods to make sure they see a
	 * consistent snapshot: If modCounts change during a traversal
	 * of segments computing size or checking containsValue, then
	 * we might have an inconsistent view of state so (usually)
	 * must retry.
	 */
	modCount int

	/**
	 * The table is rehashed when its size exceeds this threshold.
	 * (The value of this field is always <tt>(int)(capacity *
	 * loadFactor)</tt>.)
	 */
	threshold int

	/**
	 * The per-segment table.
	 */
	table []*Entry

	/**
	 * The load factor for the hash table.  Even though this value
	 * is same for all segments, it is replicated to avoid needing
	 * links to outer object.
	 * @serial
	 */
	loadFactor float32
}

func newSegment(initialCapacity int, lf float32) (s *Segment) {
	s = new(Segment)
	s.loadFactor = lf
	s.table = make([]*Entry, initialCapacity)
	return
}

/**
 * Applies a supplemental hash function to a given hashCode, which
 * defends against poor quality hash functions.  This is critical
 * because ConcurrentHashMap uses power-of-two length hash tables,
 * that otherwise encounter collisions for hashCodes that do not
 * differ in lower or upper bits.
 */
func hash(h uint32) uint32 {
	// Spread bits to regularize both segment and index locations,
	// using variant of single-word Wang/Jenkins hash.
	h += (h << 15) ^ 0xffffcd7d
	h ^= (h >> 10)
	h += (h << 3)
	h ^= (h >> 6)
	h += (h << 2) + (h << 14)
	return h ^ (h >> 16)
}
