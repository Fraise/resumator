<script lang="ts" generics="T extends { id: number }">
	import type { Snippet } from 'svelte';
	import Icon from './Icon.svelte';
	import { sortable } from '../lib/sortable.svelte';

	let {
		items,
		reorder,
		row,
		layout = 'list',
		kind = '',
		moveIn
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
		 * here. When absent, only same-list reordering is possible. */
		moveIn?: (itemId: number, index: number) => void;
	} = $props();

	let container: HTMLElement;
	const listId = `sl-${Math.random().toString(36).slice(2, 10)}`;

	let dragId = $state<number | null>(null);
	const targeted = $derived(sortable.target?.listId === listId);
	const dropIndex = $derived(targeted ? sortable.target?.index ?? null : null);

	const DRAG_THRESHOLD_PX = 4;
	const EDGE_SCROLL_PX = 24;

	let pendingId: number | null = null;
	let dragStarted = false;
	let startClientY = 0;
	let scroller: HTMLElement | null = null;

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
		const localReorder = (itemId: number, index: number) => {
			const ids = currentItems.map((i) => i.id);
			const from = ids.indexOf(itemId);
			if (from < 0) return;
			ids.splice(from, 1);
			let to = index;
			if (to > from) to -= 1;
			ids.splice(to, 0, itemId);
			const unchanged =
				ids.length === currentItems.length && ids.every((id, i) => id === currentItems[i].id);
			if (!unchanged) reorder(ids);
		};
		sortable.register(listId, {
			el: container,
			layout,
			accepts: (drag) =>
				drag.kind === kind && (drag.sourceListId === listId || !!moveIn),
			onDrop: (itemId, index) => {
				if (sortable.current?.sourceListId === listId) {
					localReorder(itemId, index);
				} else {
					moveIn?.(itemId, index);
				}
			}
		});
		return () => sortable.unregister(listId);
	});

	function onGripPointerDown(event: PointerEvent, item: T) {
		if (event.button !== 0) return;
		event.preventDefault();
		pendingId = item.id;
		dragStarted = false;
		startClientY = event.clientY;
		(event.currentTarget as HTMLElement).setPointerCapture(event.pointerId);
	}

	function onGripPointerMove(event: PointerEvent) {
		if (pendingId === null) return;
		if (!dragStarted) {
			if (Math.abs(event.clientY - startClientY) < DRAG_THRESHOLD_PX) return;
			dragStarted = true;
			dragId = pendingId;
			document.body.style.userSelect = 'none';
			scroller = findScrollParent(container);
			sortable.start({ kind, itemId: pendingId, sourceListId: listId });
		}
		event.preventDefault();
		sortable.target = sortable.locate(event.clientX, event.clientY);
		autoScroll(event.clientY);
	}

	function finishDrag(apply: boolean) {
		if (dragStarted && apply && dragId !== null) {
			const target = sortable.target;
			if (target) {
				sortable.get(target.listId)?.onDrop(dragId, target.index);
			}
		}
		pendingId = null;
		dragStarted = false;
		dragId = null;
		document.body.style.userSelect = '';
		sortable.end();
	}
</script>

<div
	class="sortable"
	class:flow={layout === 'flow'}
	class:targeted
	role="list"
	bind:this={container}
>
	{#each items as item, index (item.id)}
		{#if dropIndex === index && sortable.current?.itemId !== item.id}
			<div class="drop-marker" aria-hidden="true"></div>
		{/if}
		<div
			class="sortable-row"
			class:dragging={sortable.current?.itemId === item.id}
			role="listitem"
		>
			<span
				class="grip"
				role="button"
				tabindex="-1"
				aria-label="Drag to reorder"
				title="Drag to reorder"
				onpointerdown={(e) => onGripPointerDown(e, item)}
				onpointermove={onGripPointerMove}
				onpointerup={() => finishDrag(true)}
				onpointercancel={() => finishDrag(false)}
			>
				<Icon name="grip" size={14} />
			</span>
			<div class="row-content">
				{@render row(item)}
			</div>
		</div>
	{/each}
	{#if dropIndex === items.length && sortable.current !== null}
		<div class="drop-marker" aria-hidden="true"></div>
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
	}

	.sortable.flow .sortable-row {
		width: auto;
	}

	.sortable-row.dragging {
		opacity: 0.45;
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

	.row-content {
		flex: 1;
		min-width: 0;
	}

	.sortable.flow .row-content {
		flex: 0 1 auto;
	}

	.drop-marker {
		height: 2px;
		border-radius: 2px;
		background: var(--primary);
		margin: -1px 0;
	}

	.sortable.flow .drop-marker {
		width: 2px;
		height: auto;
		align-self: stretch;
		margin: 0 -1px;
	}
</style>
