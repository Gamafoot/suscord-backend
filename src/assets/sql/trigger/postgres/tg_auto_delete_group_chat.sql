DROP TRIGGER IF EXISTS tg_auto_delete_group_chat ON chat_members;
DROP FUNCTION IF EXISTS auto_delete_group_chat();

CREATE FUNCTION auto_delete_group_chat()
RETURNS TRIGGER
LANGUAGE plpgsql
AS $$
BEGIN
    IF NOT EXISTS (
        SELECT 1
        FROM chat_members
        WHERE chat_id = OLD.chat_id
    ) THEN
        DELETE FROM chats WHERE id = OLD.chat_id;
    END IF;

    RETURN OLD;
END;
$$;

CREATE TRIGGER tg_auto_delete_group_chat
AFTER DELETE ON chat_members
FOR EACH ROW
EXECUTE FUNCTION auto_delete_group_chat();
