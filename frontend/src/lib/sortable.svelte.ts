// Shared drag bus for sortable lists. It lets a drag started in one
// SortableList end in another (e.g. moving a skill chip between groups),
// and tracks the current drop target for marker rendering.

export interface DragInfo {
	kind: string;
	itemId: number;
	sourceListId: string;
}

export interface ListRegistration {
	el: HTMLElement;
	layout: 'list' | 'flow';
	accepts: (drag: DragInfo) => boolean;
	onDrop: (itemId: number, index: number) => void;
}

class SortableBus {
	current = $state<DragInfo | null>(null);
	target = $state<{ listId: string; index: number } | null>(null);

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

	end() {
		this.current = null;
		this.target = null;
	}

	// Finds the registered list under the pointer that accepts the current
	// drag, and the insertion index within it. Lists are checked in
	// registration order; lists that don't accept the drag are skipped, so
	// nested lists (group panels inside the panel list) resolve correctly
	// by drag kind.
	locate(clientX: number, clientY: number): { listId: string; index: number } | null {
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
			const rows = Array.from(reg.el.querySelectorAll(':scope > .sortable-row'));
			let index = rows.length;
			for (let i = 0; i < rows.length; i++) {
				const r = (rows[i] as HTMLElement).getBoundingClientRect();
				if (clientY < r.top) {
					index = i;
					break;
				}
				if (clientY <= r.bottom) {
					if (reg.layout !== 'flow' || clientX <= r.left + r.width / 2) {
						index = i;
						break;
					}
				}
			}
			return { listId: id, index };
		}
		return null;
	}
}

export const sortable = new SortableBus();
