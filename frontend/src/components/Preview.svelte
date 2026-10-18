<script lang="ts">
	import { resume } from '../lib/stores/resume.svelte';
	import type { Item, Section } from '../lib/types';

	const state = $derived(resume.state);
	const visibleSections = $derived(
		(resume.state?.sections ?? []).filter((s) => s.visible)
	);

	function included<T extends { included: boolean }>(entries: T[] | null | undefined): T[] {
		return entries?.filter((e) => e.included) ?? [];
	}

	function contactLine(section: Section): string {
		return included(section.items)
			.map((i) => String(i.data.value ?? '').trim())
			.filter(Boolean)
			.join('   •   ');
	}

	function skillText(item: Item): string {
		const name = String(item.data.name ?? '').trim();
		const level = String(item.data.level ?? '').trim();
		if (!name) return '';
		return level ? `${name} (${level})` : name;
	}

	function dateRange(item: Item): string {
		const start = String(item.data.start ?? '').trim();
		const end = String(item.data.end ?? '').trim();
		if (!start && !end) return '';
		if (!end) return `${start} – Present`;
		if (!start) return end;
		return `${start} – ${end}`;
	}

	function entryTitle(item: Item, kind: 'experience' | 'education' | 'custom'): string {
		if (kind === 'experience') {
			const role = String(item.data.role ?? '').trim();
			const company = String(item.data.company ?? '').trim();
			if (role && company) return `${role} — ${company}`;
			return role || company;
		}
		if (kind === 'education') {
			const degree = String(item.data.degree ?? '').trim();
			const field = String(item.data.field ?? '').trim();
			const school = String(item.data.school ?? '').trim();
			const left = [degree, field].filter(Boolean).join(' — ');
			if (left && school) return `${left}, ${school}`;
			return left || school;
		}
		return String(item.data.label ?? '').trim();
	}
</script>

<div class="preview-backdrop">
	<div class="page">
		{#if state}
			<header class="resume-header">
				<h1 class="name">{state.profile.name || 'Your Name'}</h1>
				{#if state.profile.headline}
					<div class="headline">{state.profile.headline}</div>
				{/if}
				{#each visibleSections as section (section.id)}
					{#if section.type === 'contact' && contactLine(section)}
						<div class="contact">{contactLine(section)}</div>
					{/if}
				{/each}
			</header>

			{#each visibleSections as section (section.id)}
				{#if section.type !== 'contact'}
					<section class="resume-section">
						<h2 class="section-title">{section.title}</h2>
						{#if section.type === 'skills'}
							{@const skills = included(section.items).map(skillText).filter(Boolean)}
							{#if skills.length}
								<p class="skills">{skills.join(', ')}</p>
							{/if}
						{:else}
							{#each included(section.items) as item (item.id)}
								{#if entryTitle(item, section.type as 'experience' | 'education' | 'custom') || dateRange(item)}
									<div class="entry-head">
										<span class="entry-title">{entryTitle(item, section.type as 'experience' | 'education' | 'custom')}</span>
										<span class="entry-dates">{dateRange(item)}</span>
									</div>
								{/if}
								{#if section.type === 'experience' && item.data.summary}
									<p class="summary">{item.data.summary}</p>
								{/if}
								{#if section.type === 'custom' && item.data.text}
									<p class="summary">{item.data.text}</p>
								{/if}
								{#if included(item.details ?? []).length}
									<ul class="bullets">
										{#each included(item.details ?? []) as detail (detail.id)}
											{#if detail.text.trim()}
												<li>{detail.text.trim()}</li>
											{/if}
										{/each}
									</ul>
								{/if}
							{/each}
						{/if}
					</section>
				{/if}
			{/each}
		{/if}
	</div>
</div>

<style>
	.preview-backdrop {
		flex: 1;
		overflow: auto;
		padding: 24px;
	}

	.page {
		width: 210mm;
		max-width: 100%;
		min-height: 297mm;
		margin: 0 auto;
		background: #ffffff;
		color: #16161a;
		padding: 14mm 14mm;
		box-shadow: 0 2px 6px rgba(0, 0, 0, 0.18), 0 12px 36px rgba(0, 0, 0, 0.14);
		border-radius: 2px;
	}

	.resume-header {
		text-align: center;
		margin-bottom: 6mm;
	}

	.name {
		font-size: 21pt;
		letter-spacing: 0.01em;
	}

	.headline {
		font-size: 10.5pt;
		font-style: italic;
		color: #64646e;
		margin-top: 1mm;
	}

	.contact {
		font-size: 9pt;
		color: #64646e;
		margin-top: 2mm;
	}

	.resume-section {
		margin-top: 5mm;
	}

	.section-title {
		font-size: 10pt;
		text-transform: uppercase;
		letter-spacing: 0.08em;
		border-bottom: 1px solid #c9c9cf;
		padding-bottom: 1mm;
	}

	.skills {
		margin: 2mm 0 0;
		font-size: 9.5pt;
	}

	.entry-head {
		display: flex;
		justify-content: space-between;
		align-items: baseline;
		gap: 6mm;
		margin-top: 2.5mm;
	}

	.entry-title {
		font-weight: 700;
		font-size: 10pt;
	}

	.entry-dates {
		font-size: 9pt;
		color: #64646e;
		white-space: nowrap;
	}

	.summary {
		margin: 1mm 0 0;
		font-size: 9.5pt;
		color: #3c3c44;
	}

	.bullets {
		margin: 1mm 0 0;
		padding-left: 5mm;
		font-size: 9.5pt;
	}

	.bullets li {
		margin: 0.5mm 0;
	}
</style>
