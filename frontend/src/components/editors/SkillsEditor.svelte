<script lang="ts">
	import { resume } from '../../lib/stores/resume.svelte';
	import type { Item, Section } from '../../lib/types';
	import type { Action } from 'svelte/action';
	import Icon from '../Icon.svelte';
	import ItemCard from '../ItemCard.svelte';
	import SortableList from '../SortableList.svelte';

	let { section }: { section: Section } = $props();

	interface Group {
		/** Synthetic id for the sortable list: divider id, or 0 for the
		 * ungrouped area holding the skills before any divider. */
		id: number;
		dividerId: number | null;
		title: string;
		showTitle: boolean;
		skills: Item[];
	}

	// Groups derive from the flat item order: a divider item starts a
	// group; the skills after it belong to it until the next divider.
	const groups = $derived.by(() => {
		const result: Group[] = [
			{ id: 0, dividerId: null, title: '', showTitle: false, skills: [] }
		];
		for (const item of section.items ?? []) {
			if (item.data.divider) {
				result.push({
					id: item.id,
					dividerId: item.id,
					title: String(item.data.title ?? ''),
					showTitle: item.included,
					skills: []
				});
			} else {
				result[result.length - 1].skills.push(item);
			}
		}
		return result;
	});

	const hasGroups = $derived(groups.length > 1);

	// id of the skill card currently being edited (expanded form).
	let editingId = $state<number | null>(null);
	const cardEls = new Map<number, HTMLElement>();

	function dividerItem(group: Group): Item | null {
		if (group.dividerId === null) return null;
		return (section.items ?? []).find((i) => i.id === group.dividerId) ?? null;
	}

	// Tracks the DOM element of the expanded card so outside clicks can be
	// detected; removes the entry when the expanded card is destroyed.
	const editingCard: Action<HTMLElement, number> = (node, id) => {
		if (id != null) cardEls.set(id, node);
		return {
			destroy() {
				if (id != null) cardEls.delete(id);
			}
		};
	};

	function onWindowPointerDown(event: PointerEvent) {
		if (editingId === null) return;
		const row = cardEls.get(editingId)?.closest('.sortable-row');
		if (!row || !row.contains(event.target as Node)) {
			editingId = null;
		}
	}

	function edit(item: Item) {
		editingId = item.id;
	}

	function doneOnEnter(event: KeyboardEvent) {
		if (event.key === 'Enter') {
			event.preventDefault();
			editingId = null;
		}
	}

	function focusInput(node: HTMLInputElement) {
		node.focus();
	}

	function flatIds(): number[] {
		return (section.items ?? []).map((i) => i.id);
	}

	// Reordering groups keeps each group's skills attached to it.
	function reorderGroups(orderedIds: number[]) {
		const byId = new Map(groups.map((g) => [g.id, g]));
		const ids: number[] = [];
		for (const groupId of orderedIds) {
			const group = byId.get(groupId);
			if (!group) continue;
			if (group.dividerId !== null) ids.push(group.dividerId);
			ids.push(...group.skills.map((s) => s.id));
		}
		resume.reorderItems(section.id, ids);
	}

	// Reorders skills within one group by patching the group's contiguous
	// run in the flat item order.
	function reorderGroupSkills(group: Group, newSkillIds: number[]) {
		const ids = flatIds();
		const runStart = group.dividerId === null ? 0 : ids.indexOf(group.dividerId) + 1;
		if (group.dividerId !== null && runStart === 0) return;
		ids.splice(runStart, group.skills.length, ...newSkillIds);
		resume.reorderItems(section.id, ids);
	}

	// Moves a skill into a group, inserted before beforeId (or at the end
	// of the group's run when beforeId is null).
	function moveSkill(itemId: number, targetDividerId: number | null, beforeId: number | null) {
		const ids = flatIds();
		const from = ids.indexOf(itemId);
		if (from < 0) return;
		ids.splice(from, 1);
		const isDivider = (id: number) =>
			(section.items ?? []).find((i) => i.id === id)?.data.divider === true;
		let insertAt: number;
		if (beforeId !== null) {
			insertAt = ids.indexOf(beforeId);
			if (insertAt < 0) return;
		} else if (targetDividerId === null) {
			// End of the ungrouped run: just before the first divider.
			const firstDivider = ids.findIndex(isDivider);
			insertAt = firstDivider < 0 ? ids.length : firstDivider;
		} else {
			// End of the group's run: just before the next divider.
			const divIdx = ids.indexOf(targetDividerId);
			if (divIdx < 0) return;
			insertAt = divIdx + 1;
			while (insertAt < ids.length && !isDivider(ids[insertAt])) insertAt++;
		}
		ids.splice(insertAt, 0, itemId);
		resume.reorderItems(section.id, ids);
	}
</script>

<svelte:window onpointerdown={onWindowPointerDown} />

