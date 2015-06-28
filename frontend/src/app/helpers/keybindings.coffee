define [], ->

    # If we keep all of the keybinding definitions
    # in this one file it will help a lot with trying to maintain
    # and document.
    KeyBindings = ->
        @focusSearch = ['q']
        @reply = 'r'
        @note = 'n'
        @delete = ['command+backspace', 'command+del', 'ctrl+backspace', 'ctrl+del', 'd']
        @previousTicket = ['j', 'ctrl+,', 'command+,']
        @nextTicket = ['k', 'ctrl+.', 'command+.']
        @submitTicket = ['ctrl+enter', 'command+enter']
        @focusCC = ['ctrl+shift+c', 'command+shift+c']
        @focusBCC = ['ctrl+shift+b', 'command+shift+b']
        @focusBody = ['shift+esc']
        @attachments = ['command+shift+a', 'ctrl+shift+a']
        @assign = ['l u', 'a']
        @spam = ['s', 'l s']
        @pending = ['p', 'l p']
        @solved = ['v', 'l v']
        @closed = ['c', 'l c']
        @waiting = ['w', 'l w']

        @closeFilters = ['esc']

        @options = ['o', 'l o'] # open optons drop menu
        @addTask = ['t', 'l t']

        @newTicket = ['shift+c'] # create a new ticket in the current inbox

        @publishArticle = ['h p']
        @unpublishArticle = ['h u']
        @draft = ['h d']

        @markActive = ['l a']

        @selectAll = ['command+a', 'ctrl+a']

        #ticketPage
        @ticketsList =
            cursorDown: ['j','down']
            cursorUp: ['k','up']
            cursorSelect: ['space']
            cursorView: ['enter']
        return

    return KeyBindings