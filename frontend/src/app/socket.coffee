define ['knockout', 'jquery'], (ko, $) ->
    Socket = (app) ->
        @app = app
        @app.otherUsersSessions = ko.observableArray()
        @eventHandlers = ko.observableArray()

        @setupTicketsWithOthersUsersViewing()
        return

    Socket::Start = () ->
        # Sanity check
        console.assert @app? && @app.currentRoute?

        # Prevent starting twice
        return if @socketIsConnected? && @socketIsConnected()

        if twDeskConfig.isDeveloperMode then addr = "http://#{location.hostname}:8840/"
        else addr = "https://desksockets2.lifevault.com:8840/"

        try
            @socketIsConnected = ko.observable false
            @socket = io.connect addr,
                "connect timeout": 1000
                reconnect: false #have to leave this off and manually reconnect at random intervals or server swamped
                "reconnection delay": 500
                "reopen delay": 2000
                "max reconnection attempts": 100
                "force new connection": true
                transports: ["websocket"]

            @socket.on "error", (reason) =>
                @socketIsConnected false
                # TODO reconnect with limit
                return

            @socket.on "connect", (e,i,o) =>
                return unless app.loggedInUser?
                @socket.emit "AUTH", {
                    secret: app.loggedInUser.authToken.hash()
                    installation_id: app.loggedInUser.authToken.installation_id()
                    user_id: app.loggedInUser.authToken.user_id()
                    timestamp: app.loggedInUser.authToken.timestamp()
                }
                return

            @socket.on "authResult", (e) =>
                if $.isArray(e) then e = e[0]
                if e.authenticated
                    @socketIsConnected true
                    @NotifyOpenedPage()
                else
                    console.log("NOT AUTHORIZED")

            @socket.on "updateNotice", (e) =>
                if $.isArray(e) then e = e[0]

                for handler in @eventHandlers()
                    if handler.NotifyEvent?
                        handler.NotifyEvent {
                            itemType: e.updateName
                            actionType: 'update'
                            eventData: e
                        }

                switch e.updateName
                    when "updateUsers"
                        @app.UpdateUsers()
                    when "updateSettings"
                        @app.UpdateSettings()
                    # This event should come in whenever the counts need to be updated
                    when "updateInboxes"
                        if e.userId? and e.userId is app.loggedInUser.id() then return
                        @app.UpdateDates()
                    when "userLogout"
                        if @app.loggedInUser? and @app.loggedInUser.id() is e.userId
                            @app.flash.Info 'Your session was logged out in another application!'
                            @app.Logout()

            @socket.on "eventNotice", (e) =>
                if $.isArray(e) then e = e[0]

                console.assert( e.eventName isnt null, "eventName required" )
                console.assert( e.socketId isnt null, "socketId must be passed" )

                # Verify that we know who this user is = or ignore
                if e.userId? && @app.FindUserById( e.userId ) is null then return

                # Handle the event
                switch e.eventName
                    when "openedPage"# Means we received a notice that another user has opened a page
                        console.assert( not isNaN( e.userId ) )
                        console.assert( e.page isnt null, "page must be passed" )
                        console.assert( e.userPageStatus isnt null, "userPageStatus must be passed" )
                        console.assert( not isNaN(e.userId), "userId must be passed" )

                        # We don't care about notifications from outselves
                        if e.userId is @app.loggedInUser.id() then return
                        # Get the user or quit if we can't find her
                        e.user = @app.FindUserById( e.userId )
                        if e.user is null then return

                        # Exit here if we are not viewing the same ticket (we don't care)
                        @UpdateOtherUserSessions( e )

                        # Let the other person know that we are also viewing this ticket - echo back
                        # or if it's the other users first load, everybody lets him know where they are
                        onSamePage = (e.page is @app.currentRoute().request_)
                        if onSamePage or e.isFirstLoad
                            r = @app.currentRoute()
                            eventNotice =
                                eventName: "viewingPage"
                                page: r.request_
                                userId: @app.loggedInUser.id()
                                userPageStatus: r.userPageStatus
                            @socket.emit "eventNotice", eventNotice

                    # When we open a room, 'viewingPage' event notices should come in from all
                    # other clients who are in the also viewing this page
                    when "viewingPage"
                        console.assert( not isNaN( e.userId ) )
                        console.assert( e.page isnt null, "page must be passed" )
                        console.assert( not isNaN(e.userId), "userId must be passed" )

                        # Get the user or quit if we can't find her
                        e.user = @app.FindUserById( e.userId )
                        if e.user is null then return

                        # Exit here if we are not viewing the same ticket (we don't care)
                        @UpdateOtherUserSessions( e )

                    # When a user disconnects just mark them as no longer viewing this page
                    when "disconnected"
                        e.userId = 0#required for UpdateOtherUserSessions
                        @UpdateOtherUserSessions( e, false )
        catch err
            console.log("Error setting up socket")


    Socket::RegisterEventHandler = (obj) ->
        @eventHandlers.push obj
        return

    Socket::UnregisterEventHandler = (obj) ->
        idx = @eventHandlers().indexOf obj

        if idx isnt -1 then @eventHandlers.splice idx, 1
        return

    # Sets up ticketsWithOthersUsersViewing
    # which remembers which tickets have users viewing them.
    # This is then used by ticket for the user presence indicator
    Socket::setupTicketsWithOthersUsersViewing = () ->

        # If this is already setup, then quietly exit
        if @app.ticketsWithOthersUsersViewing? then return

        # Watch otherUsersSessions() for chances and recalculate this as needed
        @app.ticketsWithOthersUsersViewing = ko.computed =>
            # We need this fn to know to watch @app.otherUsersSessions() so access it here
            otherUsersSessions = @app.otherUsersSessions()
             # Exit early if not logged in
            if not @app.loggedInUser? or @app.loggedInUser.id() is 0 then return {}

            loggedInUserId = @app.loggedInUser.id()
            if loggedInUserId <= 0 then return {}

            # Loop over other user sessions, find ones on a ticket and remember them in new result set
            ticketIds = {}
            for loopUserSession, i in otherUsersSessions
                if loopUserSession.userId is loggedInUserId then continue # Ignore self
                pageParts =  loopUserSession.page.split("/")

                if pageParts[0] isnt "tickets" then continue
                ticketId = parseInt( pageParts[1], 10 )

                if ticketIds[ticketId] is undefined then ticketIds[ticketId] = []
                ticketIds[ticketId].push(loopUserSession)

            return ticketIds

        return


    Socket::NotifyOpenedPage = () ->
        # Sanity check
        if not @app.loggedInUser? or typeof @app.loggedInUser.id isnt "function" or @app.loggedInUser.id() <= 0 then return
        # Build the event object
        page = @app.currentRoute().request_
        eventNotice =
            eventName: "openedPage"
            page: page
            userId: @app.loggedInUser.id()
            userPageStatus: @app.currentRoute().userPageStatus

        # Is this the first load (openedPageCount is 0), then pass that flag
        # This flag makes every other connected client echo their current status
        if not @openedPageCount? then @openedPageCount = 0
        else @openedPageCount += 1
        if @openedPageCount is 0 then eventNotice.isFirstLoad = true

        @socket.emit "eventNotice", eventNotice

    Socket::NotifyNewRoute = () ->
        # Now that we are on a new page, we want to track who else is viewing this page
        # We do so with these 2 variables attached to currentRoute
        r = @app.currentRoute()
        r.viewingUsers = @app.ko.observableArray()# Used to track what others are viewing
        r.viewingUserSessions = @app.ko.observableArray()# Used to track what others are viewing

        if not @socketIsConnected? or not @socketIsConnected() then return
        @NotifyOpenedPage()

        return


    Socket::UpdateOtherUserSessions = (e) ->
        console.assert( e.userId > 0 or e.eventName is "disconnected" )# user 0 only allowed for disconnect

        onSamePage = @app.currentRoute().request_ is e.page

        r = @app.currentRoute()
        if not r.viewingUsers? then return

        userSession =
            userId: e.userId
            user:  app.FindUserById(e.userId)
            socketId: e.socketId
            page: e.page
            userPageStatus: e.userPageStatus
            eventName: e.eventName

        # Find if we have this user (by socketId) on the ticket already
        otherUsersSessions = @app.otherUsersSessions()
        pos = -1
        for loopUserSession, i in otherUsersSessions
            if loopUserSession.socketId is userSession.socketId
                pos = i
                break

        # Remove the user session if this is a disconnect
        # Otherwise update or add the user session
        if e.eventName is "disconnected"
            if pos > -1 then otherUsersSessions.splice(pos,1)
        else
            if app.loggedInUser.id() isnt userSession.userId
                if pos is -1 then otherUsersSessions.push( userSession )
                else otherUsersSessions[ pos ] =  userSession

        # We need to remember all user sessions that are viewing the same page
        # We also need a unique array of userIds for display
        # (A single user could be connected multiple times on different devices)
        # We need to also excluse sessions with the same userId as current user.
        currentUserId = @app.loggedInUser.id()
        myPage = r.request_
        uniqueUsersIds = []
        viewingUserSessions = []
        ko.utils.arrayMap otherUsersSessions, (userSession) ->
            if userSession.page is myPage and userSession.userId isnt currentUserId
                viewingUserSessions.push(userSession)
                if uniqueUsersIds.indexOf(userSession.userId) is -1 then uniqueUsersIds.push(userSession.userId)
        r.viewingUsers(uniqueUsersIds)
        r.viewingUserSessions(viewingUserSessions)

        @app.otherUsersSessions(otherUsersSessions)
        return





    # NotifyPageStatus sends the usual "openedPage" event but
    # includes the provided status
    Socket::SetUserPageStatus = ( userPageStatus ) ->
        r = @app.currentRoute()
        if r.userPageStatus is userPageStatus then return# Ignore same status
        r.userPageStatus = userPageStatus

        # Notify others that this page status has changed
        @NotifyOpenedPage()
        return



    Socket::Stop = () ->
        if @socket then @socket.disconnect()
        if @app.socket.socketIsConnected? then @app.socket.socketIsConnected(false)
        return

    return Socket