<div class="panels" class:has-groups={hasGroups}>
	<SortableList items={groups} layout="flow" kind="skill-group" reorder={reorderGroups}>
		{#snippet row(group: Group)}
			<section class="panel" class:ungrouped={group.dividerId === null}>
				<header class="panel-header">
					{#if group.dividerId === null}
						<span class="ungrouped-label">{hasGroups ? 'Ungrouped' : 'Skills'}</span>
					{:else}
						{@const divider = dividerItem(group)}
						{#if divider}
							<input
								class="group-title"
								placeholder="Group title (e.g. Languages)"
								value={divider.data.title ?? ''}
								oninput={(e) =>
									resume.updateItemData(divider, 'title', (e.currentTarget as HTMLInputElement).value)}
							/>
							<label class="include-label" title="Include this group (title and skills) in the current resume">
								<input
									class="include"
									type="checkbox"
									checked={divider.included}
									onchange={(e) =>
										resume.toggleItem(divider.id, (e.currentTarget as HTMLInputElement).checked)}
								/>
							</label>
							<button
								class="icon-btn"
								title="Remove group (skills are kept in the previous group)"
								onclick={() => resume.deleteItem(divider.id)}
							>
								<Icon name="trash" size={13} />
							</button>
						{/if}
					{/if}
				</header>

				<SortableList
					items={group.skills}
					layout="flow"
					kind="skill"
					reorder={(ids) => reorderGroupSkills(group, ids)}
					moveIn={(itemId, beforeId) => moveSkill(itemId, group.dividerId, beforeId)}
				>
					{#snippet row(item: Item)}
						{#if editingId === item.id}
							<div class="expanded" use:editingCard={item.id}>
								<ItemCard {item}>
									<div class="fields">
										<label class="field">
											<span>Skill</span>
											<input
												use:focusInput
												placeholder="Go"
												value={item.data.name ?? ''}
												oninput={(e) =>
													resume.updateItemData(item, 'name', (e.currentTarget as HTMLInputElement).value)}
												onkeydown={doneOnEnter}
											/>
										</label>
										<label class="field">
											<span>Level (optional)</span>
											<input
												placeholder="Expert"
												value={item.data.level ?? ''}
												oninput={(e) =>
													resume.updateItemData(item, 'level', (e.currentTarget as HTMLInputElement).value)}
												onkeydown={doneOnEnter}
											/>
										</label>
									</div>
								</ItemCard>
							</div>
						{:else}
							<div class="skill-chip" class:excluded={!item.included}>
								<input
									class="include"
									type="checkbox"
									title={item.included ? 'Exclude from this resume' : 'Include in this resume'}
									checked={item.included}
									onchange={(e) =>
										resume.toggleItem(item.id, (e.currentTarget as HTMLInputElement).checked)}
								/>
								<button class="chip-open" title="Edit skill" onclick={() => edit(item)}>
									<span class="chip-name">{item.data.name || 'New skill'}</span>
									{#if item.data.level}
										<span class="chip-level">{item.data.level}</span>
									{/if}
								</button>
								<button class="icon-btn" title="Edit" onclick={() => edit(item)}>
									<Icon name="pencil" size={13} />
								</button>
							</div>
						{/if}
					{/snippet}
				</SortableList>

				<button
					class="btn small ghost add-skill"
					title={group.dividerId === null ? 'Add an ungrouped skill' : 'Add a skill to this group'}
					onclick={() => resume.addItemToGroup(section.id, group.dividerId)}
				>
					<Icon name="plus" size={13} />
					Add skill
				</button>
			</section>
		{/snippet}
	</SortableList>
</div>

<button class="btn ghost add-group-btn" onclick={() => resume.addSkillGroup(section.id)}>
	<Icon name="plus" size={14} />
	Add group
</button>

<style>
	.panels {
		display: flex;
		flex-direction: column;
	}

	.panel {
		background: var(--card);
		border: 1px solid var(--border);
		border-radius: 12px;
		padding: 8px 10px 10px;
		max-width: 100%;
		display: flex;
		flex-direction: column;
		gap: 8px;
	}

	/* The ungrouped area is just a bare container unless groups exist. */
	.panel.ungrouped {
		background: transparent;
		border: none;
		padding: 0;
	}

	.panels.has-groups .panel.ungrouped {
		background: transparent;
		border: 1px dashed var(--border);
		border-radius: 12px;
		padding: 8px 10px 10px;
	}

	.panel-header {
		display: flex;
		align-items: center;
		gap: 8px;
	}

	.ungrouped-label {
		font-size: 11px;
		font-weight: 600;
		text-transform: uppercase;
		letter-spacing: 0.06em;
		color: var(--muted-foreground);
	}

	.group-title {
		flex: 0 1 auto;
		width: 220px;
		font-weight: 600;
	}

	.include-label {
		display: inline-flex;
		align-items: center;
		cursor: pointer;
	}

	.include {
		width: auto;
	}

	/* The row holding the expanded card spans the full width. */
	:global(.sortable-row):has(.expanded) {
		width: 100%;
	}

	.fields {
		display: grid;
		grid-template-columns: 2fr 1fr;
		gap: 10px;
	}

	.skill-chip {
		display: inline-flex;
		align-items: center;
		gap: 8px;
		max-width: 100%;
		background: var(--card);
		border: 1px solid var(--border);
		border-radius: 9999px;
		padding: 3px 4px 3px 10px;
	}

	.panel.ungrouped .skill-chip {
		background: var(--card);
	}

	.skill-chip.excluded {
		opacity: 0.55;
	}

	.chip-open {
		display: inline-flex;
		align-items: center;
		gap: 6px;
		padding: 3px 8px;
		border-radius: 9999px;
		color: var(--foreground);
		min-width: 0;
		text-align: left;
	}

	.chip-open:hover {
		background: var(--muted);
	}

	.chip-name {
		overflow: hidden;
		text-overflow: ellipsis;
		white-space: nowrap;
		max-width: 300px;
		font-weight: 500;
	}

	.chip-level {
		color: var(--muted-foreground);
		font-size: 12px;
		white-space: nowrap;
	}

	.add-skill {
		align-self: flex-start;
	}

	.add-group-btn {
		justify-content: center;
		border: 1px dashed var(--border);
		border-radius: 12px;
		padding: 10px;
		color: var(--muted-foreground);
		margin-top: 12px;
	}
</style>
