/* checkout.js — оформление заказа */

function checkoutPage() {
  const root = document.querySelector('[data-checkout-root]');
  if (!root) return;

  const items = Cart.read();
  if (!items.length) {
    root.innerHTML = `<div class="rounded-2xl border border-line bg-paper px-6 py-16 text-center">
      <p class="text-xl font-bold">Корзина пуста</p>
      <a href="/catalog" class="mt-6 inline-flex h-12 items-center justify-center rounded-full bg-forest px-6 text-sm font-bold text-paper transition hover:bg-forest-dark">В каталог</a>
    </div>`;
    return;
  }

  root.innerHTML = `
    <div class="grid gap-8 lg:grid-cols-[1.6fr_1fr]">
      <form data-checkout-form class="grid gap-4 rounded-2xl border border-line bg-paper p-6">
        <label class="grid gap-2">
          <span class="text-xs font-bold uppercase tracking-[.12em] text-subtle">Имя *</span>
          <input name="name" required class="h-12 rounded-xl border border-line bg-canvas px-4 text-sm outline-none transition focus:border-forest">
        </label>
        <label class="grid gap-2">
          <span class="text-xs font-bold uppercase tracking-[.12em] text-subtle">Телефон *</span>
          <input name="phone" required class="h-12 rounded-xl border border-line bg-canvas px-4 text-sm outline-none transition focus:border-forest">
        </label>
        <label class="grid gap-2">
          <span class="text-xs font-bold uppercase tracking-[.12em] text-subtle">E-mail</span>
          <input name="email" type="email" class="h-12 rounded-xl border border-line bg-canvas px-4 text-sm outline-none transition focus:border-forest">
        </label>
        <label class="grid gap-2">
          <span class="text-xs font-bold uppercase tracking-[.12em] text-subtle">Адрес доставки</span>
          <input name="address" class="h-12 rounded-xl border border-line bg-canvas px-4 text-sm outline-none transition focus:border-forest">
        </label>
        <label class="grid gap-2">
          <span class="text-xs font-bold uppercase tracking-[.12em] text-subtle">Комментарий</span>
          <textarea name="comment" rows="3" class="rounded-xl border border-line bg-canvas p-4 text-sm outline-none transition focus:border-forest"></textarea>
        </label>
        <button type="submit" class="mt-2 h-12 rounded-full bg-forest px-6 text-sm font-bold text-paper transition hover:bg-forest-dark">Перейти к оплате</button>
        <p data-error class="hidden text-sm font-semibold text-forest"></p>
      </form>

      <aside class="h-max rounded-2xl border border-line bg-paper p-6">
        <div class="text-xs font-bold uppercase tracking-[.12em] text-subtle">Ваш заказ</div>
        <div class="mt-4 grid gap-3 text-sm">
          ${items.map(i => `<div class="flex justify-between gap-3">
            <span class="text-subtle">
              ${escapeHtml(i.name)} × ${i.qty}
              ${i.discountType === 'bundle' && i.bundleTotalQty ? `<span class="ml-1 font-bold text-forest">(${i.bundleBuyQty}+${i.bundleTotalQty - i.bundleBuyQty}=${i.bundleTotalQty})</span>` : ''}
            </span>
            <span class="font-bold">${fmtPrice(lineTotal(i))}</span>
          </div>`).join('')}
        </div>
        <div class="mt-5 flex justify-between border-t border-line pt-4">
          <span class="font-bold">Итого</span>
          <span class="text-2xl font-extrabold">${fmtPrice(Cart.total())}</span>
        </div>
      </aside>
    </div>`;

  const form = root.querySelector('[data-checkout-form]');
  const err = root.querySelector('[data-error]');

  form.addEventListener('submit', async (e) => {
    e.preventDefault();
    const fd = new FormData(form);
    const btn = form.querySelector('button[type=submit]');
    btn.disabled = true;
    btn.textContent = 'Создаём заказ…';

    try {
      const res = await fetch('/api/orders', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          name: fd.get('name'),
          phone: fd.get('phone'),
          email: fd.get('email'),
          address: fd.get('address'),
          comment: fd.get('comment'),
          items: Cart.read().map(i => ({ productId: i.id, qty: i.qty })),
        }),
      });
      if (!res.ok) throw new Error(await res.text());
      const data = await res.json();
      Cart.clear();
      location.href = '/payment?order=' + data.orderId;
    } catch (e2) {
      err.textContent = 'Не удалось создать заказ: ' + e2.message;
      err.classList.remove('hidden');
      btn.disabled = false;
      btn.textContent = 'Перейти к оплате';
    }
  });
}
