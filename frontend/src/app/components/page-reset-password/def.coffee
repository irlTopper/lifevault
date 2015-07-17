define ['knockout', 'knockout-mapping', 'text!./template.html', 'userModel'], (ko, mapping, templateMarkup, userModel) ->

    VM = (params) ->
        @siteName = ko.observable ""
        @logo = ko.observable ""
        @isSubmitting = ko.observable false
        @showTokenField = ko.observable true

        @token = ko.observable ""
        @password = ko.observable ""
        @confirmPassword = ko.observable ""

        @userId = ko.observable app.currentRoute().userId
        @name = ko.observable app.currentRoute().name

        $.getJSON "v1/settings/branding.json", (response,d,xhr) =>
            @logo response.branding.logo

        .fail (xhr) ->
            return app.showErrorLoadingMsg(xhr)

        # Fix for background not showing
        $('body').attr('id', 'page-login')

        if app.currentRoute().token?
            @token app.currentRoute().token
            @showTokenField false

        # Focus the password box
        $('#password').focus()

        app.InitOnShow(@)# Sets-up the OnShow() function
        return

    VM::onSubmitResetPassword = () ->
        if @password() isnt @confirmPassword()
            app.flash.Error('Woops, both passwords must match!')

            $('#confirmPassword').focus()
            return

        @isSubmitting true

        $.ajax({
            url: "v1/users/" + @userId() + "/resetpassword.json",
            type: 'PUT',
            dataType: 'json',
            data: {
                token: @token()
                password: @password()
            }
            success: (response,d,xhr) =>
                app.flash.Success('Your password was changed successfully!', { timer: 3000 })

                app.handleUserResponse response
                app.SaveLoggedInUser() #saves in lscache
                app.lscache.set "lastGoodUsername", app.loggedInUser.email()

                if app.requestedRouteBeforeLoginRedirect?
                    app.GoTo app.requestedRouteBeforeLoginRedirect.request_
                    app.requestedRouteBeforeLoginRedirect = undefined
                else
                    app.GoTo "journal"
            ,
            error: (xhr) =>
                @isSubmitting false

                if xhr.status is 404
                    app.GoTo 'forgotpassword'

                    app.flash.Error('Oops, that token was not found or is expired, please generate another one.', { time: 5000 })
                else
                    app.error.Ajax xhr
        })

    # This runs when the component is torn down. Put here any logic necessary to clean up,
    # for example cancelling setTimeouts or disposing Knockout subscriptions/computeds.
    VM::dispose = ->

    return {
        viewModel: VM
        template: templateMarkup
    }