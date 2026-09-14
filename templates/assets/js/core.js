/* core.js — хелперы, иконки, API */

const API = {
  categories: () => fetch('/api/categories').then(r => r.json()),
  catalog: (cat) => fetch('/api/catalog' + (cat ? '?category=' + encodeURIComponent(cat) : '')).then(r => r.json()),
  product: (slug) => fetch('/api/products/' + encodeURIComponent(slug)).then(r => r.json()),
};

const fmtPrice = (n) => new Intl.NumberFormat('ru-RU').format(n) + ' ₽';

const escapeHtml = (s) => String(s ?? '').replace(/[&<>"']/g, (c) => ({
  '&': '&amp;', '<': '&lt;', '>': '&gt;', '"': '&quot;', "'": '&#39;',
}[c]));

/* ---- скидки: проценты и «комплекты» (напр. 2+1=3) ---- */

// сколько единиц товара нужно оплатить при покупке qty штук по схеме
// «купи buy — получи total»; должно совпадать с bundlePayableQty в api/api.go
function bundlePayableQty(qty, buy, total) {
  if (!buy || !total || buy <= 0 || total <= buy) return qty;
  const groups = Math.floor(qty / total);
  const rem = qty % total;
  return groups * buy + rem;
}

// краткая подпись для бейджа на карточке/странице товара
function discountBadge(p) {
  if (p.discountType === 'percent' && p.discountPercent > 0) {
    return `−${p.discountPercent}%`;
  }
  if (p.discountType === 'bundle' && p.bundleBuyQty && p.bundleTotalQty > p.bundleBuyQty) {
    const free = p.bundleTotalQty - p.bundleBuyQty;
    return `${p.bundleBuyQty}+${free}=${p.bundleTotalQty}`;
  }
  return '';
}

// поясняющая фраза для страницы товара
function discountHint(p) {
  if (p.discountType === 'bundle' && p.bundleBuyQty && p.bundleTotalQty > p.bundleBuyQty) {
    return `Акция: купите ${p.bundleTotalQty} шт. — заплатите как за ${p.bundleBuyQty}`;
  }
  return '';
}

// стоимость строки корзины/заказа с учётом типа скидки товара
function lineTotal(item) {
  const qty = item.qty || 0;
  if (item.discountType === 'bundle') {
    return item.price * bundlePayableQty(qty, item.bundleBuyQty, item.bundleTotalQty);
  }
  return item.price * qty;
}

const icon = (name) => {
  const paths = {
    menu:   '<path d="M4 7h16M4 12h16M4 17h16"/>',
    close:  '<path d="m6 6 12 12M18 6 6 18"/>',
    search: '<circle cx="11" cy="11" r="7"/><path d="m20 20-4-4"/>',
    bag:    '<path d="M6 8h12l1 12H5L6 8Z"/><path d="M9 9V6a3 3 0 0 1 6 0v3"/>',
    arrow:  '<path d="M5 12h14M14 7l5 5-5 5"/>',
    left:   '<path d="m15 6-6 6 6 6"/>',
    right:  '<path d="m9 6 6 6-6 6"/>',
  };
  return `<svg aria-hidden="true" width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.7" stroke-linecap="round" stroke-linejoin="round">${paths[name]}</svg>`;
};