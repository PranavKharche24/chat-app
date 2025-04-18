-- lua/widgets.lua
-- Called at startup to build UI
function build_login_screen()
    clear_chat()
    set_title("Login to Super‑Chat")
    -- call Go handler to show modal-driven form, or emit events
  end
  
  function build_register_screen()
    clear_chat()
    set_title("Register New Account")
    -- ...
  end
  
  function build_chat_screen()
    clear_chat()
    set_title("Super‑Chat Room")
  end
  