/* payment.js — оплата переводом на карту/телефон + подтверждение квитанцией */

async function paymentPage() {
  const root = document.querySelector('[data-payment-root]');
  if (!root) return;

  const orderId = new URLSearchParams(location.search).get('order');
  if (!orderId) {
    root.innerHTML = `<div class="rounded-2xl border border-line bg-paper px-6 py-16 text-center"><p class="text-xl font-bold">Заказ не найден</p></div>`;
    return;
  }

  let order, settings;
  try {
    const r = await fetch('/api/orders/' + orderId);
    if (!r.ok) throw new Error();
    order = await r.json();
    settings = await API.settings().catch(() => ({}));
  } catch {
    root.innerHTML = `<div class="rounded-2xl border border-line bg-paper px-6 py-16 text-center"><p class="text-xl font-bold">Заказ не найден</p></div>`;
    return;
  }

  root.innerHTML = `
    <div class="grid gap-8 lg:grid-cols-[1.6fr_1fr]">
      <div class="grid gap-4">
        <div class="rounded-2xl border border-line bg-paper p-6">
          <h2 class="text-lg font-bold">Реквизиты для оплаты</h2>
          <div class="mt-5 grid gap-4 sm:grid-cols-2">
            <div class="rounded-xl border border-line bg-canvas p-4">
              <div class="text-xs font-bold uppercase tracking-[.12em] text-subtle">Номер карты</div>
              <div class="mt-2 text-lg font-extrabold tabular-nums">${escapeHtml(settings.paymentCard || '—')}</div>
            </div>
            <div class="rounded-xl border border-line bg-canvas p-4">
              <div class="text-xs font-bold uppercase tracking-[.12em] text-subtle">Или по номеру телефона (СБП)</div>
              <div class="mt-2 text-lg font-extrabold tabular-nums">${escapeHtml(settings.paymentPhone || '—')}</div>
            </div>
          </div>
          <p class="mt-4 text-sm text-subtle">Переведите ${fmtPrice(order.total)} и прикрепите квитанцию (PDF) — после этого заказ уйдёт на проверку менеджеру.</p>
        </div>

        <form data-confirm-form class="rounded-2xl border border-line bg-paper p-6">
          <h2 class="text-lg font-bold">Подтверждение оплаты</h2>
          <label class="mt-4 grid gap-2">
            <span class="text-xs font-bold uppercase tracking-[.12em] text-subtle">Квитанция об оплате (PDF)</span>
            <input type="file" name="receipt" accept="application/pdf" required class="rounded-xl border border-line bg-canvas px-4 py-3 text-sm outline-none file:mr-4 file:h-9 file:rounded-full file:border-0 file:bg-forest file:px-4 file:text-sm file:font-bold file:text-paper">
          </label>
          <p data-confirm-error class="mt-3 hidden text-sm font-semibold text-forest"></p>
          <button type="submit" class="mt-5 h-12 w-full rounded-full bg-forest px-6 text-sm font-bold text-paper transition hover:bg-forest-dark">Подтвердить оплату</button>
        </form>

        <div data-done class="hidden rounded-2xl border border-line bg-paper p-6 text-center">
          <p class="text-xl font-bold">Спасибо за покупку!</p>
          <p class="mt-2 text-sm text-subtle">Менеджер отправит вам в СМС трек номер вашей посылки</p>
        </div>
      </div>

      <aside class="h-max rounded-2xl border border-line bg-paper p-6">
        <div class="text-xs font-bold uppercase tracking-[.12em] text-subtle">Заказ №${order.id}</div>
        <div class="mt-4 grid gap-3 text-sm">
          ${(order.items || []).map(i => `<div class="flex justify-between gap-3">
            <span class="text-subtle">${escapeHtml(i.name)} × ${i.qty}</span>
            <span class="font-bold">${fmtPrice(i.lineTotal ?? (i.price * i.qty))}</span>
          </div>`).join('')}
        </div>
        <div class="mt-5 flex justify-between border-t border-line pt-4">
          <span class="font-bold">К оплате</span>
          <span class="text-2xl font-extrabold">${fmtPrice(order.total)}</span>
        </div>
      </aside>
    </div>`;

  const form = root.querySelector('[data-confirm-form]');
  const err = root.querySelector('[data-confirm-error]');

  form.addEventListener('submit', async (e) => {
    e.preventDefault();
    err.classList.add('hidden');
    const btn = form.querySelector('button[type=submit]');
    btn.disabled = true;
    btn.textContent = 'Отправляем…';

    try {
      const fd = new FormData(form);
      const res = await fetch(`/api/orders/${orderId}/confirm-payment`, { method: 'POST', body: fd });
      if (!res.ok) throw new Error(await res.text());

      form.classList.add('hidden');
      root.querySelector('[data-done]').classList.remove('hidden');
    } catch (e2) {
      err.textContent = 'Не удалось отправить квитанцию: ' + e2.message;
      err.classList.remove('hidden');
      btn.disabled = false;
      btn.textContent = 'Подтвердить оплату';
    }
  });
}
