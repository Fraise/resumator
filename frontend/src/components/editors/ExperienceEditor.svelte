<script lang="ts">
	import { resume } from '../../lib/stores/resume.svelte';
	import type { Item, Section } from '../../lib/types';
	import Icon from '../Icon.svelte';
	import ItemCard from '../ItemCard.svelte';
	import SortableList from '../SortableList.svelte';
	import DetailsList from '../DetailsList.svelte';

	let { section }: { section: Section } = $props();

	const addLabel = $derived(
		section.items.length === 0 ? 'Add your first position' : 'Add position'
	);
</script>

<SortableList items={section.items} reorder={(ids) => resume.reorderItems(section.id, ids)}>
	{#snippet row(item: Item)}
		<ItemCard {item}>
			<div class="fields">
				<label class="field">
					<span>Role</span>
					<input
						placeholder="Software Engineer"
						value={item.data.role ?? ''}
						oninput={(e) => resume.updateItemData(item, 'role', (e.currentTarget as HTMLInputElement).value)}
					/>
				</label>
				<label class="field">
					<span>Company</span>
					<input
						placeholder="Acme Inc."
						value={item.data.company ?? ''}
						oninput={(e) => resume.updateItemData(item, 'company', (e.currentTarget as HTMLInputElement).value)}
					/>
				</label>
				<label class="field">
					<span>Start</span>
					<input
						placeholder="2021-01"
						value={item.data.start ?? ''}
						oninput={(e) => resume.updateItemData(item, 'start', (e.currentTarget as HTMLInputElement).value)}
					/>
				</label>
				<label class="field">
					<span>End</span>
					<input
						placeholder="2024-06 (empty = present)"
						value={item.data.end ?? ''}
						oninput={(e) => resume.updateItemData(item, 'end', (e.currentTarget as HTMLInputElement).value)}
					/>
				</label>
			</div>
			<label class="field">
				<span>Summary</span>
				<textarea
					placeholder="One or two lines about what you did there"
					value={item.data.summary ?? ''}
					oninput={(e) => resume.updateItemData(item, 'summary', (e.currentTarget as HTMLTextAreaElement).value)}
				></textarea>
			</label>
			<DetailsList {item} />
		</ItemCard>
	{/snippet}
</SortableList>

<button class="btn ghost add-btn" onclick={() => resume.addItem(section.id)}>
	<Icon name="plus" size={14} />
	{addLabel}
</button>

<style>
	.fields {
		display: grid;
		grid-template-columns: 1fr 1fr;
		gap: 10px;
	}

	.add-btn {
		justify-content: center;
		border: 1px dashed var(--border);
		border-radius: 12px;
		padding: 10px;
		color: var(--muted-foreground);
	}
</style>
