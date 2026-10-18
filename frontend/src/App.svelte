<script lang="ts">
	import { onMount } from 'svelte';
	import { resume } from './lib/stores/resume.svelte';
	import { format, PAGE_SIZES } from './lib/format.svelte';
	import type { ResumeFormat } from './lib/format.svelte';
	import TopBar from './components/TopBar.svelte';
	import Sidebar from './components/Sidebar.svelte';
	import Editor from './components/Editor.svelte';
	import Preview from './components/Preview.svelte';

	const MIN_PANE_WIDTH = 480;
	const SPLIT_KEY = 'resumator-split';

	function initialWidth(): number {
		const saved = Number(localStorage.getItem(SPLIT_KEY));
		const max = window.innerWidth - MIN_PANE_WIDTH;
		if (saved >= MIN_PANE_WIDTH && saved <= max) return saved;
		return Math.min(Math.round(window.innerWidth * 0.5), max);
	}

	let leftWidth = $state(initialWidth());
	let dragging = $state(false);

	function onDividerDown(event: PointerEvent) {
		dragging = true;
		(event.currentTarget as HTMLElement).setPointerCapture(event.pointerId);
	}

	function onDividerMove(event: PointerEvent) {
		if (!dragging) return;
		const container = (event.currentTarget as HTMLElement).parentElement;
		if (!container) return;
		const rect = container.getBoundingClientRect();
		leftWidth = Math.min(
			Math.max(event.clientX - rect.left, MIN_PANE_WIDTH),
			rect.width - MIN_PANE_WIDTH
		);
	}

	function onDividerUp() {
		if (!dragging) return;
		dragging = false;
		localStorage.setItem(SPLIT_KEY, String(leftWidth));
	}

	onMount(() => {
		resume.init();
	});
</script>

<svelte:window
	onresize={() => (leftWidth = Math.min(leftWidth, window.innerWidth - MIN_PANE_WIDTH))}
/>

<TopBar />

<div class="split">
	<div class="pane pane-left" style:width="{leftWidth}px">
		<Sidebar />
		<Editor />
	</div>
	<div
		class="divider"
		class:active={dragging}
		role="separator"
		aria-orientation="vertical"
		onpointerdown={onDividerDown}
		onpointermove={onDividerMove}
		onpointerup={onDividerUp}
	></div>
	<div class="pane pane-right">
		<div class="preview-toolbar">
			<div class="format-switch" role="group" aria-label="Page format">
				{#each Object.entries(PAGE_SIZES) as [key, size] (key)}
					<button
						class:active={format.current === key}
						title={`Preview and PDF on ${size.label} paper`}
						onclick={() => {
							format.set(key as ResumeFormat);
							resume.refreshLayout();
						}}
					>
						{size.label}
					</button>
				{/each}
			</div>
		</div>
		<Preview />
	</div>
</div>

{#if !resume.loaded && !resume.error}
	<div class="loading">Loading…</div>
{/if}

<style>
	.split {
		flex: 1;
		display: flex;
		min-height: 0;
	}

	.pane {
		min-width: 0;
		min-height: 0;
		overflow: hidden;
		display: flex;
	}

	.pane-left {
		flex-shrink: 0;
		border-right: none;
		flex-direction: row;
	}

	.pane-left :global(nav.sidebar) {
		width: 232px;
		flex-shrink: 0;
		border-right: 1px solid var(--border);
		background: var(--card);
	}

	.pane-left > :global(.editor) {
		flex: 1;
		background: var(--background);
	}

	.divider {
		flex: 0 0 6px;
		cursor: col-resize;
		background: var(--border);
		touch-action: none;
	}

	.divider:hover,
	.divider.active {
		background: var(--primary);
	}

	.pane-right {
		flex: 1;
		background: var(--muted);
		flex-direction: column;
	}

	.preview-toolbar {
		display: flex;
		justify-content: flex-end;
		padding: 6px 12px;
		border-bottom: 1px solid var(--border);
		background: var(--card);
		flex-shrink: 0;
	}

	.format-switch {
		display: inline-flex;
		gap: 2px;
		background: var(--muted);
		border-radius: 9999px;
		padding: 2px;
	}

	.format-switch button {
		border-radius: 9999px;
		padding: 3px 12px;
		font-size: 12.5px;
		color: var(--muted-foreground);
	}

	.format-switch button.active {
		background: var(--card);
		color: var(--foreground);
		box-shadow: var(--shadow);
	}

	.loading {
		position: fixed;
		inset: 0;
		display: flex;
		align-items: center;
		justify-content: center;
		background: var(--background);
		color: var(--muted-foreground);
	}
</style>
