define ['knockout', 'knockout-mapping', 'text!./template.html', 'userModel'], (ko, mapping, templateMarkup, userModel) ->

    VM = (params) ->
        @hasPMInstalled = ko.observable false
        @siteName = ko.observable ""
        @hasLoginError = ko.observable false
        @isLoggingIn = ko.observable false
        @logo = ko.observable ""
        @rememberLogin = ko.observable false
        @email = ko.observable ""

        # If already logged in, then log the user out
        app.essentialDataIsLoaded(false)# This is important - so the essential data will be reloaded with the next login

        if $("#userLogin").val() == ''
            lastGoodUsername = app.lscache.get "lastGoodUsername"
            if lastGoodUsername?
                @email lastGoodUsername
                $("#password").focus()
            else $("#userLogin").focus()
        else
            $("#password").focus()

        app.InitOnShow(@)# Sets-up the OnShow() function
        return



    VM::onSubmitLogin = ->

        username = $("#userLogin").val()
        password = $("#password").val()

        if password == ''
            $("#password").focus()
            app.flash.Error("We need your password.", {timer:800} )
            return false

        @isLoggingIn true
        @hasLoginError false

        # Validate the login by loading the inboxes
        $.ajax({
            url: 'v1/login.json',
            type: 'POST',
            dataType: 'json',
            data: {
                username: username
                password: password
                rememberMe: @rememberLogin()
            }
            success: (response,d,xhr) =>
                app.handleUserResponse response
                app.SaveLoggedInUser() #saves in lscache
                app.lscache.set "lastGoodUsername", username

                if app.requestedRouteBeforeLoginRedirect?
                    app.GoTo app.requestedRouteBeforeLoginRedirect.request_
                    app.requestedRouteBeforeLoginRedirect = undefined
                else
                    app.GoTo "journal"

                # This must always be done
                app.socket.Start()
            ,
            error: (xhr) =>
                @isLoggingIn false
                @hasLoginError true
                $("#password").focus()
                if xhr.status == 401
                    if xhr.responseJSON? && xhr.responseJSON.errors?
                        app.flash.Error("<strong>" + xhr.responseJSON.errors[0] + "</strong> &mdash; Please try again", {timer:3000} )
                    else
                        app.flash.Error("<strong>Error logging in</strong> &mdash; Please try again", {timer:3000} )
                else
                    app.error.Ajax(xhr)
        })

        return








    # This runs when the component is torn down. Put here any logic necessary to clean up,
    # for example cancelling setTimeouts or disposing Knockout subscriptions/computeds.
    VM::dispose = ->

    return {
        viewModel: VM
        template: templateMarkup
    }