DROP TRIGGER IF EXISTS tg_auto_delete_group_chat;

CREATE TRIGGER tg_auto_delete_group_chat
AFTER DELETE ON chat_members
FOR EACH ROW
WHEN NOT EXISTS (
    SELECT 1
    FROM chat_members
    WHERE chat_id = OLD.chat_id
)
BEGIN
    DELETE FROM chats WHERE id = OLD.chat_id;
END;
