<script lang="ts" generics="T extends { id: number }">
	import type { Snippet } from 'svelte';
	import Icon from './Icon.svelte';
	import { sortable } from '../lib/sortable.svelte';
	import type { RowGeometry } from '../lib/sortable.svelte';

	let {
		items,
		reorder,
		row,
		layout = 'list',
		kind = '',
		moveIn,
		isDraggable = () => true
	}: {
		items: T[];
		reorder: (ids: number[]) => void;
		row: Snippet<[T]>;
		/** 'list' stacks rows vertically; 'flow' wraps them side by side. */
		layout?: 'list' | 'flow';
		/** Tag identifying what kind of thing this list drags; drags with a
		 * different kind are ignored by this list. */
		kind?: string;
		/** Called when an item from another list of the same kind is dropped
		 * here: insert itemId before beforeId, or at the end when null. */
		moveIn?: (itemId: number, beforeId: number | null) => void;
		/** Items for which this returns false are pinned: their grip is
		 * inert and they cannot be dragged. */
		isDraggable?: (item: T) => boolean;
	} = $props();

	let container: HTMLElement;
	const listId = `sl-${Math.random().toString(36).slice(2, 10)}`;

	let dragId = $state<number | null>(null);
	const targeted = $derived(sortable.target?.listId === listId);
	// A drag started in another list highlights the insertion edge with
	// an inset line; a drag started here moves the rows with transforms.
	const foreignDrag = $derived(
		sortable.current !== null && sortable.current.sourceListId !== listId && targeted
	);
	const slotEnd = $derived(foreignDrag && sortable.target?.beforeId === null);

	// Untransformed row geometry, measured when a drag starts and after
	// every items change. During a drag the DOM order never changes:
	// rows are moved with transforms only, so allocating the drop space
	// never dirties layout (that reflow was the source of the hitches).
	let geoCache: RowGeometry[] = [];
	// Per-row drag transform, item id -> css transform.
	let transforms = $state<Record<number, string>>({});
	// Suppresses the row transition for the single flush where a drop
	// commits, so rows don't double-move when the DOM reorders.
	let noAnim = $state(false);

	const dragItem = $derived(dragId === null ? null : (items.find((i) => i.id === dragId) ?? null));

	// Floating ghost follows the pointer; the grab offsets keep the exact
	// point the user grabbed under the cursor, so the card does not jump
	// when the drag starts.
	let dragX = $state(0);
	let dragY = $state(0);
	let grabDX = $state(0);
	let grabDY = $state(0);
	let ghostWidth = $state(0);

	const DRAG_THRESHOLD_PX = 4;
	const EDGE_SCROLL_PX = 24;

	let pendingId: number | null = null;
	let dragStarted = $state(false);
	let startClientY = 0;
	let scroller: HTMLElement | null = null;
	// Pointermove can fire far more often than the display refreshes;
	// the expensive work is coalesced into one pass per animation frame.
	let rafId: number | null = null;
	let lastX = 0;
	let lastY = 0;

	function measureRows(): RowGeometry[] {
		const listRect = container.getBoundingClientRect();
		const rows = Array.from(container.querySelectorAll(':scope > .sortable-row')) as HTMLElement[];
		return rows.map((r) => {
			const rect = r.getBoundingClientRect();
			return {
				id: Number(r.dataset.itemId),
				x: rect.left - listRect.left,
				y: rect.top - listRect.top,
				w: rect.width,
				h: rect.height
			};
		});
	}

	// Computes the transform that moves each row from its current DOM
	// position to where it will land after the drop: the dragged row
	// takes the slot at beforeId and the rows in between shift by
	// exactly one row stride. Pure math on the cached geometry — no DOM
	// reads, no layout.
	function computeTransforms(beforeId: number | null): Record<number, string> {
		const n = items.length;
		const f = items.findIndex((i) => i.id === dragId);
		if (f < 0 || geoCache.length !== n) return {};
		// Insertion index in the reduced (without the dragged row) order.
		let p: number;
		if (beforeId === null) {
			p = n - 1;
		} else {
			const bi = items.findIndex((i) => i.id === beforeId);
			p = bi < 0 || bi === f ? f : bi < f ? bi : bi - 1;
		}
		const reduced = items.map((_, i) => i).filter((i) => i !== f);
		const finalOrder = [...reduced.slice(0, p), f, ...reduced.slice(p)];
		const styles = getComputedStyle(container);
		const gy = parseFloat(styles.rowGap) || 0;
		const gx = parseFloat(styles.columnGap) || 0;
		const out: Record<number, string> = {};
		if (layout === 'list') {
			let y = geoCache[0].y;
			for (const oi of finalOrder) {
				out[items[oi].id] = `translate(0px, ${y - geoCache[oi].y}px)`;
				y += geoCache[oi].h + gy;
			}
		} else {
			// Flow: simulate the flex wrap to get each row's landing spot.
			const padL = geoCache[0].x;
			const padT = geoCache[0].y;
			const padR = parseFloat(styles.paddingRight) || 0;
			const contentW = container.clientWidth - padL - padR;
			let x = padL;
			let y = padT;
			let lineH = 0;
			for (const oi of finalOrder) {
				const w = geoCache[oi].w;
				const h = geoCache[oi].h;
				if (x > padL && x + w > padL + contentW) {
					x = padL;
					y += lineH + gy;
					lineH = 0;
				}
				out[items[oi].id] = `translate(${x - geoCache[oi].x}px, ${y - geoCache[oi].y}px)`;
				x += w + gx;
				lineH = Math.max(lineH, h);
			}
		}
		return out;
	}

	function findScrollParent(element: HTMLElement): HTMLElement | null {
		let el: HTMLElement | null = element.parentElement;
		while (el && el !== document.body) {
			const overflowY = getComputedStyle(el).overflowY;
			if ((overflowY === 'auto' || overflowY === 'scroll') && el.scrollHeight > el.clientHeight) {
				return el;
			}
			el = el.parentElement;
		}
		return null;
	}

	function autoScroll(clientY: number) {
		if (!scroller) return;
		const rect = scroller.getBoundingClientRect();
		if (clientY < rect.top + EDGE_SCROLL_PX) {
			scroller.scrollTop -= 12;
		} else if (clientY > rect.bottom - EDGE_SCROLL_PX) {
			scroller.scrollTop += 12;
		}
	}

	// Registers this list with the drag bus. Re-registered when props
	// change so the closures always see fresh items and callbacks.
	$effect(() => {
		const currentItems = items;
		const localReorder = (itemId: number, beforeId: number | null) => {
			const from = currentItems.findIndex((i) => i.id === itemId);
			if (from < 0) return;
			const rest = currentItems.filter((i) => i.id !== itemId);
			let to = rest.length;
			if (beforeId !== null && beforeId !== itemId) {
				const bi = rest.findIndex((i) => i.id === beforeId);
				if (bi >= 0) to = bi;
			}
			rest.splice(to, 0, currentItems[from]);
			const unchanged =
				rest.length === currentItems.length &&
				rest.every((it, i) => it.id === currentItems[i].id);
			if (!unchanged) reorder(rest.map((i) => i.id));
		};
		sortable.register(listId, {
			el: container,
			layout,
			accepts: (drag) =>
				drag.kind === kind && (drag.sourceListId === listId || !!moveIn),
			geometry: () => geoCache,
			onDrop: (itemId, beforeId) => {
				if (sortable.current?.sourceListId === listId) {
					localReorder(itemId, beforeId);
				} else {
					moveIn?.(itemId, beforeId);
				}
			}
		});
		// Re-measure once the DOM has settled (items may have changed).
		const raf = requestAnimationFrame(() => {
			geoCache = measureRows();
		});
		return () => {
			cancelAnimationFrame(raf);
			sortable.unregister(listId);
		};
	});

	function onGripPointerDown(event: PointerEvent, item: T) {
		if (event.button !== 0) return;
		event.preventDefault();
		pendingId = item.id;
		dragStarted = false;
		startClientY = event.clientY;
		lastX = event.clientX;
		lastY = event.clientY;
		dragX = event.clientX;
		dragY = event.clientY;
		const rowEl = (event.currentTarget as HTMLElement).closest('.sortable-row');
		const rect = rowEl?.getBoundingClientRect();
		if (rect) {
			grabDX = event.clientX - rect.left;
			grabDY = event.clientY - rect.top;
			ghostWidth = rect.width;
		}
		(event.currentTarget as HTMLElement).setPointerCapture(event.pointerId);
	}

	function onGripPointerMove(event: PointerEvent) {
		if (pendingId === null) return;
		lastX = event.clientX;
		lastY = event.clientY;
		if (!dragStarted) {
			if (Math.abs(lastY - startClientY) < DRAG_THRESHOLD_PX) return;
			dragStarted = true;
			dragId = pendingId;
			geoCache = measureRows();
			transforms = computeTransforms(pendingId);
			document.body.style.userSelect = 'none';
			scroller = findScrollParent(container);
			sortable.start({ kind, itemId: pendingId, sourceListId: listId });
		}
		event.preventDefault();
		if (rafId === null) {
			rafId = requestAnimationFrame(onDragFrame);
		}
	}

	function onDragFrame() {
		rafId = null;
		dragX = lastX;
		dragY = lastY;
		const t = sortable.locate(lastX, lastY);
		sortable.setTarget(t);
		if (t && t.listId === listId && dragId !== null) {
			transforms = computeTransforms(t.beforeId);
		}
		autoScroll(lastY);
	}

	function finishDrag(apply: boolean) {
		if (rafId !== null) {
			cancelAnimationFrame(rafId);
			rafId = null;
		}
		if (dragStarted && apply && dragId !== null) {
			// Flush the last pointer position so the drop lands where the
			// ghost visually is, not one frame behind.
			sortable.setTarget(sortable.locate(lastX, lastY));
			const target = sortable.target;
			// Clear the transforms without transition in the same flush
			// that commits the new order: the rows were already shown at
			// their final positions, so they land with no jump and no
			// double movement.
			noAnim = true;
			transforms = {};
			if (target) {
				sortable.get(target.listId)?.onDrop(dragId, target.beforeId);
			}
			requestAnimationFrame(() => {
				noAnim = false;
			});
		} else {
			// Cancelled: let the transition animate the rows back.
			transforms = {};
		}
		pendingId = null;
		dragStarted = false;
		dragId = null;
		document.body.style.userSelect = '';
		sortable.end();
	}
