/* contacts.js — форма обратной связи */

function contactsPage() {
  const form = document.querySelector('[data-contact-form]');
  if (!form) return;

  const state = form.querySelector('[data-form-state]');

  form.addEventListener('submit', async (e) => {
    e.preventDefault();
    const fd = new FormData(form);
    const btn = form.querySelector('button[type=submit]');
    btn.disabled = true;

    try {
      const res = await fetch('/api/contact', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          name: fd.get('name'),
          phone: fd.get('phone'),
          message: fd.get('message'),
        }),
      });
      if (!res.ok) throw new Error(await res.text());
      form.reset();
      state.textContent = 'Спасибо! Мы свяжемся с вами в ближайшее время.';
    } catch (err) {
      state.textContent = 'Не удалось отправить: ' + err.message;
    } finally {
      btn.disabled = false;
    }
  });
}
