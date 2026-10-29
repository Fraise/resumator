<script lang="ts">
	let { name, size = 16 }: { name: string; size?: number } = $props();
</script>

<svg
	aria-hidden="true"
	width={size}
	height={size}
	viewBox="0 0 16 16"
	fill="none"
	stroke="currentColor"
	stroke-width="1.6"
	stroke-linecap="round"
	stroke-linejoin="round"
>
	{#if name === 'eye'}
		<path d="M1.5 8s2.7-5.2 6.5-5.2S14.5 8 14.5 8s-2.7 5.2-6.5 5.2S1.5 8 1.5 8z" />
		<circle cx="8" cy="8" r="2.1" />
	{:else if name === 'eye-off'}
		<path d="M1.5 8s2.7-5.2 6.5-5.2S14.5 8 14.5 8s-2.7 5.2-6.5 5.2S1.5 8 1.5 8z" />
		<path d="M2.5 13.5l11-11" />
	{:else if name === 'plus'}
		<path d="M8 2.5v11M2.5 8h11" />
	{:else if name === 'trash'}
		<path d="M2.5 4h11M6 4V2.5h4V4M4 4l.8 9.5h6.4L12 4" />
	{:else if name === 'up'}
		<path d="M4 10l4-4 4 4" />
	{:else if name === 'down'}
		<path d="M4 6l4 4 4-4" />
	{:else if name === 'left'}
		<path d="M10 4L6 8l4 4" />
	{:else if name === 'right'}
		<path d="M6 4l4 4-4 4" />
	{:else if name === 'sun'}
		<circle cx="8" cy="8" r="3" />
		<path d="M8 1v2M8 13v2M1 8h2M13 8h2M3 3l1.4 1.4M11.6 11.6L13 13M13 3l-1.4 1.4M4.4 11.6L3 13" />
	{:else if name === 'moon'}
		<path d="M13.5 9.5A6 6 0 1 1 6.5 2.5a5 5 0 0 0 7 7z" />
	{:else if name === 'download'}
		<path d="M8 2v8M4.5 7l3.5 3.5L11.5 7M2.5 13.5h11" />
	{:else if name === 'user'}
		<circle cx="8" cy="5.5" r="2.8" />
		<path d="M2.5 14c.7-2.8 2.9-4.2 5.5-4.2s4.8 1.4 5.5 4.2" />
	{:else if name === 'type'}
		<path d="M3.5 4.5h9" />
		<path d="M8 4.5v9" />
		<path d="M5.5 13.5h5" />
	{:else if name === 'copy'}
		<rect x="5.5" y="5.5" width="8" height="8" rx="1.5" />
		<path d="M10.5 3.5v-1a1 1 0 0 0-1-1h-7a1 1 0 0 0-1 1v7a1 1 0 0 0 1 1h1" />
	{:else if name === 'pencil'}
		<path d="M11.5 2.5l2 2L5 13l-2.7.7L3 11z" />
	{:else if name === 'grip'}
		<g fill="currentColor" stroke="none">
			<circle cx="6" cy="3.5" r="1.3" />
			<circle cx="10" cy="3.5" r="1.3" />
			<circle cx="6" cy="8" r="1.3" />
			<circle cx="10" cy="8" r="1.3" />
			<circle cx="6" cy="12.5" r="1.3" />
			<circle cx="10" cy="12.5" r="1.3" />
		</g>
	{:else if name === 'check'}
		<path d="M3 8.5l3.5 3.5L13 4.5" />
	{:else if name === 'x'}
		<path d="M3.5 3.5l9 9M12.5 3.5l-9 9" />
	{/if}
</svg>
