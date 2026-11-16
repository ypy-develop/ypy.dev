// The site's own behaviour: the search box.
//
// Everything is delegated from the document, so nothing has to be set up again
// after YPY swaps a page in. Without this script the box is a form that asks
// /search, and works.
(() => {
	const MAX = 8;
	let index = null;
	let loading = null;
	let active = -1;

	const load = () => {
		if (index) return Promise.resolve(index);
		if (!loading) {
			loading = fetch('/search-index.json', { credentials: 'same-origin' })
				.then((r) => (r.ok ? r.json() : null))
				.then((j) => (index = (j && j.data && j.data.entries) || []))
				.catch(() => (index = []));
		}
		return loading;
	};

	// The same ranking as Index.Search in app/Docs/search.go.
	const search = (query) => {
		const words = query.toLowerCase().split(/\s+/).filter(Boolean);
		if (!words.length || !index) return [];
		const out = [];
		for (const e of index) {
			const title = e.t.toLowerCase();
			const heads = (e.h || []).join(' ').toLowerCase();
			const desc = (e.d || '').toLowerCase();
			const text = (e.x || '').toLowerCase();
			let score = 0;
			let ok = true;
			for (const w of words) {
				let s = 0;
				if (title === w) s = 40;
				else if (title.startsWith(w)) s = 24;
				else if (title.includes(w)) s = 16;
				if (heads.includes(w)) s += 6;
				if (desc.includes(w)) s += 3;
				if (text.includes(w)) s += 1;
				if (s === 0) { ok = false; break; }
				score += s;
			}
			if (ok) out.push({ e, score });
		}
		out.sort((a, b) => b.score - a.score);
		return out.slice(0, MAX).map((r) => r.e);
	};

	const box = () => document.querySelector('[data-site-search]');
	const input = () => document.querySelector('[data-site-search-input]');
	const panel = () => document.querySelector('[data-site-search-results]');

	const close = () => {
		const p = panel();
		const i = input();
		if (p) { p.hidden = true; p.textContent = ''; }
		if (i) i.setAttribute('aria-expanded', 'false');
		active = -1;
	};

	const render = (results, query) => {
		const p = panel();
		const i = input();
		if (!p || !i) return;
		p.textContent = '';
		active = -1;
		if (!query) { close(); return; }
		if (!results.length) {
			const none = document.createElement('p');
			none.className = 'site-search-none';
			none.textContent = 'Nothing found for "' + query + '".';
			p.appendChild(none);
		}
		results.forEach((r, n) => {
			const a = document.createElement('a');
			a.className = 'site-search-hit';
			a.href = r.u;
			a.id = 'site-search-hit-' + n;
			a.setAttribute('role', 'option');
			const t = document.createElement('span');
			t.className = 'site-search-hit-title';
			t.textContent = r.t;
			const w = document.createElement('span');
			w.className = 'site-search-hit-where lbl';
			w.textContent = r.s;
			a.append(t, w);
			p.appendChild(a);
		});
		p.hidden = false;
		i.setAttribute('aria-expanded', 'true');
	};

	const move = (step) => {
		const hits = panel() ? [...panel().querySelectorAll('.site-search-hit')] : [];
		if (!hits.length) return;
		active = (active + step + hits.length) % hits.length;
		hits.forEach((h, n) => h.toggleAttribute('data-active', n === active));
		hits[active].scrollIntoView({ block: 'nearest' });
		input().setAttribute('aria-activedescendant', hits[active].id);
	};

	document.addEventListener('focusin', (ev) => {
		if (ev.target.closest && ev.target.closest('[data-site-search-input]')) load();
	});

	document.addEventListener('input', (ev) => {
		const el = ev.target.closest && ev.target.closest('[data-site-search-input]');
		if (!el) return;
		const q = el.value.trim();
		load().then(() => render(search(q), q));
	});

	document.addEventListener('keydown', (ev) => {
		const el = input();
		if (!el) return;
		const inField = ev.target === el;
		if (ev.key === '/' && !inField && !/^(INPUT|TEXTAREA|SELECT)$/.test(ev.target.tagName) && !ev.target.isContentEditable && !ev.metaKey && !ev.ctrlKey) {
			ev.preventDefault();
			el.focus();
			return;
		}
		if (!inField) return;
		if (ev.key === 'ArrowDown') { ev.preventDefault(); move(1); }
		else if (ev.key === 'ArrowUp') { ev.preventDefault(); move(-1); }
		else if (ev.key === 'Escape') { close(); el.blur(); }
		else if (ev.key === 'Enter' && active >= 0) {
			const hit = panel().querySelectorAll('.site-search-hit')[active];
			if (hit) { ev.preventDefault(); hit.click(); }
		}
	});

	document.addEventListener('click', (ev) => {
		const b = box();
		if (b && !b.contains(ev.target)) close();
	});

	// A page swapped in by YPY replaces the header too, so a stale panel closes.
	document.addEventListener('ypy:after:swap', close);

	// The documentation menu is open beside the article and closed above it on a
	// narrow screen. The markup says open, so without this script it is a list.
	const wide = window.matchMedia('(min-width: 1024px)');
	const menus = () => document.querySelectorAll('.docs-menu').forEach((d) => { d.open = wide.matches; });
	menus();
	wide.addEventListener('change', menus);
	document.addEventListener('ypy:after:process', menus);
	document.addEventListener('DOMContentLoaded', menus);
})();
