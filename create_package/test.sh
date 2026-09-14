cd /code/platform_all_use/create_package

# 1. 查看 backend.tar.gz 完整文件列表（前 30 行）
echo "===== backend.tar.gz 前 30 个条目 ====="
tar -tzf bin/bin/backend.tar.gz | head -30

# 2. 精确搜索 alembic 相关文件
echo
echo "===== 搜索 alembic 相关 ====="
tar -tzf bin/bin/backend.tar.gz | grep -i alembic || echo "❌ 完全未找到 alembic 相关文件"

# 3. 检查哪些关键文件缺失
echo
echo "===== 关键文件存在性 ====="
for pattern in '^(\./)?\.venv/bin/uvicorn$' '^(\./)?\.venv/bin/alembic$' '^(\./)?src/main\.py$' '^(\./)?alembic\.ini$' '^(\./)?migrations/'; do
  if tar -tzf bin/bin/backend.tar.gz | grep -qE "$pattern"; then
    echo "✅ 匹配: $pattern"
  else
    echo "❌ 缺失: $pattern"
  fi
done
