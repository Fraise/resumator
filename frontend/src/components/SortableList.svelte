<script lang="ts" generics="T extends { id: number }">
	import type { Snippet } from 'svelte';
	import Icon from './Icon.svelte';

	let {
		items,
		reorder,
		row
	}: {
		items: T[];
		reorder: (ids: number[]) => void;
		row: Snippet<[T]>;
	} = $props();

	let container: HTMLElement;
	let dragId = $state<number | null>(null);
	let dropIndex = $state<number | null>(null);

	const DRAG_THRESHOLD_PX = 4;
	const EDGE_SCROLL_PX = 24;

	let pendingId: number | null = null;
	let dragStarted = false;
	let startClientY = 0;
	let rowBounds: { top: number; bottom: number }[] = [];
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

	// Caches each row's vertical bounds relative to the scroll parent's
	// content, so drop positions stay correct while auto-scrolling.
	function measureRows() {
		scroller = findScrollParent(container);
		const scrollerTop = scroller ? scroller.getBoundingClientRect().top - scroller.scrollTop : 0;
		rowBounds = Array.from(container.querySelectorAll('.sortable-row')).map((el) => {
			const rect = (el as HTMLElement).getBoundingClientRect();
			const top = rect.top - scrollerTop;
			return { top, bottom: top + rect.height };
		});
	}

	function contentY(clientY: number): number {
		if (!scroller) return clientY;
		return clientY - scroller.getBoundingClientRect().top + scroller.scrollTop;
	}

	function updateDropIndex(clientY: number) {
		const y = contentY(clientY);
		let index = items.length;
		for (let i = 0; i < rowBounds.length; i++) {
			const middle = (rowBounds[i].top + rowBounds[i].bottom) / 2;
			if (y < middle) {
				index = i;
				break;
			}
		}
		dropIndex = index;
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
			dropIndex = null;
			document.body.style.userSelect = 'none';
			measureRows();
		}
		event.preventDefault();
		updateDropIndex(event.clientY);
		autoScroll(event.clientY);
	}

	function finishDrag(apply: boolean) {
		if (dragStarted && apply && dragId !== null && dropIndex !== null) {
			const ids = items.map((i) => i.id);
			const from = ids.indexOf(dragId);
			if (from >= 0) {
				ids.splice(from, 1);
				let to = dropIndex;
				if (to > from) to -= 1;
				ids.splice(to, 0, dragId);
				const unchanged =
					ids.length === items.length && ids.every((id, i) => id === items[i].id);
				if (!unchanged) reorder(ids);
			}
		}
		pendingId = null;
		dragStarted = false;
		dragId = null;
		dropIndex = null;
		document.body.style.userSelect = '';
	}

	function onGripPointerUp() {
		finishDrag(true);
	}
</script>

<div class="sortable" role="list" bind:this={container}>
	{#each items as item, index (item.id)}
		{#if dropIndex === index && dragId !== null && dragId !== item.id}
			<div class="drop-marker" aria-hidden="true"></div>
		{/if}
		<div class="sortable-row" class:dragging={dragId === item.id} role="listitem">
			<span
				class="grip"
				role="button"
				tabindex="-1"
				aria-label="Drag to reorder"
				title="Drag to reorder"
				onpointerdown={(e) => onGripPointerDown(e, item)}
				onpointermove={onGripPointerMove}
				onpointerup={onGripPointerUp}
				onpointercancel={() => finishDrag(false)}
			>
				<Icon name="grip" size={14} />
			</span>
			<div class="row-content">
				{@render row(item)}
			</div>
		</div>
	{/each}
	{#if dropIndex === items.length && dragId !== null}
		<div class="drop-marker" aria-hidden="true"></div>
	{/if}
</div>

<style>
	.sortable {
		display: flex;
		flex-direction: column;
		gap: 10px;
	}

	.sortable-row {
		display: flex;
		align-items: flex-start;
		gap: 4px;
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

	.drop-marker {
		height: 2px;
		border-radius: 2px;
		background: var(--primary);
		margin: -1px 0;
	}
</style>
