<script lang="ts">
	import { resume } from '../lib/stores/resume.svelte';
	import type { Detail, Item } from '../lib/types';
	import Icon from './Icon.svelte';
	import SortableList from './SortableList.svelte';

	let { item }: { item: Item } = $props();
</script>

<div class="details">
	<div class="details-label">Highlights</div>
	<SortableList items={item.details ?? []} reorder={(ids) => resume.reorderDetails(item.id, ids)}>
		{#snippet row(detail: Detail)}
			<div class="detail-row" class:excluded={!detail.included}>
				<input
					class="include"
					type="checkbox"
					checked={detail.included}
					title={detail.included ? 'Exclude from this resume' : 'Include in this resume'}
					onchange={(e) =>
						resume.toggleDetail(detail.id, (e.currentTarget as HTMLInputElement).checked)}
				/>
				<input
					placeholder="Achievement or responsibility"
					value={detail.text}
					oninput={(e) =>
						resume.updateDetail(detail as Detail, (e.currentTarget as HTMLInputElement).value)}
				/>
				<button
					class="icon-btn" title="Remove" onclick={() => resume.deleteDetail(detail.id)}>
					<Icon name="trash" size={13} />
				</button>
			</div>
		{/snippet}
	</SortableList>
	<button class="btn small ghost" onclick={() => resume.addDetail(item.id)}>
		<Icon name="plus" size={13} />
		Add highlight
	</button>
</div>

<style>
	.details {
		display: flex;
		flex-direction: column;
		gap: 6px;
		border-top: 1px dashed var(--border);
		padding-top: 10px;
		margin-top: 10px;
	}

	.details-label {
		font-size: 11px;
		font-weight: 600;
		text-transform: uppercase;
		letter-spacing: 0.06em;
		color: var(--muted-foreground);
	}

	.detail-row {
		display: flex;
		align-items: center;
		gap: 6px;
	}

	.detail-row.excluded input:not(.include) {
		opacity: 0.5;
	}

	.include {
		width: auto;
	}
</style>