</script>

<svelte:window
	onresize={() => {
		if (!dragStarted) geoCache = measureRows();
	}}
/>

<div
	class="sortable"
	class:flow={layout === 'flow'}
	class:targeted
	class:slot-end={slotEnd}
	class:drag-active={dragStarted}
	class:no-anim={noAnim}
	role="list"
	bind:this={container}
>
	{#each items as item (item.id)}
		{@const draggable = isDraggable(item)}
		<div
			class="sortable-row"
			class:dragging={sortable.current?.itemId === item.id}
			class:slot-before={foreignDrag && sortable.target?.beforeId === item.id}
			data-item-id={item.id}
			style:transform={transforms[item.id] ?? ''}
			role="listitem"
		>
			<span
				class="grip"
				class:pinned={!draggable}
				role="button"
				tabindex="-1"
				aria-label={draggable ? 'Drag to reorder' : 'Pinned, cannot be moved'}
				title={draggable ? 'Drag to reorder' : 'Pinned, cannot be moved'}
				onpointerdown={draggable ? (e) => onGripPointerDown(e, item) : undefined}
				onpointermove={draggable ? onGripPointerMove : undefined}
				onpointerup={draggable ? () => finishDrag(true) : undefined}
				onpointercancel={draggable ? () => finishDrag(false) : undefined}
			>
				<Icon name={draggable ? 'grip' : 'lock'} size={14} />
			</span>
			<div class="row-content">
				{@render row(item)}
			</div>
		</div>
	{/each}

	{#if dragStarted && dragItem}
		<div
			class="drag-ghost"
			style:width="{ghostWidth}px"
			style:transform="translate({dragX - grabDX}px, {dragY - grabDY}px) rotate(-1.2deg)"
			aria-hidden="true"
		>
			<div class="sortable-row">
				<span class="grip"><Icon name="grip" size={14} /></span>
				<div class="row-content">{@render row(dragItem)}</div>
			</div>
		</div>
	{/if}
</div>

<style>
	.sortable {
		display: flex;
		flex-direction: column;
		gap: 10px;
		min-width: 0;
	}

	/* Flow layout: rows wrap side by side when there is room. */
	.sortable.flow {
		flex-direction: row;
		flex-wrap: wrap;
		align-content: flex-start;
		gap: 6px 4px;
	}

	.sortable.targeted {
		outline: 1.5px dashed var(--primary);
		outline-offset: 3px;
		border-radius: 8px;
	}

	.sortable-row {
		display: flex;
		align-items: flex-start;
		gap: 4px;
		/* Transform-only movement during a drag: the transition runs on
		   the compositor and never triggers a reflow. */
		transition: transform 160ms ease-out;
	}

	.sortable.drag-active .sortable-row {
		will-change: transform;
	}

	.sortable.no-anim .sortable-row {
		transition: none;
	}

	.sortable.flow .sortable-row {
		width: auto;
	}

	/* The dragged row stays in flow, moved to the slot by its transform,
	   as a dashed placeholder; the visible card is the ghost. */
	.sortable-row.dragging {
		outline: 1.5px dashed var(--primary);
		outline-offset: 3px;
		border-radius: 10px;
	}

	.sortable-row.dragging .grip,
	.sortable-row.dragging .row-content {
		visibility: hidden;
	}

	.grip {
		display: inline-flex;
		align-items: center;
		align-self: stretch;
		padding: 6px 0;
		cursor: grab;
		color: var(--muted-foreground);
		border-radius: 4px;
		touch-action: none;
	}

	.grip:hover {
		color: var(--foreground);
	}

	.grip:active {
		cursor: grabbing;
	}

	.grip.pinned {
		cursor: default;
		opacity: 0.45;
	}

	.grip.pinned:hover {
		color: var(--muted-foreground);
	}

	.row-content {
		flex: 1;
		min-width: 0;
	}

	.sortable.flow .row-content {
		flex: 0 1 auto;
	}

	/* Insertion edge for a drag coming from another list: an inset line
	   (box-shadow, so it never shifts the geometry locate() measures). */
	.sortable-row.slot-before {
		box-shadow: inset 0 2px 0 0 var(--primary);
		border-radius: 4px 4px 0 0;
	}

	.sortable.flow .sortable-row.slot-before {
		box-shadow: inset 2px 0 0 0 var(--primary);
		border-radius: 0 4px 4px 0;
	}

	.sortable.slot-end::after {
		content: '';
		display: block;
		height: 2px;
		border-radius: 2px;
		background: var(--primary);
	}

	.sortable.flow.slot-end::after {
		width: 2px;
		height: auto;
		align-self: stretch;
	}

	/* The card being dragged: fixed to the viewport and moved with the
	   pointer via transform (no layout thrash). */
	.drag-ghost {
		position: fixed;
		left: 0;
		top: 0;
		z-index: 1000;
		pointer-events: none;
		/* Own compositor layer: the per-frame transform then only
		   recomposites, instead of repainting the big shadow. */
		will-change: transform;
	}

	.drag-ghost .sortable-row {
		background: var(--card);
		border-radius: 10px;
		box-shadow:
			0 12px 32px rgba(0, 0, 0, 0.3),
			0 3px 10px rgba(0, 0, 0, 0.2);
	}
</style>
