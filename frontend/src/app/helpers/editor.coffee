define ["jquery"], ($) ->

    InsertSigVarOpts = [
        ["Mailbox - Email", "{%mailbox.email%}"]
        ["Mailbox - Name", "{%mailbox.name%}"]
        ["User - First Name", "{%user.firstName%}"]
        ["User - Last Name", "{%user.lastName%}"]
        ["User - Full Name", "{%user.fullName%}"]
        ["User - Email", "{%user.email%}"]
        ["User - Phone", "{%user.phone%}"]
        ["User - Job Title", "{%user.jobTitle,fallback=%}"]
    ]

    InsertReplyVarOpts = [
        ["Mailbox - Email", "{%mailbox.email%}"]
        ["Mailbox - Name", "{%mailbox.name%}"]
        ["Ticket - Id","{%ticket.id%}"]
        ["Customer - First Name","{%customer.firstName%}"]
        ["Customer - Last Name","{%customer.lastName%}"]
        ["Customer - Full Name","{%customer.fullName%}"]
        ["Customer - Email","{%customer.email%}"]
        ["Agent - First Name", "{%user.firstName%}"]
        ["Agent - Full Name", "{%user.fullName%}"]
        ["Agent - Email", "{%user.email%}"]
        ["Agent - Phone", "{%user.phone%}"]
        ["Agent - Job Title", "{%user.jobTitle,fallback=%}"]
    ]

    Editor = ->

        window.RedactorPlugins = {}  unless window.RedactorPlugins

        if not window.RedactorPlugins.copyCannedResponse?
            window.RedactorPlugins.copyCannedResponse = ->
                init: ->
                    @copyCannedResponseOpts = []

                    $.getJSON 'v1/inboxes/cannedresponses.json', (response) =>
                        for reply, i in response.replies
                            @copyCannedResponseOpts.push
                                title: app.utility.SafeHTML(reply.name)
                                func: (item) =>
                                    @insert.set response.replies[item].reply

                        if @copyCannedResponseOpts.length > 0
                            button = @button.add("insertCannedResponse", "Copy Canned Response")
                            @button.addDropdown button, @copyCannedResponseOpts

        # insertSigVar plugin
        if not window.RedactorPlugins.insertSigVar?
            window.RedactorPlugins.insertSigVar = ->
                init: ->
                    @insertSigVarOpts = InsertSigVarOpts
                    dropOpts = {}
                    for opt, i in @insertSigVarOpts
                        dropOpts[i] = { title:opt[0].replace('-','&mdash;'), func:@insertSigVar.insertSigVarCB }
                    button = @button.add("insertVar", "Insert Variable")
                    @button.addDropdown button, dropOpts
                    return
                insertSigVarCB: (iStr) ->
                    i = parseInt(iStr,10)
                    html = @insertSigVarOpts[i][1]
                    @insert.html html
                    return
        # insertReplyVar plugin
        if not window.RedactorPlugins.insertReplyVar?
            window.RedactorPlugins.insertReplyVar = ->
                init: ->
                    @insertReplyVarOpts = InsertReplyVarOpts
                    dropOpts = {}
                    for opt, i in @insertReplyVarOpts
                        dropOpts[i] = { title:opt[0].replace('-','&mdash;'), func:@insertReplyVar.insertReplyVarCB }
                    button = @button.add("insertVar", "Insert Variable")
                    @button.addDropdown button, dropOpts
                    return
                insertReplyVarCB: (iStr) ->
                    i = parseInt(iStr,10)
                    html = @insertReplyVarOpts[i][1]
                    @insert.html html
                    return
        # insertCannedResponse plugin
        if not window.RedactorPlugins.insertCannedResponse?
            window.RedactorPlugins.insertCannedResponse = ->
                init: ->
                    return if window.RedactorPlugins.insertCannedResponse.ticket is undefined
                    @ticket = window.RedactorPlugins.insertCannedResponse.ticket
                    @inbox = app.FindInboxById(@ticket.inboxId())
                    @user = app.loggedInUser
                    addButton = true

                    $.getJSON( 'v1/inboxes/' + @ticket.inboxId() + '/cannedresponses.json', (response,d,xhr) =>
                        @cannedResponses = response.replies
                        dropOpts = {}

                        # Save a new reply
                        if true || app.settings.desk.allowedManageCannedResponses()
                            dropOpts[0] = {
                                title: "--- Save as 'Canned Response' ---"
                                func: =>
                                    app.modal.Show "new-cannedresponse", {inboxId: @ticket.inboxId(), content: @code.get(), callback: (reply) =>
                                        app.flash.Success("This response has been saved to the inbox for future use")
                                    }, @
                            }


                        for opt, i in @cannedResponses
                            do (i) =>
                                dropOpts[i+1] =
                                    title: app.utility.SafeHTML(opt.name),
                                    func: =>
                                        @insertCannedResponse.insertCannedResponseCB(i)

                        unless true || app.settings.desk.allowedManageCannedResponses()
                            # The user is not an admin and cannot create canned responses.  If there are not
                            # any canned responses remove the button
                            addButton = @cannedResponses.length > 0

                        if addButton
                            button = @button.add("insertCannedResponse", "Insert/Create Canned Response")
                            @button.addDropdown button, dropOpts
                    )

                    return
                insertCannedResponseCB: (iStr) ->
                    i = parseInt(iStr,10)
                    vals =
                        "{%mailbox.email%}": @inbox.email()
                        "{%mailbox.name%}": @inbox.name()
                        "{%ticket.id%}": @ticket.id()
                        "{%customer.fullName%}": "#{@ticket.customer.firstName()} #{@ticket.customer.lastName()}"
                        "{%customer.firstName%}": @ticket.customer.firstName()
                        "{%customer.lastName%}": @ticket.customer.lastName()
                        "{%customer.email%}": @ticket.customer.email()
                        "{%user.fullName%}": "#{@user.firstName()} #{@user.lastName()}"
                        "{%user.firstName%}": @user.firstName()
                        "{%user.email%}": @user.email()
                        "{%user.phone%}": @user.phone()
                        "{%user.jobTitle,fallback=%}": @user.jobTitle()

                    html = @cannedResponses[i].reply
                    for k, v of vals
                        html = html.replace(new RegExp(k, 'g'), v)

                    @insert.html html
                    return

    return Editor