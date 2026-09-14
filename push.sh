#!/usr/bin/env bash
set -euo pipefail

# ==== НАСТРОЙКИ ====
DIR="$HOME/Desktop/Dragon"
URL="https://github.com/nnnpppxxx85-lang/shop.git"
BRANCH="main"
GIT_NAME="nnnpppxxx85-lang"
GIT_EMAIL="nnnpppxxx85-lang@users.noreply.github.com"
# ===================

command -v git >/dev/null || { echo "❌ git не установлен"; exit 1; }
[ -d "$DIR" ] || { echo "❌ Папка '$DIR' не найдена"; exit 1; }

cd "$DIR"
echo "📁 Папка: $DIR"

git config user.name  "$GIT_NAME"
git config user.email "$GIT_EMAIL"

# --- init если нужно ---
[ -d .git ] || { git init; echo "✅ git init"; }

# --- .gitignore если нет ---
if [ ! -f .gitignore ]; then
  cat > .gitignore <<'EOF'
# артефакты
*.exe
*.test
*.out
bin/
dist/
build/

# зависимости
node_modules/
vendor/

# окружение
.env
.env.local

# IDE
.idea/
.vscode/
*.swp

# ОС
.DS_Store
Thumbs.db

# логи
*.log
EOF
  echo "✅ создан .gitignore"
fi

# --- remote ---
if git remote get-url origin >/dev/null 2>&1; then
  git remote set-url origin "$URL"
else
  git remote add origin "$URL"
fi
echo "✅ remote: $URL"

# --- ветка ---
git branch -M "$BRANCH"

# --- commit ---
git add -A
if git diff --cached --quiet; then
  echo "ℹ️  нет изменений для коммита"
else
  git commit -m "update $(date '+%Y-%m-%d %H:%M')"
fi

# --- push ---
git push -u origin "$BRANCH"

echo "🎉 Готово: https://github.com/nnnpppxxx85-lang/shop"