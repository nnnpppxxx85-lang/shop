/* admin.js — управление товарами, категориями и заказами */

const ADMIN_TOKEN_KEY = 'dm_admin_token';

function adminToken() {
  return localStorage.getItem(ADMIN_TOKEN_KEY) || '';
}

// обёртка над fetch: подставляет токен и возвращает null при 401,
// чтобы вызывающий код мог показать экран входа вместо падения с ошибкой
async function adminFetch(url, opts = {}) {
  const headers = Object.assign({}, opts.headers, { 'X-Admin-Token': adminToken() });
  const res = await fetch(url, Object.assign({}, opts, { headers }));
  if (res.status === 401) {
    localStorage.removeItem(ADMIN_TOKEN_KEY);
    showLoginOverlay();
    throw new Error('unauthorized');
  }
  return res;
}

function showLoginOverlay() {
  const overlay = document.querySelector('[data-admin-login]');
  if (overlay) { overlay.classList.remove('hidden'); overlay.classList.add('flex'); }
}
function hideLoginOverlay() {
  const overlay = document.querySelector('[data-admin-login]');
  if (overlay) { overlay.classList.add('hidden'); overlay.classList.remove('flex'); }
}

async function adminPage() {
  const root = document.querySelector('[data-admin-root]');
  if (!root) return;

  const status = root.querySelector('[data-admin-status]');
  function say(text) {
    status.textContent = text;
    status.classList.remove('hidden');
    setTimeout(() => status.classList.add('hidden'), 3000);
  }

  /* ---------- вход / выход ---------- */

  const loginForm = document.querySelector('[data-login-form]');
  const loginError = document.querySelector('[data-login-error]');
  const logoutBtn = root.querySelector('[data-logout]');

  loginForm?.addEventListener('submit', async (e) => {
    e.preventDefault();
    const password = loginForm.elements.password.value;
    const btn = loginForm.querySelector('button[type=submit]');
    btn.disabled = true;
    try {
      const res = await fetch('/api/admin/login', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ password }),
      });
      const data = await res.json().catch(() => ({}));
      if (!res.ok || !data.ok) throw new Error(data.error || 'Не удалось войти');
      localStorage.setItem(ADMIN_TOKEN_KEY, data.token);
      loginForm.reset();
      loginError.classList.add('hidden');
      hideLoginOverlay();
      logoutBtn.classList.remove('hidden');
      await bootstrap();
    } catch (err) {
      loginError.textContent = err.message;
      loginError.classList.remove('hidden');
    } finally {
      btn.disabled = false;
    }
  });

  logoutBtn?.addEventListener('click', () => {
    localStorage.removeItem(ADMIN_TOKEN_KEY);
    logoutBtn.classList.add('hidden');
    showLoginOverlay();
  });

  /* ---------- вкладки ---------- */

  const tabButtons = root.querySelectorAll('[data-tab]');
  const panels = root.querySelectorAll('[data-tab-panel]');
  const loadedTabs = new Set();

  tabButtons.forEach((btn) => {
    btn.addEventListener('click', () => {
      const tab = btn.dataset.tab;
      tabButtons.forEach((b) => {
        const active = b === btn;
        b.classList.toggle('border-forest', active);
        b.classList.toggle('bg-forest', active);
        b.classList.toggle('text-paper', active);
        b.classList.toggle('border-line', !active);
        b.classList.toggle('bg-paper', !active);
        b.classList.toggle('text-subtle', !active);
      });
      panels.forEach((p) => {
        const active = p.dataset.tabPanel === tab;
        // hidden и grid переключаем явно вместе: lg:grid в статичной разметке
        // всегда перебивает hidden на широких экранах (медиа-запрос идёт
        // позже в каскаде), поэтому видимость нельзя доверять одному классу
        p.classList.toggle('hidden', !active);
        p.classList.toggle('grid', active);
      });
      if (tab === 'categories' && !loadedTabs.has('categories')) { loadedTabs.add('categories'); loadCategories(); }
      if (tab === 'orders' && !loadedTabs.has('orders')) { loadedTabs.add('orders'); loadOrders(); }
      if (tab === 'settings' && !loadedTabs.has('settings')) { loadedTabs.add('settings'); loadSettings(); }
    });
  });

  /* ---------- товары ---------- */

  const listNode = root.querySelector('[data-admin-list]');
  const form = root.querySelector('[data-admin-form]');
  const formTitle = root.querySelector('[data-form-title]');
  const catSelect = form.querySelector('[name=categorySlug]');
  const search = root.querySelector('[data-product-search]');
  const folderList = document.getElementById('asset-folders');

  let products = [];
  let categories = [];

  async function loadCategoriesIntoSelect() {
    categories = await API.categories();
    catSelect.innerHTML = categories.map((c) => `<option value="${c.slug}">${escapeHtml(c.name)}</option>`).join('');
  }

  async function loadAssetFolders() {
    try {
      const res = await adminFetch('/api/admin/assets');
      const folders = await res.json();
      folderList.innerHTML = folders.map((f) => `<option value="${escapeHtml(f)}">`).join('');
    } catch { /* пользователь ещё не вошёл — список появится после входа */ }
  }

  function productMatches(p, q) {
    return `${p.name} ${p.slug} ${p.categoryName} ${p.variant || ''}`.toLowerCase().includes(q);
  }

  function renderList() {
    const q = (search.value || '').trim().toLowerCase();
    const shown = products.filter((p) => productMatches(p, q));
    listNode.innerHTML = shown.length ? shown.map(productRow).join('') :
      `<div class="rounded-2xl border border-line bg-paper px-6 py-12 text-center text-subtle">Товаров не найдено</div>`;
  }

  function productRow(p) {
    const badge = discountBadge(p);
    return `
      <div class="flex items-center gap-4 rounded-2xl border border-line bg-paper p-4">
        <div class="size-16 shrink-0 overflow-hidden rounded-xl bg-warm">
          ${p.images && p.images[0] ? `<img src="${escapeHtml(p.images[0])}" alt="" class="size-full object-cover">` : ''}
        </div>
        <div class="min-w-0 flex-1">
          <div class="truncate text-base font-bold">${escapeHtml(p.name)}</div>
          <div class="mt-1 flex flex-wrap items-center gap-2 text-xs text-subtle">
            <span>${escapeHtml(p.categoryName)} · ${escapeHtml(p.slug)} · ${fmtPrice(p.finalPrice)}</span>
            ${badge ? `<span class="rounded-full bg-forest/10 px-2 py-0.5 font-bold text-forest">${badge}</span>` : ''}
            ${p.isActive ? '' : '<span class="rounded-full bg-mist px-2 py-0.5 font-bold">скрыт</span>'}
          </div>
        </div>
        <button type="button" data-edit="${p.id}" class="h-10 rounded-full border border-line px-4 text-sm font-bold transition hover:border-forest/40">Изменить</button>
        <button type="button" data-del="${p.id}" class="h-10 rounded-full border border-line px-4 text-sm font-bold text-subtle transition hover:text-ink">Удалить</button>
      </div>`;
  }

  async function loadProducts() {
    try {
      const res = await adminFetch('/api/admin/products');
      products = await res.json();
      renderList();
    } catch { /* показан экран входа */ }
  }

  function updateDiscountUI() {
    const type = form.elements.discountType.value;
    form.querySelector('[data-discount-percent]').classList.toggle('hidden', type !== 'percent');
    form.querySelector('[data-discount-bundle]').classList.toggle('hidden', type !== 'bundle');
    updateBundlePreview();
  }

  function updateBundlePreview() {
    const buy = +form.elements.bundleBuyQty.value || 0;
    const total = +form.elements.bundleTotalQty.value || 0;
    const preview = form.querySelector('[data-bundle-preview]');
    preview.textContent = total > buy && buy > 0 ? `${buy}+${total - buy}=${total}` : '—';
  }

  form.querySelectorAll('[name=discountType]').forEach((r) => r.addEventListener('change', updateDiscountUI));
  form.elements.bundleBuyQty.addEventListener('input', updateBundlePreview);
  form.elements.bundleTotalQty.addEventListener('input', updateBundlePreview);

  function fill(p) {
    form.elements.id.value = p ? p.id : '';
    form.elements.categorySlug.value = p ? p.categorySlug : (categories[0]?.slug || '');
    form.elements.name.value = p ? p.name : '';
    form.elements.slug.value = p ? p.slug : '';
    form.elements.folder.value = p ? p.folder : '';
    form.elements.variant.value = p ? (p.variant || '') : '';
    form.elements.price.value = p ? p.price : '';
    form.elements.marketPrice.value = p && p.marketPrice ? p.marketPrice : '';
    form.elements.source.value = p ? (p.source || '') : '';
    form.elements.sortOrder.value = p ? p.sortOrder : 0;
    form.elements.isActive.checked = p ? !!p.isActive : true;
    form.elements.images.value = p ? (p.images || []).join('\n') : '';

    const type = p ? p.discountType || 'none' : 'none';
    form.querySelector(`[name=discountType][value="${type}"]`).checked = true;
    form.elements.discountPercent.value = p && p.discountPercent ? p.discountPercent : 10;
    form.elements.bundleBuyQty.value = p && p.bundleBuyQty ? p.bundleBuyQty : 2;
    form.elements.bundleTotalQty.value = p && p.bundleTotalQty ? p.bundleTotalQty : 3;
    updateDiscountUI();

    formTitle.textContent = p ? 'Редактирование товара #' + p.id : 'Новый товар';
    form.querySelector('[data-cancel-edit]').classList.toggle('hidden', !p);
    window.scrollTo({ top: form.offsetTop - 80, behavior: 'smooth' });
  }

  listNode.addEventListener('click', async (e) => {
    const edit = e.target.closest('[data-edit]');
    const del = e.target.closest('[data-del]');
    if (edit) fill(products.find((p) => p.id === +edit.dataset.edit));
    if (del) {
      if (!confirm('Удалить товар?')) return;
      try {
        await adminFetch('/api/admin/products/' + del.dataset.del, { method: 'DELETE' });
        say('Товар удалён');
        await loadProducts();
      } catch { /* показан экран входа */ }
    }
  });

  root.querySelector('[data-new]').addEventListener('click', () => fill(null));
  form.querySelector('[data-cancel-edit]').addEventListener('click', () => fill(null));

  form.querySelector('[data-scan-folder]').addEventListener('click', async () => {
    const folder = form.elements.folder.value.trim();
    if (!folder) { say('Сначала укажите папку с фото'); return; }
    if (form.elements.images.value.trim() && !confirm('Заменить список фото найденными в папке?')) return;
    try {
      const res = await adminFetch('/api/admin/assets/' + encodeURIComponent(folder));
      if (!res.ok) { say(await res.text()); return; }
      const images = await res.json();
      form.elements.images.value = images.join('\n');
      say(`Найдено фото: ${images.length}`);
    } catch { /* показан экран входа */ }
  });

  form.addEventListener('submit', async (e) => {
    e.preventDefault();
    const f = form.elements;
    const body = {
      categorySlug: f.categorySlug.value,
      name: f.name.value.trim(),
      slug: f.slug.value.trim(),
      folder: f.folder.value.trim(),
      variant: f.variant.value.trim(),
      price: +f.price.value || 0,
      marketPrice: +f.marketPrice.value || 0,
      discountType: f.discountType.value,
      discountPercent: +f.discountPercent.value || 0,
      bundleBuyQty: +f.bundleBuyQty.value || 0,
      bundleTotalQty: +f.bundleTotalQty.value || 0,
      source: f.source.value.trim(),
      sortOrder: +f.sortOrder.value || 0,
      isActive: f.isActive.checked,
      images: f.images.value.split('\n').map((s) => s.trim()).filter(Boolean),
    };
    const id = f.id.value;
    try {
      const res = await adminFetch('/api/admin/products' + (id ? '/' + id : ''), {
        method: id ? 'PUT' : 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(body),
      });
      if (!res.ok) { say('Ошибка: ' + await res.text()); return; }
      say(id ? 'Товар обновлён' : 'Товар добавлен');
      fill(null);
      await loadProducts();
    } catch { /* показан экран входа */ }
  });

  form.elements.name.addEventListener('blur', () => {
    if (!form.elements.slug.value) {
      form.elements.slug.value = form.elements.name.value.trim().toLowerCase()
        .replace(/[^a-z0-9а-я]+/gi, '-').replace(/^-|-$/g, '');
      // если фокус уже перешёл на это поле (например, по Tab), выделяем
      // автосгенерированный текст — иначе следующий же символ, набранный
      // пользователем, допишется в конец вместо замены значения
      if (document.activeElement === form.elements.slug) form.elements.slug.select();
    }
    if (!form.elements.folder.value) {
      form.elements.folder.value = form.elements.slug.value.replace(/-/g, '_');
    }
  });

  search.addEventListener('input', renderList);

  /* ---------- категории ---------- */

  const catList = root.querySelector('[data-category-list]');
  const catForm = root.querySelector('[data-category-form]');
  const catFormTitle = root.querySelector('[data-category-form-title]');

  function fillCategory(c) {
    catForm.elements.originalSlug.value = c ? c.slug : '';
    catForm.elements.name.value = c ? c.name : '';
    catForm.elements.slug.value = c ? c.slug : '';
    catForm.elements.code.value = c ? (c.code || '') : '';
    catForm.elements.note.value = c ? (c.note || '') : '';
    catForm.elements.sortOrder.value = c ? (c.sortOrder || 0) : 0;
    catFormTitle.textContent = c ? 'Редактирование категории' : 'Новая категория';
    catForm.querySelector('[data-cancel-category]').classList.toggle('hidden', !c);
  }

  async function loadCategories() {
    categories = await API.categories();
    catList.innerHTML = categories.length ? categories.map((c, i) => `
      <div class="flex items-center gap-4 rounded-2xl border border-line bg-paper p-4">
        <span class="w-8 shrink-0 text-xs font-bold text-forest/60">${escapeHtml(c.code || String(i + 1))}</span>
        <div class="min-w-0 flex-1">
          <div class="truncate text-base font-bold">${escapeHtml(c.name)}</div>
          <div class="mt-1 text-xs text-subtle">${escapeHtml(c.slug)}${c.note ? ' · ' + escapeHtml(c.note) : ''}</div>
        </div>
        <button type="button" data-edit-cat="${escapeHtml(c.slug)}" class="h-10 rounded-full border border-line px-4 text-sm font-bold transition hover:border-forest/40">Изменить</button>
        <button type="button" data-del-cat="${escapeHtml(c.slug)}" class="h-10 rounded-full border border-line px-4 text-sm font-bold text-subtle transition hover:text-ink">Удалить</button>
      </div>`).join('') :
      `<div class="rounded-2xl border border-line bg-paper px-6 py-12 text-center text-subtle">Категорий пока нет</div>`;
    await loadCategoriesIntoSelect();
  }

  catList.addEventListener('click', async (e) => {
    const edit = e.target.closest('[data-edit-cat]');
    const del = e.target.closest('[data-del-cat]');
    if (edit) fillCategory(categories.find((c) => c.slug === edit.dataset.editCat));
    if (del) {
      if (!confirm('Удалить категорию?')) return;
      try {
        const res = await adminFetch('/api/admin/categories/' + encodeURIComponent(del.dataset.delCat), { method: 'DELETE' });
        if (!res.ok) { say(await res.text()); return; }
        say('Категория удалена');
        await loadCategories();
      } catch { /* показан экран входа */ }
    }
  });

  root.querySelector('[data-cancel-category]').addEventListener('click', () => fillCategory(null));

  catForm.addEventListener('submit', async (e) => {
    e.preventDefault();
    const f = catForm.elements;
    const original = f.originalSlug.value;
    const body = {
      slug: f.slug.value.trim(),
      name: f.name.value.trim(),
      code: f.code.value.trim(),
      note: f.note.value.trim(),
      sortOrder: +f.sortOrder.value || 0,
    };
    try {
      const res = await adminFetch('/api/admin/categories' + (original ? '/' + encodeURIComponent(original) : ''), {
        method: original ? 'PUT' : 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(body),
      });
      if (!res.ok) { say('Ошибка: ' + await res.text()); return; }
      say(original ? 'Категория обновлена' : 'Категория добавлена');
      fillCategory(null);
      await loadCategories();
    } catch { /* показан экран входа */ }
  });

  /* ---------- заказы ---------- */

  const ordersList = root.querySelector('[data-orders-list]');
  const statusLabels = { new: 'Новый', awaiting_confirmation: 'Ждём проверки', paid: 'Оплачен', cancelled: 'Отменён' };

  async function loadOrders() {
    try {
      const res = await adminFetch('/api/admin/orders');
      const orders = await res.json();
      ordersList.innerHTML = orders.length ? orders.map(orderRow).join('') :
        `<div class="rounded-2xl border border-line bg-paper px-6 py-12 text-center text-subtle">Заказов пока нет</div>`;
    } catch { /* показан экран входа */ }
  }

  function orderRow(o) {
    const statusOptions = Object.keys(statusLabels)
      .map((key) => `<option value="${key}" ${key === o.status ? 'selected' : ''}>${statusLabels[key]}</option>`).join('');
    return `
      <details class="rounded-2xl border border-line bg-paper p-4">
        <summary class="flex cursor-pointer flex-wrap items-center justify-between gap-3">
          <span class="font-bold">Заказ №${o.id} · ${escapeHtml(o.name)}</span>
          <span class="flex items-center gap-3 text-sm text-subtle">
            <span>${escapeHtml(o.createdAt).replace('T', ' ').slice(0, 16)}</span>
            <span class="rounded-full bg-mist px-3 py-1 text-xs font-bold text-ink">${escapeHtml(statusLabels[o.status] || o.status)}</span>
            <span class="text-base font-extrabold text-ink">${fmtPrice(o.total)}</span>
          </span>
        </summary>
        <div class="mt-4 grid gap-2 border-t border-line pt-4 text-sm">
          <div class="text-subtle">Телефон: <span class="font-semibold text-ink">${escapeHtml(o.phone)}</span></div>
          ${o.address ? `<div class="text-subtle">Адрес: <span class="font-semibold text-ink">${escapeHtml(o.address)}</span></div>` : ''}
          ${o.comment ? `<div class="text-subtle">Комментарий: <span class="font-semibold text-ink">${escapeHtml(o.comment)}</span></div>` : ''}
          ${o.referralUsername ? `<div class="text-subtle">Привёл партнёр: <span class="font-semibold text-ink">@${escapeHtml(o.referralUsername)}</span></div>` : ''}
          <div class="mt-2 grid gap-1">
            ${(o.items || []).map((i) => `<div class="flex justify-between gap-3 text-subtle">
              <span>${escapeHtml(i.name)} × ${i.qty}</span><span class="font-semibold text-ink">${fmtPrice(i.lineTotal)}</span>
            </div>`).join('')}
          </div>
          <div class="mt-3 flex flex-wrap items-center gap-3 border-t border-line pt-3">
            <select data-order-status="${o.id}" class="h-10 rounded-full border border-line bg-paper px-4 text-sm font-semibold outline-none focus:border-forest">${statusOptions}</select>
            ${o.hasReceipt ? `<button type="button" data-order-receipt="${o.id}" class="h-10 rounded-full border border-line px-4 text-sm font-bold transition hover:border-forest/40">Квитанция</button>` : ''}
          </div>
        </div>
      </details>`;
  }

  ordersList.addEventListener('change', async (e) => {
    const select = e.target.closest('[data-order-status]');
    if (!select) return;
    try {
      const res = await adminFetch(`/api/admin/orders/${select.dataset.orderStatus}/status`, {
        method: 'PUT',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ status: select.value }),
      });
      if (!res.ok) { say('Ошибка: ' + await res.text()); return; }
      say('Статус обновлён');
      await loadOrders();
    } catch { /* показан экран входа */ }
  });

  ordersList.addEventListener('click', async (e) => {
    const btn = e.target.closest('[data-order-receipt]');
    if (!btn) return;
    // окно нужно открыть синхронно в обработчике клика — иначе браузер
    // считает open() уже не жестом пользователя и молча блокирует его
    // как всплывающее окно, когда adminFetch резолвится после await
    const tab = window.open('', '_blank');
    try {
      const res = await adminFetch(`/api/admin/orders/${btn.dataset.orderReceipt}/receipt`);
      if (!res.ok) { tab?.close(); say('Квитанция не найдена'); return; }
      const blob = await res.blob();
      if (tab) tab.location.href = URL.createObjectURL(blob);
    } catch {
      tab?.close();
      /* показан экран входа */
    }
  });

  /* ---------- настройки ---------- */

  const settingsForm = root.querySelector('[data-settings-form]');

  async function loadSettings() {
    try {
      const res = await fetch('/api/settings');
      const s = await res.json();
      settingsForm.elements.paymentCard.value = s.paymentCard || '';
      settingsForm.elements.paymentPhone.value = s.paymentPhone || '';
      settingsForm.elements.consultantTelegram.value = s.consultantTelegram || '';
    } catch { /* оставим поля пустыми */ }
  }

  settingsForm.addEventListener('submit', async (e) => {
    e.preventDefault();
    const f = settingsForm.elements;
    try {
      const res = await adminFetch('/api/admin/settings', {
        method: 'PUT',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          paymentCard: f.paymentCard.value.trim(),
          paymentPhone: f.paymentPhone.value.trim(),
          consultantTelegram: f.consultantTelegram.value.trim(),
        }),
      });
      if (!res.ok) { say('Ошибка: ' + await res.text()); return; }
      say('Настройки сохранены');
    } catch { /* показан экран входа */ }
  });

  /* ---------- запуск ---------- */

  async function bootstrap() {
    await loadCategoriesIntoSelect();
    await loadAssetFolders();
    await loadProducts();
    fill(null);
  }

  if (adminToken()) {
    logoutBtn.classList.remove('hidden');
    await bootstrap();
  } else {
    showLoginOverlay();
    // категории публичны — форма уже осмысленна к моменту входа
    await loadCategoriesIntoSelect();
    fill(null);
  }
}
