define ['knockout', 'knockout-mapping', 'text!./template.html', 'userModel'], (ko, mapping, templateMarkup, userModel) ->

    VM = (params) ->
        @siteName = ko.observable ""
        @hasLoginError = ko.observable false
        @isLoggingIn = ko.observable false
        @logo = ko.observable ""
        @rememberLogin = ko.observable false
        @email = ko.observable ""

        # If already logged in, then log the user out
        app.lscache.set "loggedInUser", null
        app.essentialDataIsLoaded(false)# This is important - so the essential data will be reloaded with the next login
        if app.loggedInUser?
            $.getJSON( 'v1/logout.json' )
            app.ClearLoggedInUser()

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

        if username == ''
            $("#userLogin").focus()
            app.FlashError("We need your login.", {timer:800} )
            return false
        if password == ''
            $("#password").focus()
            app.FlashError("We need your password.", {timer:800} )
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
                app.loggedInUser = app.mapping.fromJS(response).user
                app.loggedInUser.isAdmin = ko.observable(app.loggedInUser.role() is "Admin")
                app.SaveLoggedInUser() #saves in lscache
                app.lscache.set "lastGoodUsername", username

                if app.requestedRouteBeforeLoginRedirect?
                    app.GoTo app.requestedRouteBeforeLoginRedirect.request_
                    app.requestedRouteBeforeLoginRedirect = undefined
                else
                    app.GoTo "dashboard"
                    app.InitSocket()
            ,
            error: (xhr) =>
                @isLoggingIn false
                @hasLoginError true
                $("#password").focus()
                if xhr.status == 401
                    if xhr.responseJSON? && xhr.responseJSON.errors?
                        app.FlashError("<strong>" + xhr.responseJSON.errors[0] + "</strong> &mdash; Please try again", {timer:3000} )
                    else
                        app.FlashError("<strong>Error logging in</strong> &mdash; Please try again", {timer:3000} )
                else
                    app.HandleAjaxError(xhr)
        })

        return








    # This runs when the component is torn down. Put here any logic necessary to clean up,
    # for example cancelling setTimeouts or disposing Knockout subscriptions/computeds.
    VM::dispose = ->

    return {
        viewModel: VM
        template: templateMarkup
    }