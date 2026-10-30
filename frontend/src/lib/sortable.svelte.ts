// Shared drag bus for sortable lists. It lets a drag started in one
// SortableList end in another (e.g. moving a skill chip between groups),
// and tracks the current drop target for marker rendering.

export interface DragInfo {
	kind: string;
	itemId: number;
	sourceListId: string;
}

/** Untransformed geometry of one row, relative to the list element's
 *  top-left corner. Aligned with the DOM rows. */
export interface RowGeometry {
	id: number;
	x: number;
	y: number;
	w: number;
	h: number;
}

export interface ListRegistration {
	el: HTMLElement;
	layout: 'list' | 'flow';
	accepts: (drag: DragInfo) => boolean;
	/** Untransformed row geometry aligned with the DOM rows. Read on
	 *  every locate, so it must be measured while no drag transforms
	 *  are applied. */
	geometry: () => RowGeometry[];
	/** beforeId is the item id of the row the pointer is above, or null
	 * to append at the end. */
	onDrop: (itemId: number, beforeId: number | null) => void;
}

export interface DropTarget {
	listId: string;
	/** DOM row index, kept for marker rendering. */
	index: number;
	/** Item id to insert before; null = append at the end. */
	beforeId: number | null;
}

class SortableBus {
	current = $state<DragInfo | null>(null);
	target = $state<DropTarget | null>(null);

	#lists = new Map<string, ListRegistration>();

	register(id: string, reg: ListRegistration) {
		this.#lists.set(id, reg);
	}

	unregister(id: string) {
		this.#lists.delete(id);
		if (this.target?.listId === id) this.target = null;
	}

	get(id: string): ListRegistration | undefined {
		return this.#lists.get(id);
	}

	start(drag: DragInfo) {
		this.current = drag;
		this.target = null;
	}

	// Assigns target only when it actually changed. Pointer jitter would
	// otherwise hand every list a fresh object per event and invalidate
	// all of their derived state (and re-diff every each block) for
	// nothing.
	setTarget(t: DropTarget | null) {
		const cur = this.target;
		if (t === null || cur === null) {
			if (t !== cur) this.target = t;
			return;
		}
		if (t.listId !== cur.listId || t.index !== cur.index || t.beforeId !== cur.beforeId) {
			this.target = t;
		}
	}

	end() {
		this.current = null;
		this.target = null;
	}

	// Finds the registered list under the pointer that accepts the current
	// drag, and the insertion index within it. Lists are checked in
	// registration order; lists that don't accept the drag are skipped, so
	// nested lists (group panels inside the panel list) resolve correctly
	// by drag kind.
	locate(clientX: number, clientY: number): DropTarget | null {
		const drag = this.current;
		if (!drag) return null;
		for (const [id, reg] of this.#lists) {
			if (!reg.accepts(drag)) continue;
			const rect = reg.el.getBoundingClientRect();
			const pad = 10;
			if (
				clientX < rect.left - pad ||
				clientX > rect.right + pad ||
				clientY < rect.top - pad ||
				clientY > rect.bottom + pad
			) {
				continue;
			}
			// One rect read for the list itself; row positions come from
			// the cached, untransformed geometry so hit-testing never
			// forces a reflow and is unaffected by drag transforms.
			const rows = reg.geometry();
			let index = rows.length;
			for (let i = 0; i < rows.length; i++) {
				const top = rect.top + rows[i].y;
				if (clientY < top) {
					index = i;
					break;
				}
				if (clientY <= top + rows[i].h) {
					if (reg.layout !== 'flow' || clientX <= rect.left + rows[i].x + rows[i].w / 2) {
						index = i;
						break;
					}
				}
			}
			return { listId: id, index, beforeId: index < rows.length ? rows[index].id : null };
		}
		return null;
	}
}

export const sortable = new SortableBus();
