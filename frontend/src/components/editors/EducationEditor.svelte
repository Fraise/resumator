<script lang="ts">
	import { resume } from '../../lib/stores/resume.svelte';
	import type { Item, Section } from '../../lib/types';
	import Icon from '../Icon.svelte';
	import ItemCard from '../ItemCard.svelte';
	import SortableList from '../SortableList.svelte';
	import DetailsList from '../DetailsList.svelte';

	let { section }: { section: Section } = $props();

	const addLabel = $derived(
		section.items.length === 0 ? 'Add your first school entry' : 'Add school entry'
	);
</script>

<SortableList items={section.items} reorder={(ids) => resume.reorderItems(section.id, ids)}>
	{#snippet row(item: Item)}
		<ItemCard {item}>
			<div class="fields">
				<label class="field">
					<span>School</span>
					<input
						placeholder="University of Somewhere"
						value={item.data.school ?? ''}
						oninput={(e) => resume.updateItemData(item, 'school', (e.currentTarget as HTMLInputElement).value)}
					/>
				</label>
				<label class="field">
					<span>Degree</span>
					<input
						placeholder="B.Sc. Computer Science"
						value={item.data.degree ?? ''}
						oninput={(e) => resume.updateItemData(item, 'degree', (e.currentTarget as HTMLInputElement).value)}
					/>
				</label>
				<label class="field">
					<span>Field (optional)</span>
					<input
						placeholder="Computer Science"
						value={item.data.field ?? ''}
						oninput={(e) => resume.updateItemData(item, 'field', (e.currentTarget as HTMLInputElement).value)}
					/>
				</label>
				<label class="field">
					<span>Start</span>
					<input
						placeholder="2015"
						value={item.data.start ?? ''}
						oninput={(e) => resume.updateItemData(item, 'start', (e.currentTarget as HTMLInputElement).value)}
					/>
				</label>
				<label class="field">
					<span>End</span>
					<input
						placeholder="2019"
						value={item.data.end ?? ''}
						oninput={(e) => resume.updateItemData(item, 'end', (e.currentTarget as HTMLInputElement).value)}
					/>
				</label>
			</div>
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
