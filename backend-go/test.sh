cd /code/platform_all_use/backend-go

sqlite3 ./app.db <<'SQL'
-- 媒体库：path 改为 /list（避开与 moduleId='media' 相等）
UPDATE platform_menu
   SET path        = '/list',
       update_time = CURRENT_TIMESTAMP
 WHERE name = 'MediaList' AND micro_app = 'media';

-- 父目录：path 从旧的 /media/stream 归一为 /media（仅语义修正）
UPDATE platform_menu
   SET path        = '/media',
       micro_app   = 'media',
       update_time = CURRENT_TIMESTAMP
 WHERE name = 'Stream' AND menu_type = 'M';

-- 核对
SELECT name, path, component, micro_app FROM platform_menu
 WHERE module_id='media' ORDER BY parent_id, order_num;
SQL
